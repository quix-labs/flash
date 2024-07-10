package main

import (
	"fmt"
	"github.com/quix-labs/flash"
	"github.com/quix-labs/flash/drivers/wal_logical"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	//f, err := os.Create("myprogram.prof")
	//if err != nil {
	//	panic(err)
	//}
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()
	fmt.Println((flash.OperationTruncate).IncludeAll(flash.OperationDelete))
	return
	postsListenerConfig := &flash.ListenerConfig{
		Table:              "public.posts",
		MaxParallelProcess: 1, // In most case 1 is ideal because sync between goroutine introduce some delay
		Fields:             []string{"id", "slug"},
		Conditions:         []*flash.ListenerCondition{{Column: "active", Value: true}},
	}
	postsListener, _ := flash.NewListener(postsListenerConfig)

	postsListener2Config := &flash.ListenerConfig{
		Table:              "public.posts",
		MaxParallelProcess: 1, // In most case 1 is ideal because sync between goroutine introduce some delay
		Fields:             []string{"active"},
		Conditions:         []*flash.ListenerCondition{{Column: "slug", Value: nil}},
	}
	postsListener2, _ := flash.NewListener(postsListener2Config)

	// Registering your callbacks
	var i = 0
	var mutex sync.Mutex

	stopAll, err := postsListener.On(flash.OperationAll, func(event flash.Event) {
		mutex.Lock()
		i++
		mutex.Unlock()

		switch typedEvent := event.(type) {
		case *flash.InsertEvent:
			fmt.Printf("insert - new: %+v\n", typedEvent.New)
		case *flash.UpdateEvent:
			fmt.Printf("update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *flash.DeleteEvent:
			fmt.Printf("delete - old: %+v \n", typedEvent.Old)
		case *flash.TruncateEvent:
			fmt.Printf("truncate \n")
		}
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		err := stopAll()
		if err != nil {
			panic(err)
		}
	}()

	stopAll2, err := postsListener2.On(flash.OperationAll, func(event flash.Event) {
		mutex.Lock()
		i++
		mutex.Unlock()

		switch typedEvent := event.(type) {
		case *flash.InsertEvent:
			fmt.Printf("2-insert - new: %+v\n", typedEvent.New)
		case *flash.UpdateEvent:
			fmt.Printf("2-update - old: %+v - new: %+v\n", typedEvent.Old, typedEvent.New)
		case *flash.DeleteEvent:
			fmt.Printf("2-delete - old: %+v \n", typedEvent.Old)
		case *flash.TruncateEvent:
			fmt.Printf("2-truncate \n")
		}
	})
	if err != nil {
		panic(err)
	}

	defer func() {
		err := stopAll2()
		if err != nil {
			panic(err)
		}
	}()

	//go func() {
	//	for {
	//		time.Sleep(time.Second * 1)
	//		mutex.Lock()
	//		fmt.Println(i)
	//		i = 0
	//		mutex.Unlock()
	//	}
	//}()

	// Create custom logger
	logger := zerolog.New(os.Stdout).Level(zerolog.TraceLevel).With().Caller().Stack().Timestamp().Logger()

	driver := wal_logical.NewDriver(&wal_logical.DriverConfig{
		//UseStreaming: true,
	})

	// Create client
	clientConfig := &flash.ClientConfig{
		DatabaseCnx:     "postgresql://devuser:devpass@localhost:5432/devdb",
		Logger:          &logger, // Define your custom zerolog.Logger here
		ShutdownTimeout: time.Second * 2,
		Driver:          driver,
	}
	flashClient, _ := flash.NewClient(clientConfig)
	flashClient.Attach(postsListener, postsListener2)

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
