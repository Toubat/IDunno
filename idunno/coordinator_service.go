package main

import (
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/utils"
)

func (ic *IDunnoCoordinator) Train(ctx context.Context, req *api.TrainRequest) (*api.TrainResponse, error) {
	logger.Info(fmt.Sprintf("Received Train request - Model: %v, Dataset: %v", req.GetTrainTask().GetModel(), req.GetTrainTask().GetDataset()))
	ic.OnBecomeCoordinator()

	ic.Lock()
	if ic.IsServing() {
		logger.Error("Cannot train while serving inference requests")
		ic.Unlock()
		return nil, fmt.Errorf("cannot train while serving inference requests")
	}
	ic.Unlock()

	// send train request to all worker machines
	logger.Info("Sending train request to all worker machines...")

	memList := ic.Ring.GetMembershipList()
	resChan, quitChan := make(chan *api.TrainResponse), make(chan error)

	for _, process := range memList {
		go func(process *api.Process) {
			client, close, err := ic.CreateWorkerClient(process)
			if err != nil {
				quitChan <- err
				return
			}
			defer close()

			res, err := client.Train(context.Background(), req)
			if err != nil {
				logger.Error(fmt.Sprintf("Error sending train request to worker: %s: %s", process.Address(), err.Error()))
				quitChan <- err
				return
			}

			resChan <- res
		}(process)
	}

	// wait for all workers to finish training
	ackRequired, ackReceived := len(memList), 0
	for {
		if ackReceived >= ackRequired {
			logger.Info("All workers finished training...")
			break
		}

		select {
		case <-resChan:
			ackReceived++
		case err := <-quitChan:
			return nil, err
		}
	}

	// add model to model store
	ic.Lock()
	defer ic.Unlock()
	ic.ModelStore.AddModel(utils.ModelType(req.GetTrainTask().GetModel()), utils.DatasetType(req.GetTrainTask().GetDataset()))

	// TODO: forward message to backup coordinator
	// ...

	return &api.TrainResponse{}, nil
}

func (ic *IDunnoCoordinator) Inference(ctx context.Context, req *api.InferenceRequest) (*api.InferenceResponse, error) {
	logger.Info(fmt.Sprintf("Received Inference request - Model: %v, Batch Size: %v", req.GetInferenceTask().GetModel(), req.GetInferenceTask().GetBatchSize()))
	ic.OnBecomeCoordinator()

	if !ic.ModelStore.Contains(utils.ModelType(req.GetInferenceTask().GetModel())) {
		logger.Error(fmt.Sprintf("Model %v has not been trained", req.GetInferenceTask().GetModel()))
		return nil, fmt.Errorf("model %v has not been trained", req.GetInferenceTask().GetModel())
	}

	// add inference task to task queue
	ic.Lock()
	defer ic.Unlock()
	ic.TaskQueue.Push(req.GetInferenceTask())

	// TODO: forward message to backup coordinator
	// ...

	return &api.InferenceResponse{
		Status: api.ResponseStatus_OK,
	}, nil
}

func (ic *IDunnoCoordinator) QueryData(ctx context.Context, req *api.QueryDataRequest) (*api.QueryDataResponse, error) {
	logger.Query(fmt.Sprintf("Received QueryData request - Job ID: %v", req.GetJobId()))
	ic.OnBecomeCoordinator()

	// TODO: forward message to backup coordinator
	// ...
	ic.Lock()
	defer ic.Unlock()

	worker := ic.Scheduler.ResourceManager.GetWorker(req.GetWorker().Address())
	job := ic.Scheduler.GetJob(req.GetJobId())

	if worker == nil {
		logger.Error(fmt.Sprintf("Trying to query data, but worker %v not found", req.GetWorker().Address()))
		return nil, fmt.Errorf("trying to query data, but worker %v not found", req.GetWorker().Address())
	}
	if job == nil {
		logger.Error(fmt.Sprintf("Trying to query data, but job %v not found", req.GetJobId()))
		return nil, fmt.Errorf("trying to query data, but job %v not found", req.GetJobId())
	}

	if req.GetBatchOutput() != nil {
		ic.Scheduler.OnReceiveBatchOutput(req.GetJobId(), req.GetWorker(), req.GetBatchOutput())
	}

	// job id mismatch
	if worker.JobId != req.GetJobId() {
		logger.Error(fmt.Sprintf("Trying to query data with job id %v, but worker %v is not assigned to job %v", req.GetJobId(), req.GetWorker().Address(), worker.JobId))
		return nil, fmt.Errorf("trying to query data with job id %v, but worker %v is not assigned to job %v", req.GetJobId(), req.GetWorker().Address(), worker.JobId)
	}
	// idle worker
	if worker.Idle() {
		logger.Error(fmt.Sprintf("Trying to query data from idle worker %v", req.GetWorker().Address()))
		return nil, fmt.Errorf("trying to query data from idle worker %v", req.GetWorker().Address())
	}

	// validate SDFS input dataset
	batchInput := job.FetchBatchInput()

	if batchInput != nil {
		inputs := batchInput.GetInputs()
		batchInput.Inputs = ValidateDatasetInputs(utils.DatasetType(job.Dataset), inputs)
	}

	worker.CurrBatchInput = batchInput
	return &api.QueryDataResponse{
		BatchInput: worker.CurrBatchInput,
		IsFilename: ContainFilenames(utils.DatasetType(job.Dataset)),
	}, nil
}

