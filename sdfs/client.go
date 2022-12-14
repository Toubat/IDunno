package sdfs

import (
	"fmt"
	"math"
	"mp4/utils"

	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const CHUNK_SIZE int64 = 50 * utils.MegaByte
const INTERVAL = 300 * time.Millisecond
const MAX_BUFFER_SIZE = 100 * utils.MegaByte
const REPLICA_COUNT = 4

var GRPC_OPTIONS = []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(MAX_BUFFER_SIZE), grpc.MaxCallSendMsgSize(MAX_BUFFER_SIZE))}

// Timeout
const GET_TIMEOUT = 12 * time.Second
const PUT_TIMEOUT = 12 * time.Second
const DELETE_TIMEOUT = 2 * time.Second
const LOOKUP_TIMEOUT = 12 * time.Second

// Consistency Level
const READ_CONSISTENCY = 2
const WRITE_CONSISTENCY = 3
const DELETE_CONSISTENCY = REPLICA_COUNT
const LOOKUP_CONSISTENCY = REPLICA_COUNT

type SDFSClient struct {
	SDFSServer *SDFSServer
	Printf     func(format string, a ...any) (n int, err error)
	Println    func(a ...any) (n int, err error)
	SDFSClientCLI
	SDFSClientAction
	SDFSClientFS
}

func NewSDFSClient(sdfs *SDFSServer) *SDFSClient {
	client := &SDFSClient{
		SDFSServer: sdfs,
	}
	client.EnableLogs(true)
	return client
}

func (c *SDFSClient) EnableLogs(flag bool) {
	if flag {
		c.Printf = fmt.Printf
		c.Println = fmt.Println
	} else {
		c.Printf = func(format string, a ...any) (n int, err error) { return }
		c.Println = func(a ...any) (n int, err error) { return }
	}
}

// Get hash ring size-awared consistency level
func (c *SDFSClient) GetConsistencyLevel(reqType SDFSTaskType) int {
	var level int

	switch reqType {
	case SDFS_GET:
		level = READ_CONSISTENCY
	case SDFS_PUT:
		level = WRITE_CONSISTENCY
	case SDFS_DELETE:
		level = DELETE_CONSISTENCY
	case SDFS_LIST:
		level = LOOKUP_CONSISTENCY
	default:
		level = 0
	}

	return int(math.Min(float64(level), float64(c.SDFSServer.HashRing.Len())))
}

// Get timeout for each request type
func (c *SDFSClient) GetTimeout(reqType SDFSTaskType) time.Duration {
	var timeout time.Duration

	switch reqType {
	case SDFS_GET:
		timeout = GET_TIMEOUT
	case SDFS_PUT:
		timeout = PUT_TIMEOUT
	case SDFS_DELETE:
		timeout = DELETE_TIMEOUT
	case SDFS_LIST:
		timeout = LOOKUP_TIMEOUT
	default:
		timeout = 3 * time.Second
	}

	return timeout
}

func (c *SDFSClient) HandleTaskFailure(task SDFSTask, err error) {
	c.Printf("\nError executing task: %v %v\n", task.GetType(), task.GetSDFSFile())
	c.Printf("Error message: %v\n", err)
}

func (c *SDFSClient) CalculateTime(task SDFSTask, t time.Time) {
	c.Printf("Total time taken for %s %v: %v\n", task.GetType(), task.GetSDFSFile(), time.Duration(time.Since(t).Nanoseconds()))
}
