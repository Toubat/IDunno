package main

import (
	"bufio"
	"fmt"
	"mp4/api"
	"mp4/backend"
	"mp4/logger"

	"net"
	"os"
	"strings"

	"google.golang.org/grpc"
)

func main() {
	lis, err := net.Listen("tcp", ":8889")
	if err != nil {
		logger.Error("Failed to listen: " + err.Error())
	}

	grpcServer := grpc.NewServer()
	dnsServer := NewDNSServer("dns.txt")
	defer dnsServer.Clear()

	api.RegisterDNSServiceServer(grpcServer, dnsServer)
	go attachToTerminal(dnsServer)

	// initialize stat server
	statServer := backend.NewIdunnoStatServer(7000)
	go statServer.Serve()

	logger.Init("dns", []string{})
	logger.Info("DNS server started listening on port 8889")
	if err := grpcServer.Serve(lis); err != nil {
		logger.Error("Failed to serve: " + err.Error())
	}
}

/**
 * A function that attach a RingServer to terminal
 * - Read input from stdin
 * - Parse input
 * - Call appropriate function
 */
func attachToTerminal(s *DNSServer) {
	// read input from stdin
	reader := bufio.NewReader(os.Stdin)
	host, _ := os.Hostname()
	fmt.Println("Welcome to the DNS CLI!")
	for {
		fmt.Printf("%v~$ ", host)
		text, _ := reader.ReadString('\n')
		text = text[:len(text)-1]
		args := strings.Fields(text)

		if len(args) == 0 {
			fmt.Println("Invalid command")
			continue
		}

		// ring commands
		switch args[0] {
		case "clear":
			s.Clear()
		case "hostname":
			fmt.Println(s.DNSFile)
		default:
			fmt.Println("Invalid command")
		}
	}
}
