# Flash

Flash is a lightweight Go library for managing real-time PostgreSQL changes using event management.

## Notes

**This library is currently under active development.**

Features and APIs may change.

Contributions and feedback are welcome!

## Table of Contents

* [Features](#features)
* [Installation](#installation)
* [Usage](#usage)
* [Advanced Features](#advanced-features)
    * [1. Configurable Primary Key ⏳](#1-configurable-primary-key-)
    * [2. Custom Conditions ⏳](#2-custom-conditions-)
    * [3. Partial Fields ✅](#3-partial-fields-)
* [Planned Features](#planned-features)
* [Drivers](#drivers)
* [Contributing](#contributing)
* [Credits](#credits)
* [License](#license)

## Features

- ✅ Start/Stop listening during runtime.
- ✅ Supports common PostgreSQL events: Insert, Update, Delete, Truncate.
- ✅ Driver interfaces for creating new drivers.
- ✅ Parallel Callback execution using goroutine
- ✅ Listen for changes in specific columns, not the entire row. (see [Advanced Features](#advanced-features))
- ✅ Listen changes using WAL replication (see [Drivers](#drivers))

## Installation

To install the library, run:

```bash
go get -u github.com/quix-labs/flash@main # Actually main is used for development
go get -u github.com/quix-labs/flash/drivers/trigger@main # Show below to for other drivers

# Write your main package

go mod tidy # Actually needed to download nested dependencies - Working on proper pre-install during previous go get
```

## Usage

Here's a basic example of how to use Flash:

```go
package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
	"os"
	"os/signal"
)

func main() {
	// Example with listener and client setup
	postsListener, _ := flash.NewListener(&flash.ListenerConfig{Table: "public.posts"})

	postsListener.On(flash.OperationAll, func(event flash.Event) {
		switch typedEvent := event.(type) {
		case *flash.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *flash.UpdateEvent:
			fmt.Printf("update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *flash.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		case *flash.TruncateEvent:
			fmt.Printf("truncate \n")
		}
	})

	// Create client
	flashClient, _ := flash.NewClient(&flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Driver:      trigger.NewDriver(&trigger.DriverConfig{}),
	})
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

- [Debug queries](_examples/debug_trace/main.go)
- [Trigger insert events on table](_examples/trigger_insert/main.go)
- [Trigger all events on table](_examples/trigger_all/main.go)
- [Listen for specific fields](_examples/specific_fields/main.go)
- [Parallel Callback](_examples/parallel_callback/main.go)

## Advanced Features

### 1. Configurable Primary Key ⏳

When you define a primary key, instead of receiving an update event when the column changes, you will receive two
events:

- A delete event with the old value of this column (and other fields).
- An insert event with the new value of this column (and other fields).

### 2. Custom Conditions ✅

You can configure conditions, and if a database row does not match the criteria, you will not receive any event.

In the case of an update:

- If the row previously matched the criteria but the new row does not, you will receive a delete event.
- If the row previously did not match the criteria but the new row does, you will receive an insert event.

### 3. Partial Fields ✅

Ability to listen only to certain columns in your table. If no changes occur in one of these columns, you will not
receive any event.

### ⚠️ Important Notes ⚠️

Some of these features may be incompatible with your driver.

Check [drivers/README.md](pkg/drivers/README.md) to see if the driver you have chosen supports these features.

## Planned Features

The following features are planned for future implementation:

- ⏳ Support for conditional listens.

| Operator |      trigger      |    wal_logical    |
|:--------:|:-----------------:|:-----------------:|
|  equals  |         ✅         |         ✅         |
|   neq    |         ❌         |         ❌         |
|    lt    |         ❌         |         ❌         |
|   lte    |         ❌         |         ❌         |
|   gte    |         ❌         |         ❌         |
| not null |         ❌         |         ❌         |
| is null  | ⚠️ using eq + nil | ⚠️ using eq + nil |

- ⏳ Handling custom primary for fake insert/delete when change appears
- ⬜ Remove client in favor of direct listener start
- ⬜ Support attaching/detaching new listener during runtime.
- ⬜ Tests implementation
- ... any suggestions is welcome.

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


