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

	// Registering your callbacks -> Can be simplified with types.EventAll
	stop, err := postsListener.On(types.EventTruncate|types.EventInsert|types.EventUpdate|types.EventDelete, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %+v\n", event)
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
