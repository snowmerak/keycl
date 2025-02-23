package cache

import (
	"errors"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/valkey-io/valkey-go"
)

type Cache struct {
	client valkey.Client
}

func NewCache(client valkey.Client) *Cache {
	return &Cache{client: client}
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

func NewCacheWithConfig(config valkey.ClientOption) (*Cache, error) {
	client, err := valkey.NewClient(config)
	if err != nil {
		return nil, err
	}
	return NewCache(client), nil
}
