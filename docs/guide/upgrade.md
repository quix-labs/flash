
# Upgrade from Old Structure

In the previous structure, our project was divided into three distinct sub-modules, making it cumbersome to manage and
integrate changes.

We have now merged these sub-modules into a single, unified
module: [github.com/quix-labs/flash](https://github.com/quix-labs/flash).

This consolidation simplifies the codebase and streamlines development.

### Key Changes:

* **Unified Repository**:

  The previously separate sub-modules are now combined into one repository.
  This allows for easier dependency management and a more cohesive development process.


* **Separate Driver Installation**:

  While the core functionality is now in one place, the drivers need to be installed separately.
  This modular approach ensures that you only include what you need, keeping your projects lightweight.

* **No default driver**:

  By default, we are previously using trigger driver, to keep user informed, the user require now to instanciate the
  driver and pass it in ClientConfig

### Upgrade Guide

* Replace all your `client.NewClient(&type.ClientConfig{})` by `flash.NewClient(&flash.ClientConfig{})`
* Replace all your `listeners.NewListener(types.ListenerConfig{})` by `flash.NewListener(&flash.ListenerConfig{})`
* Instantiate the `Driver` in your codebase and pass it to `flash.ClientConfig{}`

```go
package main

import (
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
)

func main() {
	// Instantiation of driver is now required
	driver := trigger.NewDriver(&trigger.DriverConfig{})
	client := flash.NewClient(&flash.ClientConfig{
		Driver: driver,
	})

	// Instead of listeners.NewListener, use flash.NewListener
	listener := flash.NewListener(&flash.ListenerConfig{})

	// Your additional code here
}
```

## Next steps

Checkout the [Start Listening Guide](./start-listening) to begin.