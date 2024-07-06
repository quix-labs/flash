# Approach 1 [DISTINCT TRIGGERS] [DRIVER pg-trigger]

## Bootstrapping

- Generation of a unique name:
    - Unique reference + action (insert, update, delete, truncate)

- Creation:
    - Create a trigger function -> pg_notify(unique_name)
    - Create a trigger FOR EACH ROW that calls the previous trigger function

- Deletion:
    - Delete the trigger function dedicated to the action (using the unique name as reference) CASCADE
    - As it cascades, the trigger will also be deleted

## Event Reception

- Each event is identifiable by a unique name that it emits in pg_notify.
    - Since we only have triggers for the requested data, we forward the event to the callback in all cases

# Approach 2 [GLOBAL UPDATE/DELETE/INSERT TRIGGER + TRUNCATE TRIGGER]

## Bootstrapping

- Generation of a unique name:
    - If TRUNCATE: Unique reference + truncate -> e.g., flash_posts_truncate
    - Otherwise, Unique reference + other -> e.g., flash_posts_other

- Creation:
    - If TRUNCATE -> CREATE TRIGGER ON ... BEFORE TRUNCATE FOR EACH STATEMENT ...
    - Otherwise:
        - If a global trigger already exists -> ignore
        - If no global trigger is registered, create it -> CREATE TRIGGER ON ... BEFORE UPDATE, DELETE, INSERT FOR EACH STATEMENT ...
            - Iterate over old_table and new_table -> for each entry call pg_notify passing TG_OP

## Event Reception

In this case, we will receive unlistened events.

We need to check if the received event is in the list of listened events.

- If yes, send it to the callback
- If not, ignore it

# Approach 3 [GLOBAL UPDATE/DELETE/INSERT TRIGGER + TRUNCATE TRIGGER] BATCH VERSION

## Bootstrapping

- Like Approach 2 but instead of calling pg_notify for each row, generate a JSON array and send the complete payload only once

## Event Reception

- Like Approach 2 but if we receive the payload, decode it and iterate over each entry to send an event for each entry

# Approach 4 [WAL REPLICATION - TO BE REFINED]

## Bootstrapping

- CREATION:
    - Internally add it to our list of events to track
- DELETION:
    - If existing, internally remove it from our list of events to track

## Event Reception

In all cases, we will receive all events, listened to or not.

- Parse the replication log, extract the operation and table + ...
    - Detect the event (INSERT, UPDATE, DELETE, TRUNCATE, ...)
        - If it does not exist internally in our list of events to track -> ignore
        - Otherwise -> forward to the callback

# Approach 5 [PG EXTENSION - FURTHER THOUGHT REQUIRED]

## Bootstrapping

- CREATION:
    - Call custom function to listen
- DELETION:
    - Call custom function to stop listening

## Event Reception

- Retrieve the emitted event
    - Forward it to the callback

# ~~Approach 6 [GATEWAY - SQL]~~

## Bootstrapping

- Open a TCP port
- Intercept SQL queries

## Event Reception

- Parse the SQL query
- Detect the altered rows
    - If listening: forward it to the callback
    - Otherwise: ignore it

## Rejected

For UPDATE FROM (SELECT id from posts) queries, it is impossible to track the rows without making database queries.
