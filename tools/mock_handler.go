package tools

import (
	"net"
	"net/http"
	"testing"

	env "github.com/caarlos0/env/v8"
	"github.com/go-faster/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc"
)

type MocksHandler struct {
	ctl             *gomock.Controller
	grpcMockHandler *GRPCMockHandler
	httpMockHandler *HTTPMockHandler
}

type MocksConfig struct {
	GrpcMockAddress  string `env:"GRPC_MOCK_ADDRESS" envDefault:"127.0.0.1:9191"`
	HTTPMockAddress  string `env:"HTTP_MOCK_ADDRESS" envDefault:"127.0.0.1:9898"`
	GrpcListenSchema string `env:"GRPC_LISTEN_SCHEMA" envDefault:"tcp"`
	HTTPListenSchema string `env:"HTTP_LISTEN_SCHEMA" envDefault:"tcp"`
}

type GrpcCB func(server *grpc.Server, ctl *gomock.Controller)
type HTTPCB func(server *http.ServeMux, ctl *gomock.Controller)

// NewMocksHandler creates a new MocksHandler with HTTP and gRPC mock servers.
func NewMocksHandler(t *testing.T, gCb GrpcCB, hCb HTTPCB) *MocksHandler {
	cfg := &MocksConfig{}
	opts := env.Options{
		Prefix: "GOAT_",
	}
	err := env.ParseWithOptions(cfg, opts)
	require.NoError(t, err, "failed to parse mocks config")

	h := &MocksHandler{
		ctl: gomock.NewController(t),
	}

	// Only create gRPC mock handler if callback is provided
	if gCb != nil {
		h.grpcMockHandler, err = NewGRPCMockHandler(cfg.GrpcListenSchema, cfg.GrpcMockAddress, func(server *grpc.Server) {
			gCb(server, h.ctl)
		})
		require.NoError(t, err, "failed to create gRPC mock handler")
	}

	// Only create HTTP mock handler if callback is provided
	if hCb != nil {
		h.httpMockHandler, err = NewHTTPMockHandler(cfg.HTTPListenSchema, cfg.HTTPMockAddress, func(server *http.ServeMux) {
			hCb(server, h.ctl)
		})
		require.NoError(t, err, "failed to create HTTP mock handler")
	}

	return h
}

func (m *MocksHandler) Start(t *testing.T) {
	if m.grpcMockHandler != nil {
		go func() {
			if err := m.grpcMockHandler.Start(); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Error(err)
			}
		}()
	}
	if m.httpMockHandler != nil {
		go func() {
			if err := m.httpMockHandler.Start(); err != nil && !errors.Is(err, net.ErrClosed) {
				t.Error(err)
			}
		}()
	}
}

func (m *MocksHandler) Stop() {
	m.ctl.Finish()
	if m.grpcMockHandler != nil {
		_ = m.grpcMockHandler.Stop() //nolint:errcheck 
	}
	if m.httpMockHandler != nil {
		_ = m.httpMockHandler.Stop() //nolint:errcheck
	}
}
