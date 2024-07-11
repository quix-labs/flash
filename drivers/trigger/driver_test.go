package trigger

import (
	"github.com/quix-labs/flash"
	"testing"
)

func TestDriver(t *testing.T) {
	flash.RunFlashDriverTestCase(t, flash.DefaultDriverTestConfig, func() *Driver {
		return NewDriver(&DriverConfig{})
	})
}