func (ic *IDunnoCoordinator) Backup(ctx context.Context, req *api.BackupRequest) (*api.BackupResponse, error) {
	logger.Info("Received Backup request")

	ic.Lock()
	defer ic.Unlock()

	if req.GetBackup() == nil {
		logger.Info("Backup request is nil")
	}

	if req.GetBackup().GetModelStore() == nil {
		req.GetBackup().ModelStore = make(map[string]string)
	}

	if req.GetBackup().GetActiveJobs() == nil {
		req.GetBackup().ActiveJobs = make([]*api.Job, 0)
	}

	if req.GetBackup().GetCompletedJobs() == nil {
		req.GetBackup().CompletedJobs = make([]*api.Job, 0)
	}

	if req.GetBackup().GetPendingJobs() == nil {
		req.GetBackup().PendingJobs = make([]*api.Job, 0)
	}

	ic.ModelStore = NewModelStore()
	ic.Scheduler.ActiveJobs = make(map[string]*api.Job)
	ic.Scheduler.CompletedJobs = make(map[string]*api.Job)
	ic.Scheduler.PendingJobs = utils.NewQueue[*api.Job]()

	for k, v := range req.GetBackup().GetModelStore() {
		ic.ModelStore.AddModel(utils.ModelType(k), utils.DatasetType(v))
	}

	for _, job := range req.GetBackup().GetActiveJobs() {
		ic.Scheduler.ActiveJobs[job.GetId()] = job
	}

	for _, job := range req.GetBackup().GetCompletedJobs() {
		ic.Scheduler.CompletedJobs[job.GetId()] = job
	}

	for _, job := range req.GetBackup().GetPendingJobs() {
		ic.Scheduler.PendingJobs.Push(job)
	}

	return &api.BackupResponse{}, nil
}

func (ic *IDunnoCoordinator) IDunnoStatus(ctx context.Context, req *api.IDunnoStatusRequest) (*api.IDunnoStatusResponse, error) {
	switch req.GetWhich() {
	case STATUS_WORKER:
		return &api.IDunnoStatusResponse{Message: ic.PrintWorkers()}, nil
	case STATUS_JSON_WORKER:
		return &api.IDunnoStatusResponse{Message: ic.PrintWorkersJSON()}, nil
	case STATUS_JOBS:
		return &api.IDunnoStatusResponse{Message: ic.PrintJobs()}, nil
	case STATUS_JSON_JOBS:
		return &api.IDunnoStatusResponse{Message: ic.PrintJobsJSON()}, nil
	case STATUS_JOB_ID:
		return &api.IDunnoStatusResponse{Message: ic.PrintJob(req.GetPayload())}, nil
	case STATUS_JOB_ID_STATS:
		return &api.IDunnoStatusResponse{Message: ic.PrintJobStatistics(req.GetPayload())}, nil
	case STATUS_JSON_JOB_ID:
		return &api.IDunnoStatusResponse{Message: ic.PrintJobJSON(req.GetPayload())}, nil
	case STATUS_COMPLETED_JOBS:
		return &api.IDunnoStatusResponse{Message: ic.PrintCompletedJobs()}, nil
	case STATUS_JSON_COMPLETED_JOBS:
		return &api.IDunnoStatusResponse{Message: ic.PrintCompletedJobsJSON()}, nil
	}

	return nil, fmt.Errorf("invalid status request")
}
