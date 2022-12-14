package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"mp4/api"
	"mp4/logger"
	"mp4/ralloc"
	"mp4/ring"
	"mp4/sdfs"
	"mp4/utils"
	"strings"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/table"
	"github.com/jedib0t/go-pretty/text"
	"github.com/montanaflynn/stats"
	"google.golang.org/grpc"
)

const PROCESS_QUEUE_INTERVAL = 1000 * time.Millisecond
const RESCHEDULE_INTERVAL = 2000 * time.Millisecond
const FLUSH_JOB_IONTERVAL = 2000 * time.Millisecond
const REFRESH_INTERVAL = 2000 * time.Millisecond
const BACKUP_INTERVAL = 3000 * time.Millisecond
const MEASURE_QPS_INTERVAL = 1000 * time.Millisecond

type WhichStatus int // command shortcuts client sends to coordinator

const (
	STATUS_WORKER              = "w"
	STATUS_JOBS                = "j"
	STATUS_JOB_ID              = "ij"
	STATUS_COMPLETED_JOBS      = "cj"
	STATUS_JOB_ID_STATS        = "ijs"
	STATUS_JSON_WORKER         = "jw"
	STATUS_JSON_JOBS           = "jj"
	STATUS_JSON_JOB_ID         = "jij"
	STATUS_JSON_COMPLETED_JOBS = "jcj"
)

type IDunnoCoordinator struct {
	TaskQueue       *utils.Queue[*api.InferenceTask]
	ModelStore      *ModelStore
	Ring            *ring.RingServer
	ResourceManager *ResourceManager
	Scheduler       *IDunnoScheduler
	SDFSClient      *sdfs.SDFSClient
	IsCoordinator   bool // flag to indicate if this coordinator is serving requests
	IsScheduling    bool // flag to indicate if this coordinator is scheduling jobs
	sync.Mutex
	api.CoordinatorServiceServer
}

func NewIDunnoCoordinator(ring *ring.RingServer, sdfsClient *sdfs.SDFSClient) *IDunnoCoordinator {
	rm := NewResourceManager()
	scheduler := NewIDunnoScheduler(rm)

	return &IDunnoCoordinator{
		TaskQueue:       utils.NewQueue[*api.InferenceTask](),
		ModelStore:      NewModelStore(),
		Ring:            ring,
		ResourceManager: rm,
		Scheduler:       scheduler,
		SDFSClient:      sdfsClient,
		IsCoordinator:   false,
		IsScheduling:    false,
	}
}

