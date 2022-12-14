package main

import (
	"context"
	"errors"
	"fmt"
	"mp4/api"
	"mp4/sdfs"
	"mp4/utils"

	"google.golang.org/grpc"
)

type IDunnoClientCLI interface {
	TrainModel(modelType string, dataset string) error
	ServeModel(modelType string, batchSize int) error
}

func (ic *IDunnoClient) TrainModel(modelType string, dataset string) error {
	if !IsValidModelType(modelType) {
		fmt.Printf("Invalid model type: %s\n", modelType)
		fmt.Printf("Supported model types are: %v\n", utils.SupportedModelTypes)
		return errors.New("invalid model type")
	}

	if !IsValidDatasetType(dataset) {
		fmt.Printf("Invalid dataset type: %s\n", dataset)
		fmt.Printf("Supported dataset types are: %v\n", utils.SupportedDatasetTypes)
		return errors.New("invalid dataset type")
	}

	err := ic.SDFSClient.List(dataset)
	if err != nil {
		fmt.Printf("Error retrieving dataset %s: %s\n", dataset, err.Error())
		return err
	}

	fmt.Printf("Training model %s on dataset %s...\n", modelType, dataset)

	// creating gRPC Coordinator client
	coordinatorAddr, err := ic.Ring.LookupLeader()
	if err != nil {
		fmt.Printf("Error retrieving coordinator: %s\n", err.Error())
		return err
	}

	// dial to coordinator
	conn, err := grpc.Dial(coordinatorAddr, sdfs.GRPC_OPTIONS...)
	if err != nil {
		fmt.Printf("Failed to dial coordinator: " + coordinatorAddr)
		return err
	}
	defer conn.Close()

	// send Train gRPC to coordinator
	client := api.NewCoordinatorServiceClient(conn)
	_, err = client.Train(context.Background(), &api.TrainRequest{
		TrainTask: &api.TrainTask{
			Model:   modelType,
			Dataset: dataset,
		},
	})
	if err != nil {
		fmt.Printf("Error sending train request to coordinator: %s\n", err.Error())
		return err
	}

	fmt.Printf("Successfully sent train request to coordinator\n")
	return nil
}

func (ic *IDunnoClient) ServeModel(modelType string, batchSize int) error {
	if !IsValidModelType(modelType) {
		fmt.Printf("Invalid model type: %s\n", modelType)
		fmt.Printf("Supported model types are: %v\n", utils.SupportedModelTypes)
		return errors.New("invalid model type")
	}

	fmt.Printf("Serving model %s with batch size %v...\n", modelType, batchSize)

	// creating gRPC Coordinator client
	coordinatorAddr, err := ic.Ring.LookupLeader()
	if err != nil {
		fmt.Printf("Error retrieving coordinator: %s\n", err.Error())
		return err
	}

	// dial to coordinator
	conn, err := grpc.Dial(coordinatorAddr, sdfs.GRPC_OPTIONS...)
	if err != nil {
		fmt.Printf("Failed to dial coordinator: " + coordinatorAddr)
		return err
	}
	defer conn.Close()

	// send Train gRPC to coordinator
	client := api.NewCoordinatorServiceClient(conn)
	_, err = client.Inference(context.Background(), &api.InferenceRequest{
		InferenceTask: &api.InferenceTask{
			Model:     modelType,
			BatchSize: int32(batchSize),
		},
	})
	if err != nil {
		fmt.Printf("Error sending train request to coordinator: %s\n", err.Error())
		return err
	}

	fmt.Printf("Successfully uploaded serving job to Idunno\n")
	return nil
}

func (ic *IDunnoClient) GetRealTimeStatus(which string, payload string) error {
	// creating gRPC Coordinator client
	coordinatorAddr, err := ic.Ring.LookupLeader()
	if err != nil {
		fmt.Printf("Error retrieving coordinator: %s\n", err.Error())
		return err
	}

	// dial to coordinator
	conn, err := grpc.Dial(coordinatorAddr, sdfs.GRPC_OPTIONS...)
	if err != nil {
		fmt.Printf("Failed to dial coordinator: " + coordinatorAddr)
		return err
	}
	defer conn.Close()

	client := api.NewCoordinatorServiceClient(conn)

	res, err := client.IDunnoStatus(context.Background(), &api.IDunnoStatusRequest{
		Which:   which,
		Payload: payload,
	})
	if err != nil {
		fmt.Println("Unable to receive real-time update from coordinator.")
		return err
	}

	fmt.Printf("\n%v\n", res.Message)
	return nil
}
