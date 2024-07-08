# Flash

Flash is a lightweight Go library for managing real-time PostgreSQL changes using event management.

## Features

- Start/Stop listening during runtime.
- Supports common PostgreSQL events: Insert, Update, Delete, Truncate.
- Listen changes using WAL replication (using wal driver, default is set to trigger)
- Extendable drivers (for listening changes)

## Notes

**This library is currently under active development.**

Features and APIs may change.

Contributions and feedback are welcome!

## Installation

To install the library, run:

```bash
go get github.com/quix-labs/flash
```

## Usage

Here's a basic example of how to use Flash:

```go
package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/listeners"
	"os"
	"os/signal"

	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	// Example with listener and client setup
	postsListenerConfig := &types.ListenerConfig{Table: "public.posts"}
	postsListener := listeners.NewListener(postsListenerConfig)
	postsListener.On(types.OperationAll, func(event types.Event) {
		switch typedEvent := event.(type) {
		case *types.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *types.UpdateEvent:
			fmt.Printf("update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *types.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		case *types.TruncateEvent:
			fmt.Printf("truncate \n")
		}
	})

	// Create client
	clientConfig := &types.ClientConfig{DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb"}
	flashClient, _ := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)

	// Start listening
	go flashClient.Start()
	defer flashClient.Close()

	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
```

For more detailed examples, check out the following files:

- [Debug queries](examples/debug_trace/debug_trace.go)
- [Trigger insert events on table](examples/trigger_insert/trigger_insert.go)
- [Trigger all events on table](examples/trigger_all/trigger_all.go)
- [Listen for specific fields](examples/specific_fields/specific_fields.go)
- [Parallel Callback](examples/parallel_callback/parallel_callback.go)

## DX - Features / Planned Features

The following features are planned for future implementation:

- ✅ Driver interfaces for creating new drivers.
- ✅ Parallel Callback execution using goroutine
- 🟨 Listen for changes in specific columns, not the entire row. Driver specific see [drivers/README.md](pkg/drivers/README.md)
- ⌛ More performant driver. See [drivers/README.md](pkg/drivers/README.md)
- ⬜ Remove client in favor of direct listener start
- ⬜ Support attaching/detaching new listener during runtime.
- ⬜ Soft-delete support: receive delete events when SQL condition is respected. Example: `deleted_at IS NOT NULL`.
- ⬜ Tests implementation
- ... any feedback is welcome.


## Drivers

You can see all drivers details [here](pkg/drivers/README.md)

## Contributing

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Commit your changes.
4. Push your branch.
5. Create a pull request.

## Credits

- [COLANT Alan](https://github.com/alancolant)
- [All Contributors](../../contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE.md) for more information.


