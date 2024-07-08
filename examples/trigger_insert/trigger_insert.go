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

	// Registering your callbacks
	stop, err := postsListener.On(types.OperationInsert, func(event types.Event) {
		typedEvent := event.(*types.InsertEvent)
		fmt.Printf("insert - new: %+v\n", typedEvent.New)
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