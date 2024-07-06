package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	postsListener := listeners.NewListener(&types.ListenerConfig{
		Table: "public.posts",
	})
	stopAll, err := postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		fmt.Printf("Event received: %d\n", event)
	})
	if err != nil {
		fmt.Println(err)
	}
	defer stopAll()

	logger := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Stack().Timestamp().Logger()
	flashClient := client.NewClient(&types.ClientConfig{
		DatabaseCnx: "postgresql://devuser:devpass@localhost:5432/devdb",
		Logger:      &logger,
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
	//
	time.Sleep(time.Second * 5)
	fmt.Println("stopAll")
	err = stopAll()
	fmt.Println(err)

	stopOther, err := postsListener.On(types.EventUpdate^types.EventInsert, func(event *types.ReceivedEvent) {
		fmt.Println("Event received")
	})
	defer stopOther()
	// Wait for interrupt signal (Ctrl+C)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-interrupt

	fmt.Println("Program terminated.")
}
