package main

import (
	"mp4/api"
	"sync"
)

type DNSServer struct {
	DNSFile string
	api.DNSServiceServer
	sync.Mutex
}

func NewDNSServer(filename string) *DNSServer {
	return &DNSServer{
		DNSFile: "",
	}
}

func (server *DNSServer) Clear() {
	// clear dns file
	server.DNSFile = ""
}
