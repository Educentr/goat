package xray

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	errors "github.com/go-faster/errors"
	testcontainers "github.com/testcontainers/testcontainers-go"

	common "github.com/Educentr/goat/services/common"
)

type Env struct {
	testcontainers.Container
	EndpointURL string
}

var (
	defaultImage = common.DockerProxy("teddysun/xray")
)

// WithConfigFile sets a custom config file path for xray.
// This option is REQUIRED - xray will not start without a config file.
func WithConfigFile(configPath string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequestOption(func(req *testcontainers.GenericContainerRequest) error {
		// Add the config file to the container
		req.Files = append(req.Files, testcontainers.ContainerFile{
			HostFilePath:      configPath,
			ContainerFilePath: "/etc/xray/config.json",
			FileMode:          0644,
		})
		return nil
	})
}

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             defaultImage,
			ImageSubstitutors: common.ImageSubstitutors(),
			WaitingFor:        wait.ForListeningPort("443/tcp"),
			ExposedPorts: []string{
				"443/tcp",
			},
		},
	}

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck // options pattern, errors handled during container creation
	}

	// Validate that config file was provided
	hasConfig := false
	for _, file := range req.Files {
		if file.ContainerFilePath == "/etc/xray/config.json" {
			hasConfig = true
			break
		}
	}
	if !hasConfig {
		return nil, errors.New("xray config file is required: use xray.WithConfigFile() to specify the config path")
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on error
		return nil, errors.Errorf("Failed to get host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "443")
	if err != nil {
		_ = container.Terminate(ctx) //nolint:errcheck // best effort cleanup on error
		return nil, errors.Errorf("Failed to get mapped port: %v", err)
	}

	return &Env{
		Container:   container,
		EndpointURL: fmt.Sprintf("%s:%d", host, mappedPort.Int()),
	}, nil
}
