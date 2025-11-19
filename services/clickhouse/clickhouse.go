package clickhouse

import (
	"context"
	"net/http"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	testcontainers "github.com/testcontainers/testcontainers-go"
	clickhouse "github.com/testcontainers/testcontainers-go/modules/clickhouse"
	wait "github.com/testcontainers/testcontainers-go/wait"

	common "github.com/Educentr/goat/services/common"
)

const (
	defaultDBName = "default"
	defaultDBUser = "default"
	defaultDBPass = ""

	envDB   = "CLICKHOUSE_DB"
	envUser = "CLICKHOUSE_USER"
	envPass = "CLICKHOUSE_PASSWORD"
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
	defaultImage = common.DockerProxy("clickhouse/clickhouse-server:23")
)

func (e *Env) Conn() (ch.Conn, error) { //nolint:ireturn
	return ch.Open(&ch.Options{
		Addr: []string{e.URI},
		Auth: ch.Auth{
			Database: e.DBName,
			Username: e.DBUser,
			Password: e.DBPass,
		},
	})
}

func Run(ctx context.Context, opts ...testcontainers.ContainerCustomizer) (*Env, error) {
	var (
		req = testcontainers.GenericContainerRequest{
			ContainerRequest: testcontainers.ContainerRequest{
				ImageSubstitutors: common.ImageSubstitutors(),
				Env:               map[string]string{},
			},
		}

		env Env
		ok  bool
	)

	for _, e := range opts {
		_ = e.Customize(&req) //nolint:errcheck
	}

	if req.Image == "" {
		req.Image = defaultImage
	}

	if env.DBUser, ok = req.Env[envUser]; !ok {
		env.DBUser = defaultDBUser
		opts = append(opts, clickhouse.WithUsername(env.DBUser))
	}

	if env.DBPass, ok = req.Env[envPass]; !ok {
		env.DBPass = defaultDBPass
		opts = append(opts, clickhouse.WithPassword(env.DBPass))
	}

	if env.DBName, ok = req.Env[envDB]; !ok {
		env.DBName = defaultDBName
		opts = append(opts, clickhouse.WithDatabase(env.DBName))
	}

	if req.WaitingFor == nil {
		opts = append(opts, testcontainers.WithWaitStrategy(wait.ForAll(
			wait.ForHTTP("/ping").WithPort("8123/tcp").WithStatusCodeMatcher(
				func(status int) bool {
					return status == http.StatusOK
				},
			),
		)))
	}

	p, err := clickhouse.RunContainer(ctx, opts...)
	if err != nil {
		return nil, err
	}

	env.URI, err = p.ConnectionString(ctx)
	if err != nil {
		return nil, err
	}

	port, err := p.MappedPort(ctx, "9000/tcp")
	if err != nil {
		return nil, err
	}

	env.DBPort = port.Port()

	host, err := p.Host(ctx)
	if err != nil {
		return nil, err
	}

	env.DBHost = host
	env.Container = p

	return &env, nil
}
