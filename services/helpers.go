package services

import testcontainers "github.com/testcontainers/testcontainers-go"

// WithMounts is a helper function to add mounts to a container request.
// This is useful for mounting configuration files or data directories into containers.
//
// Example:
//
//	mounts := testcontainers.ContainerMounts{
//		testcontainers.BindMount("/host/path", "/container/path"),
//	}
//	services.NewBuilder().WithPostgres(services.WithMounts(mounts))
func WithMounts(mounts testcontainers.ContainerMounts) testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Mounts = append(req.Mounts, mounts...)
		return nil
	}
}
