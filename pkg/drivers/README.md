# Drivers

## Implemented

| Name                                 | DB impact | INSERT | UPDATE | DELETE | TRUNCATE | Support Partial Fields |       Where clauses        |                   Graceful Shutdown/Restart                    |
|--------------------------------------|:---------:|:------:|:------:|:------:|:--------:|:----------------------:|:--------------------------:|:--------------------------------------------------------------:|
| [trigger](./trigger/README.md)       | high  ⚠️  |   ✅    |   ✅    |   ✅    |    ✅     |           ✅            | not implemented (possible) |                               ✅                                |
| [wal_replica](wal_replica/README.md) |   low ⚡   |   ✅    |   ✅    |   ✅    |    ✅     |     ❌ ***(wip)***      | not implemented (possible) | partial ⚠️ <br/>cannot restart if crash without client.Close() |

## NOT IMPLEMENTED

### GLOBAL UPDATE/DELETE/INSERT TRIGGER + TRUNCATE TRIGGER FOR EACH ROW *(Seems legit)*

#### Bootstrapping

- Generation of a unique name:
    - If TRUNCATE: Unique reference + truncate -> e.g., flash_posts_truncate
    - Otherwise, Unique reference + other -> e.g., flash_posts_other

- Creation:
    - If TRUNCATE -> CREATE TRIGGER ON ... BEFORE TRUNCATE FOR EACH STATEMENT ...
    - Otherwise:
        - If a global trigger already exists -> ignore
        - If no global trigger is registered, create it -> CREATE TRIGGER ON ... BEFORE UPDATE, DELETE, INSERT FOR EACH
          STATEMENT ...
            - Iterate over old_table and new_table -> for each entry call pg_notify passing TG_OP

#### Event Reception

In this case, we will receive unlistened events.

We need to check if the received event is in the list of listened events.

- If yes, send it to the callback
- If not, ignore it

___

### GLOBAL UPDATE/DELETE/INSERT TRIGGER + TRUNCATE TRIGGER [FOR EACH STATEMENT] *(Seems Legit)*

#### Bootstrapping

- Like Approach 2 but instead of calling pg_notify for each row, generate a JSON array and send the complete payload
  only once

#### Event Reception

- Like Approach 2 but if we receive the payload, decode it and iterate over each entry to send an event for each entry

___

### PG EXTENSION - *(FURTHER THOUGHT REQUIRED)*

#### Bootstrapping

- CREATION:
    - Call custom function to listen
- DELETION:
    - Call custom function to stop listening

#### Event Reception

- Retrieve the emitted event
    - Forward it to the callback

___

### ~~GATEWAY (Rejected)~~

#### Bootstrapping

- Open a TCP port
- Intercept SQL queries

#### Event Reception

- Parse the SQL query
- Detect the altered rows
    - If listening: forward it to the callback
    - Otherwise: ignore it

#### Rejection Reason

For UPDATE FROM (SELECT id from posts) queries, it is impossible to track the rows without making database queries.
