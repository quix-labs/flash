package main

import (
	"fmt"
	"github.com/quix-labs/flash/client"
	"os"
	"os/signal"
)

func main() {
	postsListener := client.NewListener(&client.ListenerConfig{Table: "posts"})
	stop, _ := postsListener.On(client.EventUpdate^client.EventDelete, func(event client.Event) {
		fmt.Println("Event received All" + string(event))
	})
	defer stop()

	flashClient := client.NewClient(&client.Config{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
	})
	flashClient.AddListener(postsListener)
	go flashClient.Start()
	defer flashClient.Close()
	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
