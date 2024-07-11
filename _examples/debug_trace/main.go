package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/trigger"
	"github.com/rs/zerolog"
	"os"
)

func main() {

	postsListenerConfig := &flash.ListenerConfig{Table: "public.posts"}
	postsListener, _ := flash.NewListener(postsListenerConfig)

	// Registering your callbacks
	stop, err := postsListener.On(flash.OperationInsert, func(event flash.Event) {
		typedEvent := event.(*flash.InsertEvent)
		fmt.Printf("Insert received - new: %+v\n", typedEvent.New)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stop()

	// Create custom logger with Level Trace <-> Default is Debug
	logger := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Stack().Timestamp().Logger()
	driver := trigger.NewDriver(&trigger.DriverConfig{})
	// Create client
	clientConfig := &flash.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb?sslmode=disable",
		Logger:      &logger, // Define your custom zerolog.Logger here
		Driver:      driver,
	}

	flashClient, err := flash.NewClient(clientConfig)
	if err != nil {
		fmt.Println(err)
	}
	flashClient.Attach(postsListener)

	// Start listening
	go func() {
		err := flashClient.Start()
		if err != nil {
			panic(err)
		}
	}() // Error Handling
	defer func(flashClient *flash.Client) {
		err := flashClient.Close()
		if err != nil {
			panic(err)
		}
	}(flashClient)

	// Keep process running
	select {}
}
