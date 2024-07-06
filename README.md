# Flash

Flash is a lightweight Go library for managing real-time PostgreSQL changes using event management.

The package automatically creates triggers at runtime, listens for them, and broadcasts changes to your applications.

Currently, it uses PostgreSQL's `pg_notify` system under the hood.

## Features

- Efficient event management.
- Dynamic creation and deletion of PostgreSQL triggers.
- Supports common PostgreSQL events: Insert, Update, Delete, Truncate.

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
	"os"
	"os/signal"

	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	// Example with listener and client setup
	postsListenerConfig := &types.ListenerConfig{Table: "public.posts"}
	postsListener := listeners.NewListener(postsListenerConfig)
	postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %+v\n", event)
	})

	// Create client
	clientConfig := &types.ClientConfig{DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb"}
	flashClient := client.NewClient(clientConfig)
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
- [Trigger all events on table](examples/trigger_all/trigger_all.go)
- [Listen for specific fields](examples/specific_fields/specific_fields.go)

## Drivers
You can see all drivers details [here](pkg/drivers/DRIVERS.md)

## DX - Features / Planned Features

The following features are planned for future implementation:

- [x] Driver interfaces for creating new drivers.
- [x] Listen for changes in specific columns, not the entire row.
- [ ] Remove client in favor of direct listener start
- [ ] Support attaching/detaching new listener during runtime.
- [ ] Soft-delete support: receive delete events when SQL condition is respected. Example: `deleted_at IS NOT NULL`.
- [ ] More performant driver. See [DRIVERS.md](pkg/drivers/DRIVERS.md)
- [ ] Tests implementation
- ... any feedback is welcome.

## Additional Details

You can find a temporary workflow graph [here](WORKFLOW.md).


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


