package sdfs

import (
	"context"
	"errors"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"strings"

	"strconv"
	"time"

	"google.golang.org/grpc"
)

type SDFSClientAction interface {
	ExecuteTask(task SDFSTask) (SDFSTaskResult, error)
	ExecuteCommand(command string) error
	HandleTaskResults(reqType SDFSTaskType, results []SDFSTaskResult) SDFSTaskResult
	FetchSequence() (*api.FetchSequenceResponse, error)
	RouteTask(task SDFSTask, seq *api.Sequence, replica *api.Process) (SDFSTaskResult, error)
}

func (c *SDFSClient) ExecuteTask(task SDFSTask) (SDFSTaskResult, error) {
	// Fetch sequence from leader
	res, err := c.FetchSequence(task)
	if err != nil {
		return nil, err
	}
	if res.GetStatus() == api.ResponseStatus_NOT_CONVERGED {
		return nil, nil
	}

	// Find replica set
	replicas := c.SDFSServer.HashRing.FindReplicas(task.GetSDFSFile(), REPLICA_COUNT)

	signal := make(chan SDFSTaskResult)
	for _, r := range replicas {
		go func(replica *api.Process) {
			res, err := c.RouteTask(task, res.GetSeq(), replica)
			if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
				logger.Error(fmt.Sprintf("Failed to %s file %v from %v: %v", task.GetType(), task.GetSDFSFile(), replica.Address(), err))
				return
			}
			signal <- res
		}(r)
	}

	// waiting for all acks in a certain consistency level
	ackRequired := c.GetConsistencyLevel(task.GetType())
	ackResults := make([]SDFSTaskResult, 0)
	ticker := time.NewTicker(c.GetTimeout(task.GetType()))

	// logger.Info("Waiting for " + strconv.Itoa(ackRequired) + " acks...")
	for {
		// check if enough acks received
		if len(ackResults) >= ackRequired {
			// logger.Info("Received enough acks for " + string(task.GetType()) + " request")
			return c.HandleTaskResults(task.GetType(), ackResults)
		}

		select {
		case res := <-signal:
			// logger.Info(fmt.Sprintf("Received ack %v", res.GetStatus()))
			// collect ack messages
			ackResults = append(ackResults, res)
		case <-ticker.C:
			// timeout waiting for ack
			logger.Error("Timeout waiting for acks")

			switch task.GetType() {
			case SDFS_GET:
				return SDFSGetTaskResult{Status: api.ResponseStatus_NOT_FOUND}, nil
			case SDFS_LIST:
				return SDFSListTaskResults{Status: api.ResponseStatus_NOT_FOUND}, nil
			default:
				return nil, errors.New("timeout waiting for acks")
			}
		}
	}
}

func (c *SDFSClient) HandleTaskResults(reqType SDFSTaskType, results []SDFSTaskResult) (SDFSTaskResult, error) {
	if len(results) == 0 {
		return nil, errors.New("no results to handle")
	}

	switch reqType {
	case SDFS_GET:
		latest := results[0].(SDFSGetTaskResult).Seq
		data := results[0].(SDFSGetTaskResult).Data

		for _, res := range results {
			if latest.Less(res.(SDFSGetTaskResult).Seq) {
				latest = res.(SDFSGetTaskResult).Seq
				data = res.(SDFSGetTaskResult).Data
			}
		}
		return SDFSGetTaskResult{Status: api.ResponseStatus_OK, Seq: latest, Data: data}, nil

	case SDFS_PUT:
		return SDFSPutTaskResult{Status: api.ResponseStatus_OK}, nil

	case SDFS_DELETE:
		return SDFSDeleteTaskResult{Status: api.ResponseStatus_OK}, nil

	case SDFS_LIST:
		lookups := make([]SDFSListTaskResult, 0)

		for _, res := range results {
			lookups = append(lookups, res.(SDFSListTaskResult))
		}
		return SDFSListTaskResults{Status: api.ResponseStatus_OK, Results: lookups}, nil

	default:
		return nil, errors.New("unknown request type")
	}
}

