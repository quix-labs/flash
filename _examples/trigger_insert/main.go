package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
)

func main() {
	postsListener, _ := flash.NewListener(&flash.ListenerConfig{Table: "public.posts"})

	// Registering your callbacks
	stop, err := postsListener.On(flash.OperationInsert, func(event flash.Event) {
		typedEvent := event.(*flash.InsertEvent)
		fmt.Printf("insert - new: %+v\n", typedEvent.New)
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
