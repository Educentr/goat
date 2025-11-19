package minio

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go/wait"

	errors "github.com/go-faster/errors"
	minio "github.com/minio/minio-go/v7"
	credentials "github.com/minio/minio-go/v7/pkg/credentials"
	testcontainers "github.com/testcontainers/testcontainers-go"

	common "github.com/Educentr/goat/services/common"
)

type (
	Env struct {
		testcontainers.Container
		EndpointURL     string
		AccessKeyID     string
		SecretAccessKey string
		Region          string
		Token           string
	}
)

var (
	defaultImage = common.DockerProxy("minio/minio")
)

func (env *Env) GetMinioClient() (*minio.Client, error) {
	return minio.New(env.EndpointURL, &minio.Options{
		Creds:  credentials.NewStaticV4(env.AccessKeyID, env.SecretAccessKey, ""),
		Secure: false,
	})
}

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	req := testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:             defaultImage,
			ImageSubstitutors: common.ImageSubstitutors(),
			Cmd:               []string{"server", "/data"},
			ExposedPorts:      []string{"9000/tcp"},
			Env: map[string]string{
				"MINIO_ACCESS_KEY": "minioadmin",
				"MINIO_SECRET_KEY": "minioadmin",
			},
			WaitingFor: wait.ForListeningPort("9000/tcp"),
		},
	}

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx) //nolint:errcheck
		return nil, errors.Errorf("Failed to get host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "9000")
	if err != nil {
		_ = container.Terminate(ctx) //nolint:errcheck
		return nil, errors.Errorf("Failed to get mapped port: %v", err)
	}

	return &Env{
		Container:       container,
		EndpointURL:     fmt.Sprintf("%s:%d", host, mappedPort.Int()),
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		Region:          "us-east-1",
	}, nil
}
