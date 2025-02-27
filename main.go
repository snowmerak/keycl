package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/snowmerak/keycl/lib/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	c := cli.New(cli.Valkey, "")

	if err := c.CreateCluster(ctx, 0, "127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003"); err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	if err := c.AddNode(ctx, "127.0.0.1", 7005, "127.0.0.1", 7002); err != nil {
		panic(err)
	}

	time.Sleep(3 * time.Second)

	if err := c.ReshardAll(ctx, "127.0.0.1", 7001); err != nil {
		panic(err)
	}

	nodes, err := c.GetClusterNodes(ctx, "127.0.0.1", 7001)
	if err != nil {
		panic(err)
	}

	for _, node := range nodes {
		log.Info().Any("node", node).Msg("node")
	}

	if err := c.ExceptNode(ctx, "127.0.0.1", 7001, "2b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
		panic(err)
	}

	if err := c.Rebalance(ctx, "127.0.0.1", 7001); err != nil {
		panic(err)
	}

	log.Info().Msg("done")
}
