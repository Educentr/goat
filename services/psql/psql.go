package psql

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/docker/go-connections/nat"
	_ "github.com/lib/pq" // PostgreSQL driver
	testcontainers "github.com/testcontainers/testcontainers-go"
	postgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	common "github.com/Educentr/goat/services/common"
)

const (
	defaultDBUser = "app"
	defaultDBPass = "app"
	defaultDBName = "app"

	userNameEnvKey = "POSTGRES_USER"
	userPassEnvKey = "POSTGRES_PASSWORD" //nolint:gosec // G101: environment variable name, not a credential
	dbNameEnvKey   = "POSTGRES_DB"

	startTimeout = 60 * time.Second
)

type (
	Env struct {
		testcontainers.Container
		URI    string
		DBName string
		DBUser string
		DBPass string
		DBPort string
		DBHost string
	}
)

var (
	defaultImage = common.DockerProxy("postgres:15.3-alpine3.18")
)

func (e *Env) SQL() (*sql.DB, error) { return sql.Open("postgres", e.URI) }

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			ImageSubstitutors: common.ImageSubstitutors(),
		},
	}

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck 
	}

	if req.Image == "" {
		req.Image = defaultImage
	}

	var env Env
	if v, ok := req.Env[userNameEnvKey]; ok {
		env.DBUser = v
	} else {
		opts = append(opts, postgres.WithUsername(defaultDBUser))
		env.DBUser = defaultDBUser
	}

	if v, ok := req.Env[userPassEnvKey]; ok {
		env.DBPass = v
	} else {
		opts = append(opts, postgres.WithPassword(defaultDBPass))
		env.DBPass = defaultDBPass
	}

	if v, ok := req.Env[dbNameEnvKey]; ok {
		env.DBName = v
	} else {
		opts = append(opts, postgres.WithDatabase(defaultDBName))
		env.DBName = defaultDBName
	}

	if req.WaitingFor == nil {
		opts = append(opts, testcontainers.WithWaitStrategy(
			wait.ForSQL("5432/tcp", "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf(
					"postgres://%s:%s@%s/%s?sslmode=disable",
					defaultDBUser,
					defaultDBPass,
					net.JoinHostPort(host, port.Port()),
					defaultDBName,
				)
			}).
				WithStartupTimeout(startTimeout),
		))
	}

	p, err := postgres.RunContainer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	env.URI, err = p.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}
	env.URI += " sslmode=disable"
	env.Container = p

	port, err := p.Container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		return nil, err
	}
	env.DBPort = port.Port()

	host, err := p.Container.Host(ctx)
	if err != nil {
		return nil, err
	}
	env.DBHost = host
	return &env, nil
}
