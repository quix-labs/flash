package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
)

func main() {
	postsListenerConfig := &types.ListenerConfig{
		Table: "public.posts",
		//Fields: []string{"id", "slug"},
	}
	postsListener := listeners.NewListener(postsListenerConfig)

	// Registering your callbacks
	stop, err := postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %+v\n", event)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	// Create custom logger
	logger := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Stack().Timestamp().Logger()

	// Create client
	clientConfig := &types.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Logger:      &logger, // Define your custom zerolog.Logger here
	}
	flashClient := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)

	// Start listening
	go func() {
		err := flashClient.Start()
		if err != nil {
			panic(err)
		}
	}() // Error Handling
	defer flashClient.Close()

	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
