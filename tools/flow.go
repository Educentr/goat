package tools

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type Flow struct {
	mocks *MocksHandler
	app   BaseExecutor
	env   *Env
}

func NewFlow(t *testing.T, env *Env, exe BaseExecutor, hcb HTTPCB, gCb GrpcCB) *Flow {
	t.Helper()
	return &Flow{
		env:   env,
		mocks: NewMocksHandler(t, gCb, hcb),
		app:   exe,
	}
}

func (f *Flow) Start(t *testing.T, before, after func(env *Env) error) {
	if before != nil {
		require.NoError(t, before(f.env))
	}

	f.mocks.Start(t)
	require.NoError(t, f.app.Start(), "failed to run app")

	if after != nil {
		require.NoError(t, after(f.env))
	}
}

func (f *Flow) Stop(t *testing.T, before, after func(env *Env) error) {
	if before != nil {
		require.NoError(t, before(f.env))
	}

	f.mocks.Stop()
	require.NoError(t, f.app.Stop(), "failed to stop app")
	_ = f.app.Stop()

	if after != nil {
		require.NoError(t, after(f.env))
	}

	if executor, ok := f.app.(*Executor); ok && executor.fieldsParser != nil {
		f.env.mergeLogFieldStats(executor.fieldsParser.fields, executor.fieldsParser.unmarshalErrors)
	}
}
