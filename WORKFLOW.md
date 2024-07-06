```mermaid
sequenceDiagram
    participant Your App
    participant Listener
    participant Client
    participant Driver
    participant Database
    participant External

    alt Before Client Start
        Your App->>Listener: on(eventUpdate^EventInsert)
        Your App->>Client: AddListener(listener)
        Note over Your App, Client: Create and register Listener for eventUpdate and EventInsert
    end

    alt Client Start and Trigger Creation
        Your App->>Client: start()
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
        Your App->>Listener: on(eventDelete)
        Listener->>Client: start listening for delete
        Client->>Driver: send listen for delete signal
        Driver->>Database: create trigger on delete
        Note over Your App, Database: Your App starts listening for eventDelete
    end

    alt DELETE operation in Database (handled)
        External-->>Database: DELETE FROM ...
        Database-->>Driver: Send pg_notify events
        Driver-->>Client: Dispatch received event
        Client-->>Listener: Notify listener
        Note over Client, Listener: Client notifies the Listener of the received event
        alt Listener processes event
            Listener-->>Your App: Event processed
            Note over Listener, Your App: Listener processes the event
        end
    end

    alt Stop Listening for Deletion Events
        Your App->>Listener: off(eventDelete)
        Listener->>Client: stop listening for delete
        Client->>Driver: send stop listen for delete signal
        Driver->>Database: delete trigger on delete
        Note over Your App, Database: Client and Driver stop listening for delete events
    end

    alt Application Shutdown
        Your App->>Client: stop()
        activate Client
        Client->>Listener: Listener.Close()
        Listener->>Listener: List all listened events
        
        Listener->>Client: stop listening for insert
        Client->>Driver: send stop listen for insert signal
        Driver-->>Database: delete trigger on insert
        
        Listener->>Client: stop listening for update
        Client->>Driver: send stop listen for update signal
        Driver-->>Database: delete trigger on update
        deactivate Client
        Note over Listener, Driver: Client stops listening for all events
    end
```