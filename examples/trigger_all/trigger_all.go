package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	postsListenerConfig := &types.ListenerConfig{Table: "public.posts"}
	postsListener := listeners.NewListener(postsListenerConfig)

	// Registering your callbacks
	stop, err := postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %+v\n", event)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	clientConfig := &types.ClientConfig{DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb"}
	flashClient := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)
	go flashClient.Start() // Error Handling
	defer flashClient.Close()

	select {} // Keep process running
}
