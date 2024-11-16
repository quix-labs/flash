# Advanced Features

:::warning ⚠️ Important Notes ⚠️

Some of these features may be incompatible with your driver.

Check [Drivers Overview](./drivers/) to see if the driver you have chosen supports these features.

:::

[//]: # (TODO Better doc)

For more detailed examples, check out the following files:

- [Debug queries](https://github.com/quix-labs/flash/tree/main/_examples/debug_trace/main.go)
- [Trigger insert events on table](https://github.com/quix-labs/flash/tree/main/_examples/trigger_insert/main.go)
- [Trigger all events on table](https://github.com/quix-labs/flash/tree/main/_examples/trigger_all/main.go)
- [Listen for specific fields](https://github.com/quix-labs/flash/tree/main/_examples/specific_fields/main.go)
- [Parallel Callback](https://github.com/quix-labs/flash/tree/main/_examples/parallel_callback/main.go)



## 1. Configurable Primary Key ⏳

When you define a primary key, instead of receiving an update event when the column changes, you will receive two
events:

- A delete event with the old value of this column (and other fields).
- An insert event with the new value of this column (and other fields).

## 2. Custom Conditions ⏳

You can configure conditions, and if a database row does not match the criteria, you will not receive any event.

In the case of an update:

- If the row previously matched the criteria but the new row does not, you will receive a delete event.
- If the row previously did not match the criteria but the new row does, you will receive an insert event.

## 3. Partial Fields ✅

Ability to listen only to certain columns in your table. If no changes occur in one of these columns, you will not
receive any event.

