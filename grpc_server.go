package goat

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
)

type GRPCMockHandler struct {
	server   *grpc.Server
	listener net.Listener
}

func NewGRPCMockHandler(schema, address string, cb func(server *grpc.Server)) (*GRPCMockHandler, error) {
	h := &GRPCMockHandler{
		server: grpc.NewServer(),
	}
	cb(h.server)
	grpcListen, err := net.Listen(schema, address)
	if err != nil {
		return nil, fmt.Errorf("listen failed: %w", err)
	}
	h.listener = grpcListen
	return h, nil
}

func (h *GRPCMockHandler) Start() error {
	return h.server.Serve(h.listener)
}

func (h *GRPCMockHandler) Stop() error {
	return h.listener.Close()
}
