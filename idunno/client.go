package main

import (
	"errors"
	"fmt"
	"math"
	"mp4/ring"
	"mp4/sdfs"
	"mp4/utils"
	"strconv"
	"strings"
)

type IDunnoClient struct {
	// ring failure detector
	Ring *ring.RingServer
	// sdfs client
	SDFSClient *sdfs.SDFSClient
	IDunnoClientCLI
}

func NewIDunnoClient(ring *ring.RingServer, sdfsClient *sdfs.SDFSClient) *IDunnoClient {
	return &IDunnoClient{
		Ring:       ring,
		SDFSClient: sdfsClient,
	}
}

func (ic *IDunnoClient) ExecuteCommand(command string) error {
	args := strings.Split(command, " ")

	switch args[0] {
	case "train":
		if len(args) != 3 {
			fmt.Println("format: train model dataset")
			return errors.New("invalid arguments")
		}
		model, dataset := args[1], args[2]
		return ic.TrainModel(model, dataset)

	case "serve":
		if len(args) != 3 {
			fmt.Println("format: serve model batch_size")
			return errors.New("invalid arguments")
		}
		model, batchSize := args[1], args[2]
		size, err := strconv.Atoi(batchSize)
		if err != nil {
			fmt.Println("batch_size must be an integer")
			return errors.New("invalid arguments")
		}
		return ic.ServeModel(model, size)

	case "idunno-worker", "iw", "w":
		if len(args) != 1 {
			fmt.Println("format: idunno-worker")
			return errors.New("invalid arguments")
		}
		return ic.GetRealTimeStatus(STATUS_WORKER, "")

	case "idunno-jobs", "j":
		if len(args) != 1 {
			fmt.Println("format: idunno-jobs")
			return errors.New("invalid arguments")
		}
		return ic.GetRealTimeStatus(STATUS_JOBS, "")

	case "idunno-job", "ij":
		if len(args) != 2 {
			fmt.Println("format: idunno-job job_id")
			return errors.New("invalid arguments")
		}
		return ic.GetRealTimeStatus(STATUS_JOB_ID, args[1])

	case "idunno-completed-jobs", "cj":
		if len(args) != 1 {
			fmt.Println("format: idunno-completed-jobs")
			return errors.New("invalid arguments")
		}
		return ic.GetRealTimeStatus(STATUS_COMPLETED_JOBS, "")

	case "idunno-job-stats", "ijs":
		if len(args) != 2 {
			fmt.Println("format: idunno-job-stats job_id")
			return errors.New("invalid arguments")
		}
		return ic.GetRealTimeStatus(STATUS_JOB_ID_STATS, args[1])

	case "qps":
		if len(args) != 2 || !(args[1] == "local" || args[1] == "global") {
			fmt.Println("format: qps [local|global]")
			return errors.New("invalid arguments")
		}
		if args[1] == "local" {
			utils.SetLastSeconds(10)
		} else {
			utils.SetLastSeconds(math.MaxFloat64)
		}
		return nil

	default:
		return errors.New("invalid command")
	}
}
