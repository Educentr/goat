package victoriametrics

import (
	"context"
	"fmt"
	"net"

	errors "github.com/pkg/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"
	wait "github.com/testcontainers/testcontainers-go/wait"

	common "github.com/Educentr/goat/services/common"
)

type (
	Env struct {
		testcontainers.Container
		Address string
	}
)

var (
	defaultImage = common.DockerProxy("victoriametrics/victoria-metrics:v1.103.0")
)

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             defaultImage,
			ImageSubstitutors: common.ImageSubstitutors(),
			ExposedPorts:      []string{"8428/tcp"},
			WaitingFor:        wait.ForLog("starting server at"),
			Cmd: []string{
				"-retentionPeriod=12",
				"-search.cacheTimestampOffset=43200h",
				"-search.latencyOffset=1s",
			},
		},
		Started: true,
	}

	for _, e := range opts {
		err := e.Customize(&req)

		if err != nil {
			return nil, err
		}
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		err = container.Terminate(ctx)
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("Failed to get host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "8428")
	if err != nil {
		err = container.Terminate(ctx)
		if err != nil {
			return nil, err
		}

		return nil, errors.Errorf("Failed to get mapped port: %v", err)
	}

	address := fmt.Sprintf("http://%s", net.JoinHostPort(host, mappedPort.Port()))

	return &Env{
		Container: container,
		Address:   address,
	}, nil
}
