package redis

import (
	"context"
	"fmt"
	"time"

	testcontainers "github.com/testcontainers/testcontainers-go"
	redis "github.com/testcontainers/testcontainers-go/modules/redis"
	wait "github.com/testcontainers/testcontainers-go/wait"

	common "github.com/Educentr/goat/services/common"
)

type (
	Env struct {
		testcontainers.Container
		Address     string
		AddressHost string
		AddressPort string
	}
)

var (
	defaultImage = common.DockerProxy("redis:7.2.2-alpine")
)

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ImageSubstitutors: common.ImageSubstitutors(),
		},
	}

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck 
	}

	var env Env
	if req.Image == "" {
		req.Image = defaultImage
	}

	if req.WaitingFor == nil {
		opts = append(opts, testcontainers.WithWaitStrategy(wait.
			ForExposedPort().
			WithStartupTimeout(30*time.Second),
		))
	}

	container, err := redis.RunContainer(ctx, opts...)
	if err != nil {
		return nil, err
	}
	env.Container = container

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return nil, err
	}

	env.AddressHost = host
	env.AddressPort = port.Port()
	env.Address = fmt.Sprintf("%s:%s", env.AddressHost, env.AddressPort)

	return &env, nil
}
