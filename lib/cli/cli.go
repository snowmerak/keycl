package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func RunCommand(ctx context.Context, command string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to run command %s %v: %w", command, args, err)
	}
	return []byte{}, nil
}

type CliName string

const (
	Redis  CliName = "redis-cli"
	Valkey CliName = "valkey-cli"
)

type CLI struct {
	name CliName
}

func New(name CliName) *CLI {
	return &CLI{name: name}
}

func (cli *CLI) CreateCluster(ctx context.Context, addresses []string, replicas int) error {
	args := []string{"--cluster", "create"}
	args = append(args, addresses...)
	args = append(args, "--cluster-replicas", strconv.FormatInt(int64(replicas), 10))

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	cmd.Stdin = bytes.NewReader([]byte("yes\n"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	return nil
}
