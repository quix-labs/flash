# Trigger Driver (trigger)

## Description

For each event that is listened to, this driver dynamically creates a trigger that uses `pg_notify` to notify the application.

This approach can introduce latencies in the database due to the overhead of creating and managing triggers on-the-fly.

## Prerequisites

### Database Setup

- No configuration needed, triggers are natively supported in PostgreSQL.

## How to Use

Initialize this driver and pass it to the `clientConfig` Driver parameter.

```go
package main
import (
    "github.com/quix-labs/flash/drivers/trigger"
    "github.com/quix-labs/flash"
)

func main() {
	// ... BOOTSTRAPPING
	driver := trigger.NewDriver(&trigger.DriverConfig{})
	clientConfig := &flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Driver:      driver,
	}
	// ...START
}
```
## Configuration

### Schema

- **Type**: `string`
- **Default**: `flash`
- **Description**: Must be unique across all your instances. This schema is used to sandbox all created resources.

## Notes

This driver creates a schema. If you have multiple instances without distinct `Schema` values, you may create conflicts between your applications.

When running multiple clients in parallel, ensure each has unique values for these configurations to avoid conflicts.


## Manually deletion

If you encounter any artifacts, you can simply drop the PostgreSQL schema with your custom-defined schema or the default `flash`. Use `CASCADE` to ensure triggers are deleted.


## Detailed Information

### Advanced Features support

See [Drivers Overview](/drivers/) for compatibility table

### Internal workflow

You can find a workflow graph [here](./WORKFLOW).