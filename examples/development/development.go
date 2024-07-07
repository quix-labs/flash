package main

import (
	"fmt"
	"github.com/quix-labs/flash/pkg/client"
	"github.com/quix-labs/flash/pkg/listeners"
	"github.com/quix-labs/flash/pkg/types"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"time"
)

func main() {
	f, err := os.Create("myprogram.prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	postsListenerConfig := &types.ListenerConfig{
		Table:              "public.posts",
		MaxParallelProcess: 1,
		Fields:             []string{"id", "slug"},
	}
	postsListener, _ := listeners.NewListener(postsListenerConfig)

	// Registering your callbacks
	var i = 0
	var mutex sync.Mutex
	stop, err := postsListener.On(types.EventsAll, func(event *types.ReceivedEvent) {
		mutex.Lock()
		i++
		mutex.Unlock()
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		err := stop()
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		for {
			time.Sleep(time.Second)
			mutex.Lock()
			fmt.Println(i)
			i = 0
			mutex.Unlock()
		}

	}()

	// Create custom logger
	logger := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Stack().Timestamp().Logger()

	// Create client
	clientConfig := &types.ClientConfig{
		DatabaseCnx:     "postgresql://devuser:devpass@localhost:5432/devdb",
		Logger:          &logger, // Define your custom zerolog.Logger here
		ShutdownTimeout: time.Second * 2,
	}
	flashClient, _ := client.NewClient(clientConfig)
	flashClient.Attach(postsListener)

	// Start listening
	go func() {
		err := flashClient.Start()
		if err != nil {
			panic(err)
		}
	}() // Error Handling

	defer func() {
		err := flashClient.Close()
		if err != nil {
			panic(err)
		}
	}()

	// Wait for interrupt signal (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt

	fmt.Println("Program terminated.")
}
