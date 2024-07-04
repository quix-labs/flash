package client

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"strconv"
)

type Config struct {
	DatabaseCnx string
}

func NewClient(config *Config) *Client {
	if config == nil {
		config = &Config{}
	}
	return &Client{Config: config}
}

type Client struct {
	Config    *Config
	listeners []*Listener
	conn      *pgx.Conn
}

func (c *Client) AddListener(l *Listener) {
	c.listeners = append(c.listeners, l)
}

func (c *Client) Start() error {
	fmt.Println("Init PostgreSQL client")
	var err error
	if c.conn, err = pgx.Connect(context.Background(), c.Config.DatabaseCnx); err != nil {
		return err
	}

	fmt.Println("Init listeners")
	for _, l := range c.listeners {
		err := l.Init(func(event Event) error {
			sql, err := GetCreateTriggerSqlForEvent(l, event, "unique-TODO-"+strconv.Itoa(int(event)), "public")
			if err != nil {
				return err
			}
			_, err = c.conn.Exec(context.Background(), sql)
			if err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		}, func(event Event) error {
			sql, err := GetDeleteTriggerSqlForEvent(l, event, "unique-TODO-"+strconv.Itoa(int(event)), "public")
			if err != nil {
				return err
			}
			_, err = c.conn.Exec(context.Background(), sql)
			if err != nil {
				fmt.Println(err)
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	fmt.Println("Listeners initialized")
	return nil
}

func (c *Client) Close() {
	//if c.conn != nil {
	//	_ = c.conn.Close(context.Background())
	//}
	fmt.Println("CLOSING")
}
