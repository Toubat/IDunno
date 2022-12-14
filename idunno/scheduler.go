package main

import (
	"fmt"
	"math"
	"mp4/api"
	"mp4/logger"
	"mp4/ralloc"
	"mp4/utils"
	"sort"
)

type IDunnoScheduler struct {
	ActiveJobs      map[string]*api.Job
	PendingJobs     *utils.Queue[*api.Job]
	CompletedJobs   map[string]*api.Job
	ResourceManager *ResourceManager
}

func NewIDunnoScheduler(rm *ResourceManager) *IDunnoScheduler {
	return &IDunnoScheduler{
		ActiveJobs:      make(map[string]*api.Job),
		PendingJobs:     utils.NewQueue[*api.Job](),
		CompletedJobs:   make(map[string]*api.Job),
		ResourceManager: rm,
	}
}

func (is *IDunnoScheduler) AddJob(job *api.Job) {
	logger.Info(fmt.Sprintf("Adding job %v to scheduler", job.Id))

	is.ActiveJobs[job.Id] = job
}

func (is *IDunnoScheduler) GetJob(jobId string) *api.Job {
	if job, ok := is.ActiveJobs[jobId]; ok {
		return job
	}
	return nil
}

func (is *IDunnoScheduler) RefreshWorkerStatus() {
	for _, worker := range *is.ResourceManager {
		if _, ok := is.ActiveJobs[worker.JobId]; !ok {
			worker.Reset()
		}
	}
}

func (is *IDunnoScheduler) RefreshBatchStatus() {
	for _, job := range is.ActiveJobs {
		workers := is.ResourceManager.GetWorkersById(job.Id)

		for id, batchState := range job.BatchStates {
			if batchState.Status != api.BatchStatus_InProgress {
				continue
			}

			inProgress := false
			for _, worker := range workers {
				if worker.CurrBatchInput != nil && worker.CurrBatchInput.GetBatchId() == int32(id) {
					inProgress = true
					break
				}
			}

			// if no worker is working on this batch input, then it is failed
			// make it available again
			if !inProgress {
				batchState.Status = api.BatchStatus_Available
			}
		}
	}
}

func (is *IDunnoScheduler) RefreshSchedule() map[string][]*Worker {
	if is.ResourceManager.Len() == 0 || len(is.ActiveJobs) == 0 {
		return make(map[string][]*Worker)
	}

	jobs := make([]*api.Job, 0)      // all jobs
	ids := make([]string, 0)         // job ids
	allocMap := make(map[string]int) // #VMs to allocate for each job

	for id := range is.ActiveJobs {
		ids = append(ids, id)
	}
	// important to sort id first to make ralloc output to be deterministic
	sort.Strings(ids)

	for _, id := range ids {
		jobs = append(jobs, is.ActiveJobs[id])
	}

	var resources []int
	if utils.LAST_SECONDS == math.MaxFloat64 {
		// Global fair-time scheduling
		qps := ralloc.JobToQPS(jobs, is.ResourceManager.Len())
		resources, _ = ralloc.GlobalFairTimeRalloc(len(jobs), is.ResourceManager.Len(), qps)
	} else {
		// Local fair-time scheduling
		resources, _ = ralloc.LocalFairTimeRalloc(jobs, is.ResourceManager.Len())
	}

	for i := range jobs {
		allocMap[ids[i]] = resources[i]
	}
	// make worker that are no longer needs to be used idle
	for _, job := range jobs {
		workers := is.ResourceManager.GetWorkersById(job.Id)
		sort.Sort(WorkerList(workers)) // sort by lastQueryTime (most recent first)

		for i := 0; i < len(workers)-allocMap[job.Id]; i++ {
			workers[i].JobId = utils.EMPTY_STRING

			if workers[i].CurrBatchInput == nil {
				continue
			}

			// this batch input never finishes, so we make it available again
			batchId := workers[i].CurrBatchInput.GetBatchId()
			job.BatchStates[batchId].Status = api.BatchStatus_Available
			workers[i].CurrBatchInput = nil
		}
	}


	// allocate idle workers to jobs
	idleWorkers, idx := is.ResourceManager.GetIdleWorkers(), 0
	newSchedule := make(map[string][]*Worker) // job id -> worker list
	for _, job := range jobs {
		workerCount := is.ResourceManager.GetWorkerCountById(job.Id)

		for workerCount < allocMap[job.Id] {
			if idx >= len(idleWorkers) {
				logger.Error("Not enough idle workers to allocate")
				break
			}

			newSchedule[job.Id] = append(newSchedule[job.Id], idleWorkers[idx])
			workerCount++
			idx++
		}
	}

	if idx != len(idleWorkers) {
		logger.Error("did not use all VM resources")
	}

	return newSchedule
}

func (is *IDunnoScheduler) OnReceiveBatchOutput(jobId string, workerProcess *api.Process, batchOutput *api.BatchOutput) {
	logger.Info(fmt.Sprintf("Received batch output for job %v", jobId))

	worker := is.ResourceManager.GetWorker(workerProcess.Address())
	job := is.GetJob(jobId)

	if job == nil {
		logger.Error("job " + jobId + " not found")
		return
	}

	// update job info
	batchId := batchOutput.GetBatchId()
	job.BatchStates[batchId].BatchOutput = batchOutput
	job.BatchStates[batchId].Status = api.BatchStatus_Completed
	job.BatchStates[batchId].ReceiveTime = api.CurrentTimestamp()
	job.CompletedQueries = int32(job.GetCompletedBatchCount())

	// update worker info
	worker.LastQueryTime = api.CurrentTimestamp()
	worker.CurrBatchInput = nil

	if job.CompletedQueries < job.TotalQueries {
		return
	}

	for _, job := range *is.PendingJobs {
		if job.Id == jobId {
			return
		}
	}

	is.PendingJobs.Push(job)
}

func (is *IDunnoScheduler) OnWorkerFailed(process *api.Process) {
	logger.Info(fmt.Sprintf("Worker %v failed", process.Address()))

	// remove worker from resource manager
	failedWorker := is.ResourceManager.RemoveWorker(process)

	if failedWorker == nil || failedWorker.JobId == utils.EMPTY_STRING || failedWorker.CurrBatchInput == nil {
		return
	}

	batchId := failedWorker.CurrBatchInput.GetBatchId()
	job := is.GetJob(failedWorker.JobId)

	if job == nil {
		return
	}

	// make batch input available again
	job.BatchStates[batchId].Status = api.BatchStatus_Available
}

func (is *IDunnoScheduler) OnWorkerJoined(process *api.Process) {
	logger.Info(fmt.Sprintf("Worker %v joined", process.Address()))

	// add worker to resource manager
	is.ResourceManager.AddWorker(process)
}

func (is *IDunnoScheduler) OnWorkerLeaved(process *api.Process) {
	logger.Info(fmt.Sprintf("Worker %v leaved", process.Address()))

	// clear all workers
	*is.ResourceManager = make(map[string]*Worker)
	is.ActiveJobs = make(map[string]*api.Job)
}