// Tasks that runs periodically
func (ic *IDunnoCoordinator) Corn() {
	go func() {
		for {
			time.Sleep(BACKUP_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.BackupCoordinatorData()
		}
	}()

	go func() {
		for {
			time.Sleep(PROCESS_QUEUE_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.ProcessQueuedJob()
		}
	}()

	go func() {
		for {
			time.Sleep(RESCHEDULE_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.RescheduleJobs()
		}
	}()

	go func() {
		for {
			time.Sleep(FLUSH_JOB_IONTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.FlushPendingJobs()
		}
	}()

	go func() {
		for {
			time.Sleep(REFRESH_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.RefreshBatchStatus()
		}
	}()

	go func() {
		for {
			time.Sleep(REFRESH_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.RefreshWorkerStatus()
		}
	}()

	go func() {
		for {
			time.Sleep(MEASURE_QPS_INTERVAL)
			if !ic.IsCoordinator {
				continue
			}
			ic.MesureStats()
		}
	}()
}

func (ic *IDunnoCoordinator) MesureStats() {
	ic.Lock()
	defer ic.Unlock()

	for _, job := range ic.Scheduler.ActiveJobs {
		job.MeasureQPS()
		job.MeasureQueryProcessTime()
	}
}

// Coordinator backup its data to its successor
func (ic *IDunnoCoordinator) BackupCoordinatorData() {
	ic.Lock()
	defer ic.Unlock()

	client, close, err := ic.CreateCoordinatorClient()
	if err != nil {
		return
	}
	defer close()

	modelStore := make(map[string]string)
	activeJobs := make([]*api.Job, 0)
	completedJobs := make([]*api.Job, 0)
	pendingJobs := make([]*api.Job, 0)

	for k, v := range *ic.ModelStore {
		modelStore[string(k)] = string(v)
	}

	for _, job := range ic.Scheduler.ActiveJobs {
		activeJobs = append(activeJobs, job)
	}

	for _, job := range ic.Scheduler.CompletedJobs {
		completedJobs = append(completedJobs, job)
	}

	for _, job := range ic.Scheduler.PendingJobs.ToSlice() {
		pendingJobs = append(pendingJobs, job)
	}

	backup := &api.CoordinatorBackup{
		ModelStore:    modelStore,
		ActiveJobs:    activeJobs,
		CompletedJobs: completedJobs,
		PendingJobs:   pendingJobs,
	}

	logger.Info("Sending backup data ...")
	_, err = client.Backup(context.Background(), &api.BackupRequest{
		Backup: backup,
	})
	if err != nil {
		logger.Error("Failed to backup coordinator data: " + err.Error())
	}
	logger.Info("Backup data sent")
}

// For all batches that are scheduled in failed workers
// reset their status to available
func (ic *IDunnoCoordinator) RefreshBatchStatus() {
	ic.Lock()
	defer ic.Unlock()

	ic.Scheduler.RefreshBatchStatus()
}

// For all workers that are not running on a particular
// job, reset it to idle state
func (ic *IDunnoCoordinator) RefreshWorkerStatus() {
	ic.Lock()
	defer ic.Unlock()

	ic.Scheduler.RefreshWorkerStatus()
}

// Process completed job by sending its outoput to SDFS
// and add to CompletedJobs list
func (ic *IDunnoCoordinator) FlushPendingJobs() {
	ic.Lock()
	defer ic.Unlock()

	if ic.Scheduler.PendingJobs.Empty() {
		return
	}

	job := ic.Scheduler.PendingJobs.Top()
	delete(ic.Scheduler.ActiveJobs, job.Id)

	workers := ic.ResourceManager.GetWorkersById(job.Id)
	wg := sync.WaitGroup{}

	// send finish inference requests
	for _, worker := range workers {
		wg.Add(1)
		go func(worker *Worker) {
			defer worker.Reset()
			defer wg.Done()

			client, close, err := ic.CreateWorkerClient(worker.Process)
			if err != nil {
				logger.Error("Trying to finish inference on worker " + worker.Process.Address() + " failed: " + err.Error())
				return
			}
			defer close()

			_, err = client.FinishInference(context.Background(), &api.FinishInferenceRequest{})
			if err != nil {
				logger.Error("Trying to finish inference on worker " + worker.Process.Address() + " failed: " + err.Error())
				return
			}
		}(worker)
	}
	wg.Wait()

	logger.Info("Job " + job.Id + " completed, processing results...")

	results, metric := job.GetResults()
	lines := make([]string, 0)
	for _, result := range results {
		lines = append(lines, fmt.Sprintf("%s %s", result.GetInput(), result.GetOutput()))
	}
	lines = append(lines, fmt.Sprintf("\n%f", metric))

	err := ic.SDFSClient.WriteLocalFile(job.Id, []byte(strings.Join(lines, "\n")))
	if err != nil {
		logger.Error("Failed to write results to SDFS: " + err.Error())
		return
	}
	defer ic.SDFSClient.DeleteLocalFile(job.Id)

	err = ic.SDFSClient.Put(job.Id, job.Id)
	if err != nil {
		logger.Error("Failed to put results to SDFS: " + err.Error())
		return
	}

	ic.Scheduler.PendingJobs.Pop()
	ic.Scheduler.CompletedJobs[job.Id] = job
	job.FinishTime = api.CurrentTimestamp()

	logger.Info("Job " + job.Id + " completed!")
	logger.Info("Pending job len: " + fmt.Sprint(ic.Scheduler.PendingJobs.Len()))
	logger.Info("Completed job len: " + fmt.Sprint(len(ic.Scheduler.CompletedJobs)))
}

// Core function of the coordinator that  reschedule
// job to different workers in a fair-time fashion
func (ic *IDunnoCoordinator) RescheduleJobs() {
	ic.SetSchedulingStatus(true)

	ic.Lock()
	schedule := ic.Scheduler.RefreshSchedule()
	ic.Unlock()
	if len(schedule) == 0 {
		return
	}

	for jobId, workers := range schedule {
		for _, worker := range workers {
			logger.Info(fmt.Sprintf("Sending job %v to %v", jobId, worker.Process.Address()))

			go func(jobId string, worker *Worker) {
				// create worker client
				client, close, err := ic.CreateWorkerClient(worker.Process)
				if err != nil {
					return
				}
				defer close()

				// construct request struct
				req := &api.InferenceRequest{
					InferenceTask: &api.InferenceTask{
						Model:     string(ic.Scheduler.ActiveJobs[jobId].ModelType),
						BatchSize: int32(ic.Scheduler.ActiveJobs[jobId].BatchSize),
					},
					JobId: jobId,
				}

				// send start inference request
				res, err := client.Inference(context.Background(), req)
				if err != nil {
					logger.Error(fmt.Sprintf("Failed to reschedule job %v to worker %v: error: %v", jobId, worker.Process, err))
					return
				}

				if res.GetStatus() != api.ResponseStatus_OK {
					logger.Error(fmt.Sprintf("Failed to reschedule job %v to worker %v: status: %v", jobId, worker.Process, res.GetStatus()))
					return
				}

				// update worker info
				ic.Lock()
				worker.JobId = jobId
				worker.LastQueryTime = api.CurrentTimestamp()
				ic.Unlock()
				logger.Schedule(jobId, worker.Process.Address())
			}(jobId, worker)
		}
	}

	time.Sleep(200 * time.Millisecond)

	ic.SetSchedulingStatus(false)
}

// Process queued inference request in FIFO order
func (ic *IDunnoCoordinator) ProcessQueuedJob() {
	if ic.TaskQueue.Empty() {
		return
	}

	ic.Lock()
	task := ic.TaskQueue.Top()
	ic.TaskQueue.Pop()
	ic.Unlock()

	model, batchSize := utils.ModelType(task.GetModel()), int(task.GetBatchSize())

	// fetch dataset folder (containing list of sdfs filenames)
	dataset := ic.ModelStore.GetDataset(model)
	localFile := utils.CreateTempFilename()

	// if dataset is not found, meaning it is deleted before inference starts, simply return
	err := ic.SDFSClient.Get(localFile, string(dataset), sdfs.LATEST_VERSION)
	if err != nil {
		logger.Error("Failed to get dataset from SDFS: " + err.Error())
		return
	}

	// read dataset file from temporary local file
	data, err := ic.SDFSClient.ReadLocalFile(localFile)
	if err != nil {
		logger.Error("Failed to read local file: " + err.Error())
		return
	}
	defer ic.SDFSClient.DeleteLocalFile(localFile)

	// split by "\n" to get list of sdfs filenames (or raw inputs)
	inputs := strings.Split(string(data), "\n")
	logger.Info(fmt.Sprintf("Dataset %v has %v inputs", dataset, len(inputs)))

	// split inputs into a set of batches
	batchStates := make([]*api.BatchState, 0)
	for i := 0; i < len(inputs); i += batchSize {
		end := int(math.Min(float64(i+batchSize), float64(len(inputs))))
		batchStates = append(batchStates, &api.BatchState{
			Status: api.BatchStatus_Available,
			BatchInput: &api.BatchInput{
				BatchId: int32(len(batchStates)),
				Inputs:  inputs[i:end],
			},
			BatchOutput: nil,
			QueryTime:   nil,
			ReceiveTime: nil,
		})
	}

	// create job info struct
	prefix := fmt.Sprintf("%v:%v", model, batchSize)
	job := &api.Job{
		Id:                utils.CreateId(prefix),
		ModelType:         string(model),
		BatchSize:         int32(batchSize),
		Dataset:           string(dataset),
		StartTime:         api.CurrentTimestamp(),
		FinishTime:        nil,
		TotalQueries:      int32(math.Ceil(float64(len(inputs)) / float64(batchSize))),
		CompletedQueries:  0,
		BatchStates:       batchStates,
		QueryRates:        make([]float32, 0),
		QueryProcessTimes: make([]float32, 0),
	}

	ic.Lock()
	ic.Scheduler.AddJob(job)
	ic.Unlock()

	logger.NewJob(job)
}

// Callback function for worker that is called when a machine
// becomes the coordinator
func (ic *IDunnoCoordinator) OnBecomeCoordinator() {
	ic.Lock()
	defer ic.Unlock()

	if ic.IsCoordinator {
		return
	}

	defer logger.Info("I am the new coordinator! My address is " + ic.Ring.Address())
	ic.IsCoordinator = true
}

// Callback function for worker that is called when
// ring's membership changes
func (ic *IDunnoCoordinator) OnMemberUpdate(action ring.MemAction, process *api.Process) {
	ic.Lock()
	defer ic.Unlock()

	switch action {
	case ring.MEMBER_INSERT:
		ic.Scheduler.OnWorkerJoined(process)
	case ring.MEMBER_DELETE:
		ic.Scheduler.OnWorkerFailed(process)
	case ring.MEMBER_LEAVED:
		ic.Scheduler.OnWorkerLeaved(process)
	}
}

func (ic *IDunnoCoordinator) CreateCoordinatorClient() (api.CoordinatorServiceClient, func(), error) {
	successors := ic.Ring.Successors()

	if len(successors) == 0 {
		return nil, nil, errors.New("not enough successors")
	}

	next := successors[0]
	conn, err := grpc.Dial(next.Address(), sdfs.GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to connect to successor: " + err.Error())
		return nil, nil, err
	}

	return api.NewCoordinatorServiceClient(conn), func() { conn.Close() }, nil
}

func (ic *IDunnoCoordinator) CreateWorkerClient(sendToProcess *api.Process) (api.WorkerServiceClient, func(), error) {
	conn, err := grpc.Dial(sendToProcess.Address(), sdfs.GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to connect to worker service: " + err.Error())
		return nil, nil, err
	}

	return api.NewWorkerServiceClient(conn), func() { conn.Close() }, nil
}

func (ic *IDunnoCoordinator) PrintWorkers() string {
	ic.Lock()
	defer ic.Unlock()

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Worker Address",
		"Join Time",
		"Running Job",
		"Idle",
		"Last Query Time",
	})
	for _, worker := range *ic.ResourceManager {
		t.AppendRow(table.Row{
			worker.Process.Address(),
			worker.Process.JoinTime.AsTime().Format("2006-01-02 15:04:05"),
			worker.JobId,
			worker.Idle(),
			worker.LastQueryTime.AsTime().Format("2006-01-02 15:04:05"),
		})
	}
	t.AppendFooter(table.Row{
		"Total Workers",
		len(*ic.ResourceManager),
	})

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle
	t.SortBy([]table.SortBy{{Name: "Join Time", Mode: table.Asc}})

	return t.Render()
}

func (ic *IDunnoCoordinator) PrintWorkersJSON() string {
	ic.Lock()
	defer ic.Unlock()

	responses := make([]map[string]interface{}, 0)
	for _, worker := range *ic.ResourceManager {
		responses = append(responses, map[string]interface{}{
			"address":       worker.Process.Address(),
			"joinTime":      worker.Process.JoinTime.AsTime().Format("2006-01-02 15:04:05"),
			"runningJob":    worker.JobId,
			"lastQueryTime": worker.LastQueryTime.AsTime().Format("2006-01-02 15:04:05"),
		})
	}

	marshalled, err := json.Marshal(responses)
	if err != nil {
		logger.Error("Failed to marshal workers to JSON: " + err.Error())
		return utils.EMPTY_STRING
	}

	return string(marshalled)
}

func (ic *IDunnoCoordinator) PrintJobs() string {
	ic.Lock()
	defer ic.Unlock()

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Job ID",
		"Model Type",
		"Batch Size",
		"Total Queries",
		"Completed Queries",
		"Total Query Time",
		"Running VMs",
		"Progress",
		"Query/Sec",
		"Time Left",
	})

	qps := make([]float64, 0)
	for _, job := range ic.Scheduler.ActiveJobs {
		workerCount := ic.ResourceManager.GetWorkerCountById(job.Id)
		timeLeft := job.GetExpectedTimeLeft(workerCount)

		if timeLeft == math.MaxFloat64 {
			timeLeft = 99999
		}

		queries := 0.0
		if utils.LAST_SECONDS == math.MaxFloat64 {
			queries = job.GetExpectedQPS(workerCount)
		} else {
			queries = job.GetQPS(utils.LAST_SECONDS)
		}

		qps = append(qps, queries)
		t.AppendRow(table.Row{
			job.Id,
			job.ModelType,
			job.BatchSize,
			job.TotalQueries,
			job.CompletedQueries,
			fmt.Sprintf("%.2f sec", job.TotalQueryTime().Seconds()),
			workerCount,
			fmt.Sprintf("%.2f%%", float64(job.CompletedQueries)/float64(job.TotalQueries)*100),
			fmt.Sprintf("%.2f", queries),
			fmt.Sprintf("%.2f sec", timeLeft),
		})
	}

	t.AppendFooter(table.Row{
		"Total Running Jobs",
		len(ic.Scheduler.ActiveJobs),
	})

	t.AppendFooter(table.Row{
		"Relative QPS Difference",
		fmt.Sprintf("%.2f%%", ralloc.GetRelQPSDiff(qps)*100),
	})

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle
	t.SortBy([]table.SortBy{{Name: "Job ID", Mode: table.Asc}})

	return t.Render()
}

func (ic *IDunnoCoordinator) PrintJobsJSON() string {
	ic.Lock()
	defer ic.Unlock()
	qps := make([]float64, 0)
	jobs := make([]map[string]interface{}, 0)
	for _, job := range ic.Scheduler.ActiveJobs {
		workerCount := ic.ResourceManager.GetWorkerCountById(job.Id)
		timeLeft := job.GetExpectedTimeLeft(workerCount)

		if timeLeft == math.MaxFloat64 {
			timeLeft = 99999
		}

		queries := 0.0
		if utils.LAST_SECONDS == math.MaxFloat64 {
			queries = job.GetExpectedQPS(workerCount)
		} else {
			queries = job.GetQPS(utils.LAST_SECONDS)
		}
		qps = append(qps, queries)
		jobs = append(jobs, map[string]interface{}{
			"id":               job.Id,
			"modelType":        job.ModelType,
			"batchSize":        job.BatchSize,
			"totalQueries":     job.TotalQueries,
			"completedQueries": job.CompletedQueries,
			"totalQueryTime":   job.TotalQueryTime().Seconds(),
			"runningVMs":       workerCount,
			"progress":         float64(job.CompletedQueries) / float64(job.TotalQueries) * 100,
			"qps":              queries,
			"timeLeft":         timeLeft,
		})
	}

	var response = map[string]interface{}{
		"jobs":                  jobs,
		"relativeQPSDifference": ralloc.GetRelQPSDiff(qps) * 100,
	}

	marshalled, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal jobs to JSON: " + err.Error())
		return utils.EMPTY_STRING
	}

	return string(marshalled)
}

func (ic *IDunnoCoordinator) PrintCompletedJobs() string {
	ic.Lock()
	defer ic.Unlock()

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Job ID",
		"Model Type",
		"Batch Size",
		"Total Queries",
		"Total Query Time",
		"Query/Sec",
	})

	for _, job := range ic.Scheduler.CompletedJobs {
		t.AppendRow(table.Row{
			job.Id,
			job.ModelType,
			job.BatchSize,
			job.TotalQueries,
			fmt.Sprintf("%.2f sec", job.TotalQueryTime().Seconds()),
			fmt.Sprintf("%.2f", float64(job.TotalQueries)/job.TotalQueryTime().Seconds()),
		})
	}

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle
	t.SortBy([]table.SortBy{{Name: "Job ID", Mode: table.Asc}})

	return t.Render()
}

