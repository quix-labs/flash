package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
)

func main() {
	postsListener, _ := flash.NewListener(&flash.ListenerConfig{
		Table:              "public.posts",
		MaxParallelProcess: 50, // Default to 1, you can use -1 for infinite goroutine
	})

	stop, err := postsListener.On(flash.OperationInsert|flash.OperationDelete, func(event flash.Event) {
		switch typedEvent := event.(type) {
		case *flash.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *flash.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		}
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	flashClient, _ := flash.NewClient(&flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Driver:      trigger.NewDriver(&trigger.DriverConfig{}),
	})
	flashClient.Attach(postsListener)
	go flashClient.Start() // Error Handling
	defer flashClient.Close()

	select {} // Keep process running
}
