package client

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

type Client struct {
	client valkey.Client
}

func New(client valkey.Client) *Client {
	return &Client{client: client}
}

func ClientOptionFromEnv() valkey.ClientOption {
	address := os.Getenv("VALKEY_ADDRESS")
	if address == "" {
		log.Warn().Err(errors.New("VALKEY_ADDRESS is not set, using default")).Str("default_value", "localhost:6379").Msg("using default address")
		address = "localhost:6379"
	}

	password := os.Getenv("VALKEY_PASSWORD")
	if password == "" {
		log.Warn().Err(errors.New("VALKEY_PASSWORD is not set")).Msg("using empty password")
	}

	return valkey.ClientOption{
		InitAddress: []string{address},
		Password:    password,
	}
}

func NewClientWithConfig(config valkey.ClientOption) (*Client, error) {
	client, err := valkey.NewClient(config)
	if err != nil {
		return nil, err
	}
	return New(client), nil
}
