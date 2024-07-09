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
	postsListener.On(types.OperationAll, func(event types.Event) {
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
