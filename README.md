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
	"github.com/quix-labs/flash/client"
)

func main() {
	postsListener := client.NewListener(&client.ListenerConfig{Table: "posts"})
	stop, _ := postsListener.On(client.EventUpdate^client.EventDelete, func(event client.Event) {
		fmt.Println("Event received All" + string(event))
	})
	defer stop() // Or ignore explicitly if you want to keep running during all application lifetime

	flashClient := client.NewClient(&client.Config{
		DatabaseCnx: "postgresql://user:pass@localhost:5432/db",
	})
	flashClient.AddListener(postsListener)
	go flashClient.Start()
	defer flashClient.Close()
	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
```

## Planned Features

The following features are planned for future implementation:

- Listen for changes in specific columns, not the entire table.
- Soft-delete support: receive delete events when SQL condition is respected. Example: `deleted_at IS NOT NULL`.
- Driver interfaces for creating new drivers.
- New driver that uses logical replication.
- ... any feedback is welcome.


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


