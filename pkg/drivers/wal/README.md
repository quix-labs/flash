# WAL driver (wal)

### Bootstrapping

- CREATION:
    - Internally add it to our list of events to track
- DELETION:
    - If existing, internally remove it from our list of events to track

### Event Reception

In all cases, we will receive all events, listened to or not.

- Parse the replication log, extract the operation and table + ...
    - Detect the event (INSERT, UPDATE, DELETE, TRUNCATE, ...)
        - If it does not exist internally in our list of events to track -> ignore
        - Otherwise -> forward to the callback
