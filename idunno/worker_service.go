package main

import (
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/utils"
)

func (iw *IDunnoWorker) Train(ctx context.Context, req *api.TrainRequest) (*api.TrainResponse, error) {
	logger.Info(fmt.Sprintf("Worker %v received Train request - Model: %v, Dataset: %v", iw.Ring.Address(), req.GetTrainTask().GetModel(), req.GetTrainTask().GetDataset()))

	// Create runner client
	client, close, err := iw.CreateInferenceClient()
	if err != nil {
		return nil, err
	}
	defer close()

	// Send train request to runner
	logger.Info(fmt.Sprintf("Worker sending Train request to runner: %v", iw.ModelRunner.Address()))
	res, err := client.Train(context.Background(), req)
	if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
		logger.Error(fmt.Sprintf("Error sending Train request to runner: %v", res.GetStatus()))
		return nil, fmt.Errorf("error sending train request to inference service")
	}
	logger.Info(fmt.Sprintf("Worker %v completed training", iw.ModelRunner.Address()))

	return res, nil
}

func (iw *IDunnoWorker) Inference(ctx context.Context, req *api.InferenceRequest) (*api.InferenceResponse, error) {
	iw.Lock()
	defer iw.Unlock()

	task := req.GetInferenceTask()
	logger.Info(fmt.Sprintf("Worker %v received Inference request - Model: %v, BatchSize: %v", iw.Ring.Address(), task.GetModel(), task.GetBatchSize()))

	client, close, err := iw.CreateInferenceClient()
	if err != nil {
		return nil, err
	}
	defer close()

	logger.Info(fmt.Sprintf("Worker %v sending Inference request to runner", iw.ModelRunner.Address()))
	res, err := client.ServeModel(context.Background(), &api.ServeModelRequest{
		Model: task.GetModel(),
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Error sending Inference request to runner: %v", err))
		return nil, fmt.Errorf("error sending inference request to inference service")
	}
	if res.GetStatus() != api.ResponseStatus_OK {
		logger.Error(fmt.Sprintf("Error sending Inference request to runner: %v", res.GetStatus()))
		return nil, fmt.Errorf("error sending inference request to inference service")
	}

	// update inference job & reset previous batch output
	iw.JobId = req.GetJobId()
	iw.BatchOutput = nil

	logger.Info(fmt.Sprintf("Worker %v started inference task", iw.Ring.Address()))
	return &api.InferenceResponse{Status: api.ResponseStatus_OK}, nil
}

func (iw *IDunnoWorker) FinishInference(ctx context.Context, req *api.FinishInferenceRequest) (*api.FinishInferenceResponse, error) {
	iw.BatchOutput = nil
	iw.JobId = utils.EMPTY_STRING
	return &api.FinishInferenceResponse{}, nil
}
