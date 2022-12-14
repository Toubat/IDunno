package main

import (
	"bufio"
	"context"
	"fmt"
	"mp4/api"
	"mp4/logger"
	"mp4/ring"
	"mp4/sdfs"
	"net"
	"time"

	"os"
	"strconv"
	"strings"

	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
)

var ServerArgs struct {
	Port int `arg:"-p" help:"port number" default:"5000"`
}
var IgnoredLogTypes = []string{logger.PING, logger.UPDATE}

func main() {
	arg.MustParse(&ServerArgs)
	port := ServerArgs.Port

	logger.Init(strconv.Itoa(port), IgnoredLogTypes)

	host, _ := os.Hostname()

	// create a UDP connection for failure detector
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: port,
		IP:   net.ParseIP(host),
	})
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer conn.Close()
	logger.Info("Process is listening on " + host + ":" + strconv.Itoa(port))

	// initialize SDFS server
	sdfsServer := sdfs.NewSDFSServer()
	// initialize a ring failure detector with an additional SDFS related callback
	ringServer := ring.NewRingServer(conn, host, int32(port))
	sdfsServer.Ring = ringServer
	sdfsServer.ClearSDFSFiles()
	InitDataFolder(ringServer.Address())

	// gRPC client initialization
	sdfsClient := sdfs.NewSDFSClient(sdfsServer)
	idunnoClient := NewIDunnoClient(ringServer, sdfsClient)
	sdfsClient.EnableLogs(false)

	// initialize Idunno worker
	worker := NewIDunnoWorker(host, port, ringServer, sdfsClient)
	// initialize Idunno coordinator
	coordinator := NewIDunnoCoordinator(ringServer, sdfsClient)

	ringServer.SetMemberUpdateCallback(func(process *api.Process, action ring.MemAction) {
		sdfsServer.OnMemberUpdate(action, process)
		coordinator.OnMemberUpdate(action, process)
	})

	// initialize gRPC server
	grpcServer := grpc.NewServer(
		grpc.MaxRecvMsgSize(sdfs.MAX_BUFFER_SIZE),
		grpc.MaxSendMsgSize(sdfs.MAX_BUFFER_SIZE),
	)
	api.RegisterSDFSServiceServer(grpcServer, sdfsServer)
	api.RegisterCoordinatorServiceServer(grpcServer, coordinator)
	api.RegisterWorkerServiceServer(grpcServer, worker)

	// corn jobs & listen for incoming requests
	go sdfsServer.Ring.Cron()
	go sdfsServer.Cron()
	go coordinator.Corn()
	go worker.Cron()
	go sdfsServer.Ring.Listen()
	go InitTerminal(sdfsServer, worker, sdfsClient, idunnoClient, grpcServer)

	// serve grpc
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Failed to grpc Server listen: " + err.Error())
	}
	if err := grpcServer.Serve(lis); err != nil {
		fmt.Println("Failed to grpc Server serve: " + err.Error())
	}

	fmt.Println("Server is shutting down...")
}

/**
 * A function that attach a RingServer to terminal
 * - Read input from stdin
 * - Parse input
 * - Call appropriate function
 */
func InitTerminal(ss *sdfs.SDFSServer, worker *IDunnoWorker, sc *sdfs.SDFSClient, ic *IDunnoClient, g *grpc.Server) {
	// read input from stdin
	reader := bufio.NewReader(os.Stdin)
	host, _ := os.Hostname()

	fmt.Println("Welcome to the SDFS CLI!")
	for {
		fmt.Printf("%v~$ ", host)
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		args := strings.Fields(text)

		if len(args) == 0 {
			fmt.Println("Invalid command")
			continue
		}

		// general commands
		switch args[0] {
		case "list_mem", "lm":
			ss.Ring.ListMembers()
		case "list_self", "l":
			ss.Ring.ListSelf()
		case "join":
			ss.Ring.Join()
		case "leave":
			ss.Ring.Leave()
		case "clear-log":
			logger.Init(strconv.Itoa(int(ss.Ring.Port)), IgnoredLogTypes)
		case "stat":
			logger.Stats()
		case "clear-sdfs":
			ss.ClearSDFSFiles()
		case "clear-local":
			sc.ClearFiles()
		case "stop":
			g.Stop()
			return
		}

		// debug command
		switch args[0] {
		case "debug:greet":
			conn, err := grpc.Dial(worker.ModelRunner.Address(), sdfs.GRPC_OPTIONS...)
			if err != nil {
				logger.Error("Failed to dial worker")
				continue
			}
			// Fetch sequence from leader
			client := api.NewInferenceServiceClient(conn)
			res, err := client.Greet(context.Background(), &api.GreetRequest{Name: args[1]})
			if err != nil {
				logger.Error("Failed to greet worker")
				continue
			}
			fmt.Println(res.Message)
			conn.Close()
		case "debug:idunno":
			sc.ExecuteCommand("putdir imagenet imagenet")
			sc.ExecuteCommand("put emotion.txt emotion.txt")
			ic.ExecuteCommand("train resnet50 imagenet")
			ic.ExecuteCommand("train albert emotion.txt")
			ic.ExecuteCommand("serve resnet50 4")
			time.Sleep(3 * time.Second)
			ic.ExecuteCommand("serve albert 8")
		case "debug:train":
			sc.ExecuteCommand("putdir imagenet imagenet")
			sc.ExecuteCommand("put emotion.txt emotion.txt")
			ic.ExecuteCommand("train resnet50 imagenet")
			ic.ExecuteCommand("train albert emotion.txt")
		case "debug:r":
			ic.ExecuteCommand("serve resnet50 2")
		case "debug:a":
			ic.ExecuteCommand("serve albert 8")
		}

		// sdfs client commands
		sc.ExecuteCommand(text)
		// idunno client commands
		ic.ExecuteCommand(text)
	}
}

func InitDataFolder(addr string) {
	// Create data folder if not exist
	if _, err := os.Stat("../data"); os.IsNotExist(err) {
		err = os.Mkdir("../data", 0777)
		if err != nil {
			logger.Error("Failed to create directory data")
			panic(err)
		}
	}

	// Create folder by its address if not exist
	if _, err := os.Stat("../data/" + addr); os.IsNotExist(err) {
		err = os.Mkdir("../data/"+addr, 0777)
		if err != nil {
			logger.Error("Failed to create directory data")
			panic(err)
		}
	}
}
