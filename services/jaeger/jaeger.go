package jaeger

import (
	"context"
	"fmt"
	"net"

	"github.com/go-faster/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Educentr/goat/services/common"
)

type (
	Env struct {
		testcontainers.Container
		HTTPCollectorAddress string
		GRPCCollectorAddress string
		Address              string // UI
	}
)

var defaultImage = common.DockerProxy("jaegertracing/all-in-one:1.51")

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (env *Env, err error) {
	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             defaultImage,
			ImageSubstitutors: common.ImageSubstitutors(),
			ExposedPorts: []string{
				"14250/tcp",
				"14268/tcp",
				"14269/tcp",
				"16686/tcp",
				"4317/tcp",
				"4318/tcp",
				"5778/tcp",
				"9411/tcp",
			},
			WaitingFor: wait.ForListeningPort("4317/tcp"),
		},
	}

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck // options pattern, errors handled during container creation
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			err = errors.Join(err, container.Terminate(ctx))
		}
	}()

	host, err := container.Host(ctx)
	if err != nil {
		return nil, errors.Errorf("Failed to get host: %v", err)
	}

	uiPort, err := container.MappedPort(ctx, "16686")
	if err != nil {
		return nil, errors.Errorf("Failed to get UI port: %v", err)
	}

	grpcCollectorPort, err := container.MappedPort(ctx, "4317")
	if err != nil {
		return nil, errors.Errorf("Failed to get gRPC collector port: %v", err)
	}

	httpCollectorPort, err := container.MappedPort(ctx, "4318")
	if err != nil {
		return nil, errors.Errorf("Failed to get HTTP collector port: %v", err)
	}

	jaegerURL := fmt.Sprintf("http://%s:%s", host, uiPort.Port())
	urlCollectorGRPC := net.JoinHostPort(host, grpcCollectorPort.Port())
	urlCollectorHTTP := net.JoinHostPort(host, httpCollectorPort.Port())

	return &Env{
		Container:            container,
		Address:              jaegerURL,
		GRPCCollectorAddress: urlCollectorGRPC,
		HTTPCollectorAddress: urlCollectorHTTP,
	}, nil
}
