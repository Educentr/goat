package s3

import (
	"context"
	"fmt"

	aws "github.com/aws/aws-sdk-go/aws"
	credentials "github.com/aws/aws-sdk-go/aws/credentials"
	session "github.com/aws/aws-sdk-go/aws/session"
	s3 "github.com/aws/aws-sdk-go/service/s3"
	nat "github.com/docker/go-connections/nat"
	testcontainers "github.com/testcontainers/testcontainers-go"
	tcLocalstack "github.com/testcontainers/testcontainers-go/modules/localstack"

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
	defaultImage = common.DockerProxy("localstack/localstack:1.4.0")
)

func (env *Env) GetS3Client() (*s3.S3, error) {
	awsConfig := &aws.Config{
		Region:                        aws.String(env.Region),
		CredentialsChainVerboseErrors: aws.Bool(true),
		Credentials: credentials.NewStaticCredentials(
			env.AccessKeyID,
			env.SecretAccessKey,
			env.Token,
		),
		S3ForcePathStyle: aws.Bool(true),
		Endpoint:         aws.String("http://" + env.EndpointURL),
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, err
	}

	s3Client := s3.New(sess)
	return s3Client, nil
}

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	opts = append(opts,
		testcontainers.WithImage(defaultImage),
		testcontainers.WithImageSubstitutors(common.NewImageSubstitutor()),
	)

	lsContainer, err := tcLocalstack.RunContainer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	mappedPort, err := lsContainer.MappedPort(ctx, nat.Port("4566/tcp"))
	if err != nil {
		return nil, err
	}

	host, err := lsContainer.Container.Host(ctx)
	if err != nil {
		return nil, err
	}

	return &Env{
		Container:       lsContainer,
		EndpointURL:     fmt.Sprintf("%s:%d", host, mappedPort.Int()),
		AccessKeyID:     "access_key_id",
		SecretAccessKey: "secret_access_key",
		Token:           "token",
		Region:          "us-east-1",
	}, nil
}