func (c *SDFSClient) RouteTask(task SDFSTask, seq *api.Sequence, replica *api.Process) (SDFSTaskResult, error) {
	// Make grpc connection
	conn, err := grpc.Dial(replica.Address(), GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to dial " + replica.Address() + ": " + err.Error())
		return nil, err
	}
	defer conn.Close()

	// Create grpc client
	client := api.NewSDFSServiceClient(conn)

	switch task.GetType() {
	case SDFS_GET:
		res, err := client.Read(context.Background(), &api.ReadRequest{
			Filename:      task.GetSDFSFile(),
			Version:       task.(SDFSGetTask).Version,
			LocalFilename: task.(SDFSGetTask).LocalFile,
			Seq:           seq,
		})
		if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
			return nil, err
		}
		return SDFSGetTaskResult{
			Status: res.GetStatus(),
			Seq:    res.GetSeq(),
			Data:   res.GetData(),
		}, nil

	case SDFS_PUT:
		res, err := client.Write(context.Background(), &api.WriteRequest{
			Filename: task.GetSDFSFile(),
			Data:     task.(SDFSPutTask).Data,
			WriteId:  task.(SDFSPutTask).WriteId,
			Seq:      seq,
		})
		if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
			return nil, err
		}
		return SDFSPutTaskResult{
			Status: res.GetStatus(),
		}, nil

	case SDFS_DELETE:
		res, err := client.Delete(context.Background(), &api.DeleteRequest{
			Filename: task.GetSDFSFile(),
			Seq:      seq,
		})
		if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
			return nil, err
		}
		return SDFSDeleteTaskResult{
			Status: res.GetStatus(),
		}, nil

	case SDFS_LIST:
		res, err := client.Lookup(context.Background(), &api.LookupRequest{
			Filename: task.GetSDFSFile(),
			Seq:      seq,
		})
		if err != nil || res.GetStatus() == api.ResponseStatus_ERROR {
			return nil, err
		}
		return SDFSListTaskResult{
			Status: res.GetStatus(),
			Ip:     res.GetIp(),
			Port:   res.GetPort(),
		}, nil

	default:
		return nil, errors.New("invalid request type")
	}
}

func (c *SDFSClient) FetchSequence(task SDFSTask) (*api.FetchSequenceResponse, error) {
	leaderAddr, err := c.SDFSServer.Ring.LookupLeader()
	if err != nil {
		return nil, err
	}

	// Dial to leader
	conn, err := grpc.Dial(leaderAddr, GRPC_OPTIONS...)
	if err != nil {
		logger.Error("Failed to dial leader while trying to to send request " + task.GetSDFSFile())
		return nil, err
	}
	defer conn.Close()

	// Fetch sequence from leader
	client := api.NewSDFSServiceClient(conn)
	res, err := client.FetchSequence(context.Background(), &api.FetchSequenceRequest{})
	if err != nil {
		logger.Error("Failed to fetch sequence while trying to send request " + task.GetSDFSFile())
		return nil, err
	}

	return res, nil
}

func (c *SDFSClient) ExecuteCommand(command string) error {
	args := strings.Split(command, " ")

	switch args[0] {
	case "get":
		if len(args) != 3 {
			fmt.Println("format: get sdfsfilename localfilename")
			return errors.New("invalid arguments")
		}
		sdfsFile, localFile := args[1], args[2]
		return c.Get(localFile, sdfsFile, LATEST_VERSION)

	case "put":
		if len(args) != 3 {
			fmt.Println("format: put localfilename sdfsfilename")
			return errors.New("invalid arguments")
		}
		localFile, sdfsFile := args[1], args[2]
		return c.Put(localFile, sdfsFile)

	case "delete":
		if len(args) != 2 {
			fmt.Println("format: delete sdfsfilename")
			return errors.New("invalid arguments")
		}
		sdfsFile := args[1]
		return c.Delete(sdfsFile)

	case "ls":
		if len(args) != 2 {
			fmt.Println("format: ls sdfsfilename")
			return errors.New("invalid arguments")
		}
		sdfsFile := args[1]
		return c.List(sdfsFile)

	case "store":
		if len(args) != 1 {
			fmt.Println("format: store")
			return errors.New("invalid arguments")
		}
		return c.Store()

	case "get-versions":
		if len(args) < 4 {
			fmt.Println("format: get-versions sdfsfilename num-versions localfilename")
			return errors.New("invalid arguments")
		}
		sdfsFile, localFile := args[1], args[3]
		numVersions, err := strconv.Atoi(args[2])
		if err != nil {
			fmt.Println("format: get-versions sdfsfilename num-versions localfilename")
			return errors.New("invalid arguments")
		}
		return c.GetVersions(localFile, sdfsFile, numVersions)

	case "putdir":
		if len(args) != 3 {
			fmt.Println("format: putdir localdirname sdfsdirname")
			return errors.New("invalid arguments")
		}
		localDir, sdfsDir := args[1], args[2]
		return c.PutDir(localDir, sdfsDir)

	case "valdir":
		if len(args) != 2 {
			fmt.Println("format: valdir sdfsdirname")
			return errors.New("invalid arguments")
		}
		sdfsDir := args[1]
		return c.ValidateDir(sdfsDir)

	case "deldir":
		if len(args) != 2 {
			fmt.Println("format: deldir sdfsdirname")
			return errors.New("invalid arguments")
		}

		sdfsDir := args[1]
		return c.DeleteDir(sdfsDir)

	case "enable-log":
		if len(args) != 1 {
			fmt.Println("format: enable-log")
			return errors.New("invalid arguments")
		}
		c.EnableLogs(true)
		return nil

	case "disable-log":
		if len(args) != 1 {
			fmt.Println("format: disable-log")
			return errors.New("invalid arguments")
		}
		c.EnableLogs(false)
		return nil
	}

	return errors.New("invalid command")
}
