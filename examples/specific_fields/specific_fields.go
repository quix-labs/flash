package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
)

func main() {
	postsListener, _ := listeners.NewListener(&types.ListenerConfig{
		Table:  "public.posts",
		Fields: []string{"id", "slug"},
	})
	postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %+v\n", event)
	})

	// Create client
	clientConfig := &types.ClientConfig{DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb"}
	flashClient, _ := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)

	go func() {
		err := flashClient.Start()
		if err != nil {
			panic(err)
		}
	}()
	defer flashClient.Close()

	// Keep process running
	select {}
}
