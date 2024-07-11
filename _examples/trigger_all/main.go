package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
)

func main() {
	postsListener, _ := flash.NewListener(&flash.ListenerConfig{Table: "public.posts"})

	// Registering your callbacks -> Can be simplified with types.EventAll
	stop, err := postsListener.On(flash.OperationTruncate|flash.OperationInsert|flash.OperationUpdate|flash.OperationDelete, func(event flash.Event) {
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
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	flashClient, _ := flash.NewClient(&flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb?sslmode=disable",
		Driver:      trigger.NewDriver(&trigger.DriverConfig{}),
	})
	flashClient.Attach(postsListener)
	go flashClient.Start() // Error Handling
	defer flashClient.Close()

	select {} // Keep process running
}
