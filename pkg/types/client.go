package types

import (
	"github.com/rs/zerolog"
	"time"
)

type ClientConfig struct {
	DatabaseCnx string
	Driver      Driver
	Logger      *zerolog.Logger

	ShutdownTimeout time.Duration
}

type DatabaseEvent struct {
	ListenerUid   string
	ReceivedEvent *ReceivedEvent
}
type DatabaseEventsChan chan *DatabaseEvent