func (ic *IDunnoCoordinator) PrintCompletedJobsJSON() string {
	ic.Lock()
	defer ic.Unlock()

	responses := make([]map[string]interface{}, 0)
	for _, job := range ic.Scheduler.CompletedJobs {
		responses = append(responses, map[string]interface{}{
			"id":             job.Id,
			"modelType":      job.ModelType,
			"batchSize":      job.BatchSize,
			"totalQueries":   job.TotalQueries,
			"totalQueryTime": job.TotalQueryTime().Seconds(),
			"qps":            float64(job.TotalQueries) / job.TotalQueryTime().Seconds(),
		})
	}

	marshalled, err := json.Marshal(responses)
	if err != nil {
		logger.Error("Failed to marshal completed jobs to JSON: " + err.Error())
		return utils.EMPTY_STRING
	}

	return string(marshalled)
}

func (ic *IDunnoCoordinator) PrintJob(jobId string) string {
	ic.Lock()
	defer ic.Unlock()

	job := ic.Scheduler.GetJob(jobId)
	if job == nil {
		job = ic.Scheduler.CompletedJobs[jobId]
		if job == nil {
			return "Job " + jobId + " not found\n"
		}
	}

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Input",
		"Output",
	})

	results, metric := job.GetResults()
	for _, result := range results {
		// limit each input length to 100 characters
		t.AppendRow(table.Row{
			fmt.Sprintf("%.100s...", result.Input),
			fmt.Sprintf("%.100s", result.Output),
		})
	}

	t.AppendFooter(table.Row{
		"Metric",
		fmt.Sprintf("%.2f%%", metric*100),
	})

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle

	return t.Render()
}

