package types

import (
	"github.com/rs/zerolog"
)

type ClientConfig struct {
	DatabaseCnx string
	Driver      Driver
	Logger      *zerolog.Logger
}

type DatabaseEvent struct {
	ListenerUid   string
	ReceivedEvent *ReceivedEvent
}
type DatabaseEventsChan chan *DatabaseEvent
