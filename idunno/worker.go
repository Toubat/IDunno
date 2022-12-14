package main

import (
	"context"
	"errors"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/ring"
	"mp4/sdfs"
	"mp4/utils"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

const RUNNER_PORT_OFFSET = 1000
const RESTART_QUERY_INTERVAL = 1000 * time.Millisecond
const QUERY_INTERVAL = 800 * time.Millisecond
const QUERY_DATA_DEADLINE = 2500 * time.Millisecond

type IDunnoWorker struct {
	ModelRunner *api.Process
	JobId       string
	BatchOutput *api.BatchOutput
	SDFSClient  *sdfs.SDFSClient
	Ring        *ring.RingServer
	api.WorkerServiceServer
	sync.Mutex
}

func NewIDunnoWorker(hostname string, port int, ring *ring.RingServer, sdfsClient *sdfs.SDFSClient) *IDunnoWorker {
	// initiate python gRPC server for model inference
	runnerPort := port + RUNNER_PORT_OFFSET
	runnerCmd := exec.Command("python3.10", "../inference/worker.py", "--port", strconv.Itoa(runnerPort), "--filepath", sdfsClient.GetLocalFilePath(""))
	runnerCmd.Stderr = os.Stderr
	// runnerCmd.Stdout = os.Stdout

	// start python inference server
	err := runnerCmd.Start()
	if err != nil {
		fmt.Println("Failed to start inference service: " + err.Error())
		panic(err)
	}

	runner := &api.Process{
		Ip:   hostname,
		Port: int32(runnerPort),
	}

	return &IDunnoWorker{
		ModelRunner: runner,
		JobId:       utils.EMPTY_STRING,
		BatchOutput: nil,
		SDFSClient:  sdfsClient,
		Ring:        ring,
	}
}

func (iw *IDunnoWorker) Cron() {
	go iw.QueryDataCycle()
}

func (iw *IDunnoWorker) QueryDataCycle() {
	for {
		err := iw.QueryData()
		time.Sleep(QUERY_INTERVAL)
		if err != nil {
			time.Sleep(RESTART_QUERY_INTERVAL)
		}
	}
}

func (iw *IDunnoWorker) QueryData() error {
	// coordinator client
	coordClient, closeCoord, err := iw.CreateCoordinatorClient()
	if err != nil {
		logger.Error("Failed to create coordinator client: " + err.Error())
		return err
	}
	defer closeCoord()

	// inference client
	runnerClient, closeRunner, err := iw.CreateInferenceClient()
	if err != nil {
		logger.Error("Failed to create inference client: " + err.Error())
		return err
	}
	defer closeRunner()

	iw.Lock()
	defer iw.Unlock()

	// stop querying if no job is assigned
	if iw.JobId == utils.EMPTY_STRING {
		time.Sleep(time.Second)
		return nil
	}

	// set up context deadline
	ctx, cancel := context.WithTimeout(context.Background(), QUERY_DATA_DEADLINE)
	defer cancel()

	// query data from coordinator
	res, err := coordClient.QueryData(ctx, &api.QueryDataRequest{
		JobId:       iw.JobId,
		Worker:      iw.Ring.Process,
		BatchOutput: iw.BatchOutput,
	})
	if err != nil {
		return err
	}

	// no new batch is given, skip current round
	if res.GetBatchInput() == nil {
		return nil
	}

	if res.GetBatchInput().GetInputs() == nil {
		iw.BatchOutput = &api.BatchOutput{
			BatchId: res.GetBatchInput().GetBatchId(),
			Results: nil,
			Metric:  0,
		}
		return nil
	}

	modelInputs := make([]string, 0)
	if res.GetIsFilename() {
		wg := sync.WaitGroup{}

		// fetch all SDFS filenmaes into local file system
		for _, input := range res.GetBatchInput().GetInputs() {
			wg.Add(1)
			go func(input string) {
				defer wg.Done()

				// fetch file from SDFS
				err := iw.SDFSClient.Get(input, input, sdfs.LATEST_VERSION)
				if err != nil {
					logger.Error("Worker failed to fetch SDFS file " + input + ": " + err.Error())
					return
				}

				modelInputs = append(modelInputs, iw.SDFSClient.GetLocalFilePath(input))
			}(input)
		}
		wg.Wait()

		defer func() {
			logger.Info("Cleaning up temp files")
			// delete all files in local file system
			for _, file := range res.GetBatchInput().GetInputs() {
				err := iw.SDFSClient.DeleteLocalFile(file)
				if err != nil {
					logger.Error("Worker failed to delete local file " + file + ": " + err.Error())
				}
			}
			logger.Info("Done cleaning up temp files")
		}()
	} else {
		modelInputs = res.GetBatchInput().GetInputs()
	}

	evalRes, err := runnerClient.Evaluate(context.Background(), &api.EvaluateRequest{
		Inputs: modelInputs,
	})
	if err != nil {
		logger.Error("Worker failed to evaluate model: " + err.Error())
		return err
	}
	if evalRes.GetStatus() != api.ResponseStatus_OK {
		logger.Error("Worker failed to evaluate model: " + evalRes.GetStatus().String())
		return errors.New(evalRes.GetStatus().String())
	}

	iw.BatchOutput = &api.BatchOutput{
		BatchId: res.GetBatchInput().GetBatchId(),
		Results: evalRes.GetResults(),
		Metric:  evalRes.GetMetric(),
	}
	return nil
}

func (iw *IDunnoWorker) CreateInferenceClient() (api.InferenceServiceClient, func(), error) {
	conn, err := grpc.Dial(iw.ModelRunner.Address(), sdfs.GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to connect to inference service: " + err.Error())
		return nil, nil, err
	}

	return api.NewInferenceServiceClient(conn), func() { conn.Close() }, nil
}

func (iw *IDunnoWorker) CreateCoordinatorClient() (api.CoordinatorServiceClient, func(), error) {
	leaderAddr, err := iw.Ring.LookupLeader()
	if err != nil {
		logger.Error("Failed to lookup leader")
		return nil, nil, err
	}

	conn, err := grpc.Dial(leaderAddr, sdfs.GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to connect to inference service: " + err.Error())
		return nil, nil, err
	}

	return api.NewCoordinatorServiceClient(conn), func() { conn.Close() }, nil
}
