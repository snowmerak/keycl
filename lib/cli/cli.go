package cli

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

type ClusterNode struct {
	ID          string
	Host        string
	ClusterPort int
	Flags       []string
	LinkState   string
	Slots       []int
}

func (cli *CLI) GetClusterNodes(ctx context.Context, host string, port int) ([]*ClusterNode, error) {
	cmd := exec.CommandContext(ctx, string(cli.name), "-h", host, "-p", strconv.Itoa(port), "cluster", "nodes")
	resp, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run command %s %v: %w", cli.name, []string{"-h", host, "-p", strconv.Itoa(port), "cluster", "nodes"}, err)
	}

	nodes := make([]*ClusterNode, 0)

	sc := bufio.NewScanner(bytes.NewReader(resp))
	for sc.Scan() {
		line := strings.TrimPrefix(sc.Text(), "txt:")
		if line == "" {
			continue
		}

		node := &ClusterNode{}
		prevIdx := 0
		state := 0
	loop:
		for i := 0; i < len(line); i++ {
			switch state {
			case 0:
				if line[i] == ' ' {
					node.ID = line[prevIdx:i]
					state = 1
					prevIdx = i + 1
				}
			case 1:
				if line[i] == '@' {
					node.Host = line[prevIdx:i]
					state = 2
					prevIdx = i + 1
				}
			case 2:
				if line[i] == ' ' {
					port, _ := strconv.Atoi(line[prevIdx:i])
					node.ClusterPort = port
					state = 3
					prevIdx = i + 1
				}
			case 3:
				if line[i] == ' ' {
					node.Flags = append(node.Flags, strings.Split(line[prevIdx:i], ",")...)
					state = 4
					prevIdx = i + 1
				}
			case 4:
				if 'a' <= line[i] && line[i] <= 'z' || 'A' <= line[i] && line[i] <= 'Z' {
					prevIdx = i
					state = 5
				}
			case 5:
				if line[i] == ' ' {
					node.LinkState = line[prevIdx:i]
					state = 6
				}
			case 6:
				if line[i] == ' ' || len(line)-1 == i {
					if len(line)-1 == i {
						i++
					}
					slots := strings.Split(line[prevIdx:i], "-")
					if len(slots) == 1 {
						slot, _ := strconv.Atoi(slots[0])
						node.Slots = append(node.Slots, slot)
					} else {
						start, _ := strconv.Atoi(slots[0])
						end, _ := strconv.Atoi(slots[1])
						node.Slots = append(node.Slots, start, end)
					}
					state = 7
				}
			case 7:
				break loop
			}
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

type ClusterInfo struct {
	ClusterState         string `json:"cluster_state"`
	ClusterSlotsAssigned int    `json:"cluster_slots_assigned"`
	ClusterSlotsOk       int    `json:"cluster_slots_ok"`
	ClusterSlotsPfail    int    `json:"cluster_slots_pfail"`
	ClusterSlotsFail     int    `json:"cluster_slots_fail"`
	ClusterKnownNodes    int    `json:"cluster_known_nodes"`
	ClusterSize          int    `json:"cluster_size"`
	ClusterMyEpoch       int    `json:"cluster_my_epoch"`
}

func (cli *CLI) GetClusterInfo(ctx context.Context, host string, port int) (*ClusterInfo, error) {
	cmd := exec.CommandContext(ctx, string(cli.name), "-h", host, "-p", strconv.Itoa(port), "cluster", "info")
	resp, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run command %s %v: %w", cli.name, []string{"-h", host, "-p", strconv.Itoa(port), "cluster", "info"}, err)
	}

	info := &ClusterInfo{}

	sc := bufio.NewScanner(bytes.NewReader(resp))
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			continue
		}

		idx := len(line)
		for ; idx > 0; idx-- {
			if line[idx-1] == ':' {
				break
			}
		}
		key := line[:idx-1]
		value := line[idx:]
		switch key {
		case "cluster_state":
			info.ClusterState = value
		case "cluster_slots_assigned":
			info.ClusterSlotsAssigned, _ = strconv.Atoi(value)
		case "cluster_slots_ok":
			info.ClusterSlotsOk, _ = strconv.Atoi(value)
		case "cluster_slots_pfail":
			info.ClusterSlotsPfail, _ = strconv.Atoi(value)
		case "cluster_slots_fail":
			info.ClusterSlotsFail, _ = strconv.Atoi(value)
		case "cluster_known_nodes":
			info.ClusterKnownNodes, _ = strconv.Atoi(value)
		case "cluster_size":
			info.ClusterSize, _ = strconv.Atoi(value)
		case "cluster_my_epoch":
			info.ClusterMyEpoch, _ = strconv.Atoi(value)
		}
	}

	return info, nil
}

func (cli *CLI) AddNode(ctx context.Context, newNode, existingNode string) error {
	args := []string{"--cluster", "add-node", newNode, existingNode}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	return nil
}

func (cli *CLI) Reshard(ctx context.Context, node string) error {
	args := []string{"--cluster", "reshard", node, "--cluster-yes"}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	// cmd.Stdin = bytes.NewReader([]byte("all\nall\nyes\n"))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	return nil
}