func (ic *IDunnoCoordinator) PrintJobJSON(jobId string) string {
	ic.Lock()
	defer ic.Unlock()

	job := ic.Scheduler.GetJob(jobId)
	if job == nil {
		job = ic.Scheduler.CompletedJobs[jobId]
		if job == nil {
			return utils.EMPTY_STRING
		}
	}

	batches := make([]map[string]interface{}, 0)
	results, metric := job.GetResults()
	for _, result := range results {

		batches = append(batches, map[string]interface{}{
			"batchInput":  result.Input,
			"batchOutput": result.Output,
		})
	}
	var response = map[string]interface{}{
		"metric":            metric,
		"batches":           batches,
		"id":                job.Id,
		"queryRates":        job.QueryRates,
		"queryProcessTimes": job.QueryProcessTimes,
	}

	marshalled, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal job to JSON: " + err.Error())
		return utils.EMPTY_STRING
	}

	return string(marshalled)
}

func (ic *IDunnoCoordinator) PrintJobStatistics(jobId string) string {
	ic.Lock()
	defer ic.Unlock()

	job := ic.Scheduler.GetJob(jobId)
	if job == nil {
		job = ic.Scheduler.CompletedJobs[jobId]
		if job == nil {
			return "Job " + jobId + " not found\n"
		}
	}

	qps := job.QueryRates
	time := job.QueryProcessTimes

	if qps == nil {
		qps = make([]float32, 0)
	}

	if time == nil {
		time = make([]float32, 0)
	}

	qpsData := stats.LoadRawData(utils.ConvertTo64(qps))
	timeData := stats.LoadRawData(utils.ConvertTo64(time))

	t := table.NewWriter()

	t.AppendHeader(table.Row{
		"Stat",
		"Query Rate",
		"Query Processing Time",
	})

	qpsMean, _ := stats.Mean(qpsData)
	qpsMedian, _ := stats.Median(qpsData)
	qpsStdDev, _ := stats.StandardDeviation(qpsData)
	qpsP90, _ := stats.Percentile(qpsData, 90)
	qpsP95, _ := stats.Percentile(qpsData, 95)
	qpsP99, _ := stats.Percentile(qpsData, 99)

	timeMean, _ := stats.Mean(timeData)
	timeMedian, _ := stats.Median(timeData)
	timeStdDev, _ := stats.StandardDeviation(timeData)
	timeP90, _ := stats.Percentile(timeData, 90)
	timeP95, _ := stats.Percentile(timeData, 95)
	timeP99, _ := stats.Percentile(timeData, 99)

	// mean
	t.AppendRow(table.Row{
		"Mean",
		fmt.Sprintf("%.2f", qpsMean),
		fmt.Sprintf("%.2f", timeMean),
	})

	// median
	t.AppendRow(table.Row{
		"Median",
		fmt.Sprintf("%.2f", qpsMedian),
		fmt.Sprintf("%.2f", timeMedian),
	})

	// standard deviation
	t.AppendRow(table.Row{
		"Standard Deviation",
		fmt.Sprintf("%.2f", qpsStdDev),
		fmt.Sprintf("%.2f", timeStdDev),
	})

	// 90-percentile
	t.AppendRow(table.Row{
		"90-percentile",
		fmt.Sprintf("%.2f", qpsP90),
		fmt.Sprintf("%.2f", timeP90),
	})

	// 95-percentile
	t.AppendRow(table.Row{
		"95-percentile",
		fmt.Sprintf("%.2f", qpsP95),
		fmt.Sprintf("%.2f", timeP95),
	})

	// 99-percentile
	t.AppendRow(table.Row{
		"99-percentile",
		fmt.Sprintf("%.2f", qpsP99),
		fmt.Sprintf("%.2f", timeP99),
	})

	t.SetAutoIndex(true)
	t.SetStyle(table.StyleLight)
	t.Style().Format.Header = text.FormatTitle
	t.Style().Format.Footer = text.FormatTitle

	return t.Render()
}

func (ic *IDunnoCoordinator) IsServing() bool {
	return len(ic.Scheduler.ActiveJobs) > 0
}

func (ic *IDunnoCoordinator) SetSchedulingStatus(status bool) {
	ic.Lock()
	defer ic.Unlock()
	ic.IsScheduling = status
}
