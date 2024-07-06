package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/types"
	"os"
	"os/signal"
)

func main() {
	postsListener := client.NewListener(&types.ListenerConfig{
		Table: "posts",
	})
	stop, _ := postsListener.On(types.EventsAll, func(event types.Event) {
		fmt.Println("Event received All" + string(event))
	})
	defer stop()

	flashClient := client.NewClient(&types.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
	})
	flashClient.AddListener(postsListener)
	go func() {
		err := flashClient.Start()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
	}()
	defer flashClient.Close()
	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
