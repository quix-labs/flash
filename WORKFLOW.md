```mermaid
sequenceDiagram
    participant Main
    participant Listener
    participant Client
    participant Driver
    participant Database
    participant External

    alt Before Client Start
        Main->>Listener: on(eventUpdate^EventInsert)
        Main->>Client: AddListener(listener)
        Note over Main, Client: Main adds Listener for eventUpdate and EventInsert
    end

    alt Client Start and Trigger Creation
        Main->>Client: start()
        activate Client
        Client->>Driver: send listen for update signal
        Driver->>Database: create trigger on update
        Client->>Driver: send listen for insert signal
        Driver->>Database: create trigger on insert
        deactivate Client
        Note over Client, Driver: Client starts listening for update and insert events
    end

    alt DELETE operation in Database (unhandled)
        External-->>Database: DELETE FROM ...
    end

    alt Listen for New Database Changes on Deletion (DELETE)
        Main->>Listener: on(eventDelete)
        Listener->>Client: start listening for delete
        Client->>Driver: send listen for delete signal
        Driver->>Database: create trigger on delete
        Note over Main, Database: Main starts listening for eventDelete
    end

    alt DELETE operation in Database (handled)
        External-->>Database: DELETE FROM ...
        Database-->>Driver: Send pg_notify events
        Driver-->>Client: Dispatch received event
        Client-->>Listener: Notify listener
        Note over Client, Listener: Client notifies the Listener of the received event
        alt Listener processes event
            Listener-->>Main: Event processed
            Note over Listener, Main: Listener processes the event
        end
    end

    alt Stop Listening for Deletion Events
        Main->>Listener: off(eventDelete)
        Listener->>Client: stop listening for delete
        Client->>Driver: send stop listening for delete signal
        Driver->>Database: delete trigger on delete
        Note over Main, Database: Client and Driver stop listening for delete events
    end

    alt Application Shutdown
        Main->>Client: stop()
        activate Client
        Client->>Driver: stop listening for insert
        Driver-->>Database: delete trigger on insert
        Client->>Driver: stop listening for update
        Driver-->>Database: delete trigger on update
        deactivate Client
        Note over Client, Driver: Client stops listening for all events
    end
```
