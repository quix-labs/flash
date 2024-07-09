package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	postsListenerConfig := &types.ListenerConfig{Table: "public.posts"}
	postsListener, _ := listeners.NewListener(postsListenerConfig)

	// Registering your callbacks -> Can be simplified with types.EventAll
	stop, err := postsListener.On(types.OperationTruncate|types.OperationInsert|types.OperationUpdate|types.OperationDelete, func(event types.Event) {
		switch typedEvent := event.(type) {
		case *types.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *types.UpdateEvent:
			fmt.Printf("update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *types.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		case *types.TruncateEvent:
			fmt.Printf("truncate \n")
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
