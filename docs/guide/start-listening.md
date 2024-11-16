# Start listening

Here's a basic example of how to use Flash:

```go
package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
	"os"
	"os/signal"
)

func main() {
	// Example with listener and client setup
	postsListener, _ := flash.NewListener(&flash.ListenerConfig{Table: "public.posts"})

	postsListener.On(flash.OperationAll, func(event flash.Event) {
		switch typedEvent := event.(type) {
		case *flash.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *flash.UpdateEvent:
			fmt.Printf("update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *flash.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		case *flash.TruncateEvent:
			fmt.Printf("truncate \n")
		}
	})

	// Create client
	flashClient, _ := flash.NewClient(&flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Driver:      trigger.NewDriver(&trigger.DriverConfig{}),
	})
	flashClient.Attach(postsListener)

	// Start listening
	go flashClient.Start()
	defer flashClient.Close()

	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}

```

## TODO: How events working ? (Listen for all, for truncate + delete, ...)