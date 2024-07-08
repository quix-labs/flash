package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	postsListenerConfig := &types.ListenerConfig{
		Table:              "public.posts",
		MaxParallelProcess: 50, // Default to 1, you can use -1 for infinite goroutine
	}
	postsListener, _ := listeners.NewListener(postsListenerConfig)

	stop, err := postsListener.On(types.OperationInsert|types.OperationDelete, func(event types.Event) {
		switch typedEvent := event.(type) {
		case *types.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *types.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		}
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	clientConfig := &types.ClientConfig{DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb"}
	flashClient, _ := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)
	go flashClient.Start() // Error Handling
	defer flashClient.Close()

	select {} // Keep process running
}
