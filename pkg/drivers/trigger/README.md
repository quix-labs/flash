# Trigger Driver (trigger)

### Bootstrapping

- Generation of a unique name:
    - Unique reference + action (insert, update, delete, truncate)

- Creation:
    - Create a trigger function -> pg_notify(unique_name)
    - Create a trigger FOR EACH ROW that calls the previous trigger function

- Deletion:
    - Delete the trigger function dedicated to the action (using the unique name as reference) CASCADE
    - As it cascades, the trigger will also be deleted

### Event Reception

- Each event is identifiable by a unique name that it emits in pg_notify.
    - Since we only have triggers for the requested data, we forward the event to the callback in all cases


### WORKFLOW

You can find a workflow graph [here](./WORKFLOW.md).

