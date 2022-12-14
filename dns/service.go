package main

import (
	"context"
	"mp4/api"
	"mp4/logger"
)

func (server *DNSServer) Lookup(ctx context.Context, req *api.LookupLeaderRequest) (*api.LookupLeaderResponse, error) {
	server.Lock()
	defer server.Unlock()

	if server.DNSFile != "" {
		logger.Info("Lookup leader address: " + server.DNSFile)
		return &api.LookupLeaderResponse{Address: server.DNSFile}, nil
	}

	logger.Error("Failed to lookup leader address")
	return &api.LookupLeaderResponse{Address: ""}, nil
}

func (server *DNSServer) Update(ctx context.Context, req *api.UpdateLeaderRequest) (*api.UpdateLeaderResponse, error) {
	server.Lock()
	defer server.Unlock()

	address := req.GetLeader().Address()
	server.DNSFile = address

	return &api.UpdateLeaderResponse{
		Status: api.ResponseStatus_OK,
	}, nil
}
