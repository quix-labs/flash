# What is Flash?

Flash is a lightweight Go library that monitors and processes real-time changes in PostgreSQL databases. Designed for event-driven architectures, Flash makes it easy to track database events like inserts, updates, deletes, and truncations while minimizing performance overhead.

Built for developers who need precision and reliability, Flash ensures your database changes are handled efficiently and seamlessly integrated into your application workflows.

<div class="tip custom-block" style="padding-top: 8px">

Want to try it out? Jump straight to the [Quickstart](./installation).

</div>

## Use Cases

Flash is perfect for scenarios where real-time database monitoring is essential:

- **Live Data Dashboards**: Update UI components dynamically as data changes in your database.
- **Event-Driven Architectures**: Trigger workflows or notifications in response to specific database events.
- **Data Syncing**: Sync changes to downstream systems like caches, search engines, or analytics platforms.
- **Audit Logging**: Track and log database modifications for compliance and traceability.

## Features

- ✅ Start/Stop listening during runtime.
- ✅ Supports common PostgreSQL events: Insert, Update, Delete, Truncate.
- ✅ Driver interfaces for creating new drivers.
- ✅ Parallel Callback execution using goroutine
- ✅ Listen for changes in specific columns, not the entire row. (see [Advanced Features](./advanced-features.md))
- ✅ Listen changes using WAL replication (see [Drivers](./drivers/))


## Supported Platforms

**Database**: PostgreSQL  
**Drivers**:
- Trigger-based
- WAL-based


---

Check out the [Quickstart](./installation) and see how easy it is to integrate real-time database monitoring into your Go applications.
