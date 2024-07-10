# WAL Logical Driver (wal_logical)

## Description

This driver operates as a replica slave to intercept the replication logs, capturing changes from the primary database.

This approach has a minimal impact on the database performance as it leverages PostgreSQL's built-in replication mechanisms.

## Prerequisites

### Database setup
- Set `replication_level=logical`.
- Set `max_replication_slots` with value of 1 or more.
- Set up your `DatabaseCnx` using a user with replication privileges.

## How to Use

Initialize this driver and pass it to the `clientConfig` Driver parameter.

```go
package main

import (
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/wal_logical"
)

func main() {
	// ... BOOTSTRAPPING
	driver := wal_logical.NewDriver(&wal_logical.DriverConfig{})
	clientConfig := &flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Driver:      driver,
	}
	// ...START
}

```
## Configuration


### PublicationSlotPrefix

- **Type**: `string`
- **Default**: `flash_publication`
- **Description**: Must be unique across all your instances. This prefix is used to create publication slots in the PostgreSQL database.

### ReplicationSlot
- **Type**: `string`
- **Default**: `flash_replication`
- **Description**: Must be unique across all your instances. This slot is used to manage replication data.

### UseStreaming
- **Type**: `bool`
- **Default**: false
- **Description**: Allows the usage of streaming for large transactions. Enabling this can have a significant memory impact.

## Notes

This driver creates a replication slot. If you have multiple instances without distinct `PublicationSlotPrefix` and `ReplicationSlot` values, you may create conflicts between your applications. 

When running multiple clients in parallel, ensure each has unique values for these configurations to avoid conflicts.

## Known Issues

* Currently, this driver can crash on restart if it was not properly closed by calling `client.Close()` during shutdown.
    
    If you encounter this issue, you can manually delete all publication slots from your PostgreSQL instance that start with your defined `PublicationSlotPrefix` or the default fallback `flash_publication`.


## Detailed Information

### Advanced Features support

See [drivers/README.MD](../README.md) for compatibility table

### Internal workflow

You can find a workflow graph [here](./WORKFLOW.md).