package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/snowmerak/keycl/lib/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	c := cli.New(cli.Valkey)
	if err := c.CreateCluster(ctx, []string{"127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003"}, 0); err != nil {
		panic(err)
	}
}
