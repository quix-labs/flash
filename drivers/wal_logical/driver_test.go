package wal_logical

import (
	"github.com/quix-labs/flash"
	"github.com/testcontainers/testcontainers-go"
	"testing"
)

func WithWalLogical() testcontainers.CustomizeRequestOption {
	return func(req *testcontainers.GenericContainerRequest) error {
		req.Cmd = append(req.Cmd, "-c", "wal_level=logical")
		return nil
	}
}
func TestDriver(t *testing.T) {
	driverConfig := flash.DefaultDriverTestConfig
	driverConfig.ContainerCustomizers = append(driverConfig.ContainerCustomizers, WithWalLogical())
	flash.RunFlashDriverTestCase(t, driverConfig, func() *Driver {
		return NewDriver(&DriverConfig{})
	})
}
