package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
)

const (
	MaxSlotCount = 16384
)

func RunCommand(ctx context.Context, command string, args []string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, command, args...)
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
	name     CliName
	password string
}

func New(name CliName, password string) *CLI {
	return &CLI{name: name, password: password}
}

func (cli *CLI) CreateCluster(ctx context.Context, replicas int, address ...string) error {
	args := []string{"--cluster", "create"}
	args = append(args, address...)
	args = append(args, "--cluster-replicas", strconv.FormatInt(int64(replicas), 10))

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("create cluster")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	cmd.Stdin = bytes.NewReader([]byte("yes\n"))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish create cluster")

	return nil
}

type ClusterNode struct {
	ID          string
	Host        string
	ClusterPort int
	Flags       []string
	MasterID    string
	LinkState   string
	Slots       []int
}

func (cli *CLI) GetClusterNodes(ctx context.Context, host string, port int) ([]*ClusterNode, error) {
	args := []string{string(cli.name), "-h", host, "-p", strconv.Itoa(port), "cluster", "nodes"}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("get cluster nodes")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	resp, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run command %s %v: %w", cli.name, []string{"-h", host, "-p", strconv.Itoa(port), "cluster", "nodes"}, err)
	}

	nodes := make([]*ClusterNode, 0)

	isSlave := false
	checkMasterID := false
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
				inner:
					for _, flag := range node.Flags {
						if flag == "slave" {
							isSlave = true
							break inner
						}
					}
					state = 4
					prevIdx = i + 1
				}
			case 4:
				if isSlave && !checkMasterID {
					if line[i] == ' ' {
						checkMasterID = true
						node.MasterID = line[prevIdx:i]
						prevIdx = i + 1
						continue loop
					}
					continue loop
				}
				if 'a' <= line[i] && line[i] <= 'z' || 'A' <= line[i] && line[i] <= 'Z' {
					prevIdx = i
					state = 5
				}
			case 5:
				if line[i] == ' ' || len(line)-1 == i {
					node.LinkState = line[prevIdx:i]
					prevIdx = i + 1
					state = 6
				}
			case 6:
				if len(line)-1 == i {
					i++
					slotSet := strings.Split(line[prevIdx:i], " ")
					fmt.Printf("slotSet: %v\n", slotSet)
					for _, slotPair := range slotSet {
						slots := strings.Split(slotPair, "-")
						if len(slots) == 1 {
							slot, _ := strconv.Atoi(strings.TrimSpace(slots[0]))
							node.Slots = append(node.Slots, slot)
						} else {
							start, _ := strconv.Atoi(slots[0])
							end, _ := strconv.Atoi(slots[1])
							node.Slots = append(node.Slots, start, end)
						}
					}
					state = 7
				}
			case 7:
				break loop
			}
		}

		nodes = append(nodes, node)
	}

	log.Info().Interface("nodes", nodes).Msg("finish get cluster nodes")

	return nodes, nil
}

func (cli *CLI) GetNoSlotNodes(ctx context.Context, host string, port int) ([]string, int, error) {
	nodes, err := cli.GetClusterNodes(ctx, host, port)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	noSlotNodes := make([]string, 0)
	for _, node := range nodes {
		if len(node.Slots) == 0 {
			noSlotNodes = append(noSlotNodes, node.ID)
		}
	}

	return noSlotNodes, len(nodes), nil
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
	args := []string{string(cli.name), "-h", host, "-p", strconv.Itoa(port), "cluster", "info"}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("get cluster info")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
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

	log.Info().Interface("info", info).Msg("finish get cluster info")

	return info, nil
}

func (cli *CLI) AddNode(ctx context.Context, newNodeHost string, newNodePort int, existingNodeHost string, existingNodePort int) error {
	newNode := newNodeHost + ":" + strconv.FormatInt(int64(newNodePort), 10)
	existingNode := existingNodeHost + ":" + strconv.FormatInt(int64(existingNodePort), 10)
	args := []string{"--cluster", "add-node", newNode, existingNode}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("add node")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish add node")

	return nil
}

func (cli *CLI) Reshard(ctx context.Context, host string, port int, targetNode string, slots int, sourceNode string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args := []string{"--cluster", "reshard", host + ":" + strconv.FormatInt(int64(port), 10), "--cluster-yes"}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("reshard")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	ipr, ipw := io.Pipe()
	opr, opw := io.Pipe()

	writeBuffer := opw
	readBuffer := ipr
	reactor := NewReactor(readBuffer, writeBuffer)
	reactor.AddReaction("How many slots do you want to move", strconv.FormatInt(int64(slots), 10))
	reactor.AddReaction("What is the receiving node ID", targetNode)
	reactor.AddReaction("Please enter all the source node IDs", sourceNode)
	if sourceNode != "all" {
		reactor.AddReaction("Source node #2", "done")
	}
	reactor.AddReaction("Do you want to proceed with the proposed reshard plan", "yes")

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	cmd.Stdin = opr
	cmd.Stdout = ipw
	cmd.Stderr = ipw
	go func() {
		reactor.React(func() {
			cancel()

			if cmd.ProcessState != nil && !cmd.ProcessState.Exited() {
				if err := cmd.Process.Kill(); err != nil {
					log.Error().Err(err).Msg("failed to kill process")
				}
			} else {
				log.Info().Int("exitCode", cmd.ProcessState.ExitCode()).Msg("process exited")
			}

			if err := ipw.Close(); err != nil {
				log.Error().Err(err).Msg("failed to close pipe")
			}

			if err := opr.Close(); err != nil {
				log.Error().Err(err).Msg("failed to close pipe")
			}
		})
	}()
	if err := cmd.Run(); err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, io.ErrClosedPipe) {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish reshard")

	return nil
}

func (cli *CLI) ReshardAll(ctx context.Context, host string, port int) error {
	noSlotNodes, allNodeCount, err := cli.GetNoSlotNodes(ctx, host, port)
	if err != nil {
		return fmt.Errorf("failed to get nodes without slots: %w", err)
	}

	log.Info().Strs("noSlotNodes", noSlotNodes).Int("allNodeCount", allNodeCount).Msg("all node information")

	slotCount := MaxSlotCount / allNodeCount
	for _, node := range noSlotNodes {
		if err := cli.Reshard(ctx, host, port, node, slotCount, "all"); err != nil {
			return fmt.Errorf("failed to reshard: %w", err)
		}
	}

	log.Info().Msg("finish reshard all")

	return nil
}

func (cli *CLI) ForgetNode(ctx context.Context, host string, port int, nodeID string) error {
	args := []string{"-h", host, "-p", strconv.FormatInt(int64(port), 10), "-c", "cluster", "forget", nodeID}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("forget node")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish forget node")

	return nil
}

func (cli *CLI) DeleteNode(ctx context.Context, host string, port int, nodeID string) error {
	args := []string{"-h", host, "-p", strconv.FormatInt(int64(port), 10), "--cluster", "del-node", host + ":" + strconv.FormatInt(int64(port), 10), nodeID}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("delete node")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish delete node")

	return nil
}

func (cli *CLI) ReplicateNode(ctx context.Context, host string, port int, masterNodeID string) error {
	args := []string{"-h", host, "-p", strconv.FormatInt(int64(port), 10), "-c", "cluster", "replicate", masterNodeID}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("replicate node")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish replicate node")

	return nil
}

func (cli *CLI) Rebalance(ctx context.Context, host string, port int) error {
	args := []string{"-h", host, "-p", strconv.FormatInt(int64(port), 10), "--cluster", "rebalance", host + ":" + strconv.FormatInt(int64(port), 10)}

	log.Info().Str("command", string(cli.name)).Strs("args", args).Msg("rebalance")

	if cli.password != "" {
		args = append(args, "-a", cli.password)
	}

	cmd := exec.CommandContext(ctx, string(cli.name), args...)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run command %s %v: %w", cli.name, args, err)
	}

	log.Info().Msg("finish rebalance")

	return nil
}

func (cli *CLI) ExceptNode(ctx context.Context, host string, port int, exceptionNode string) error {
	nodes, err := cli.GetClusterNodes(ctx, host, port)
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	var underBorderSlot, overBorderSlot int
	var slotCount int
firstLoop:
	for _, node := range nodes {
		if len(node.Slots) == 0 {
			continue firstLoop
		}

		for _, flag := range node.Flags {
			if flag == "slave" {
				continue firstLoop
			}
		}

		if node.ID == exceptionNode {
			underBorderSlot = node.Slots[0]
			overBorderSlot = node.Slots[len(node.Slots)-1]
			slotCount = node.Slots[len(node.Slots)-1] - node.Slots[0] + 1
		}
	}

	var underBorderID, overBorderID string
secondLoop:
	for _, node := range nodes {
		if len(node.Slots) == 0 {
			continue secondLoop
		}

		for _, flag := range node.Flags {
			if flag == "slave" {
				continue secondLoop
			}
		}

		if node.Slots[0] == overBorderSlot+1 {
			overBorderID = node.ID
		}

		if node.Slots[len(node.Slots)-1] == underBorderSlot-1 {
			underBorderID = node.ID
		}
	}

	if underBorderID == "" || overBorderID == "" {
		return fmt.Errorf("failed to find underBorderID or overBorderID")
	}

	selectedBorderNode := ""
	if underBorderID == "" {
		selectedBorderNode = overBorderID
	}
	if overBorderID == "" {
		selectedBorderNode = underBorderID
	}
	if underBorderID != "" && overBorderID != "" {
		selectedBorderNode = [2]string{underBorderID, overBorderID}[rand.Intn(2)]
	}

	log.Info().Str("selectedBorderNode", selectedBorderNode).Int("slotCount", slotCount).Msg("except node")

	if err := cli.Reshard(ctx, host, port, selectedBorderNode, slotCount, exceptionNode); err != nil {
		return fmt.Errorf("failed to reshard: %w", err)
	}

	return nil
}

func (cli *CLI) MergeNode(ctx context.Context, host string, port int, targetNodeID string, sourceNodeID string) error {
	nodes, err := cli.GetClusterNodes(ctx, host, port)
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	var targetNode, sourceNode *ClusterNode
	for _, node := range nodes {
		if node.ID == targetNodeID {
			targetNode = node
		}
		if node.ID == sourceNodeID {
			sourceNode = node
		}
	}

	if targetNode == nil || sourceNode == nil {
		return fmt.Errorf("failed to find targetNode or sourceNode")
	}

	log.Info().Str("targetNode", targetNode.ID).Str("sourceNode", sourceNode.ID).Msg("merge node")

	slotCount := 0
	for i := 0; i < len(sourceNode.Slots); i += 2 {
		start := sourceNode.Slots[i]
		end := sourceNode.Slots[i+1]
		slotCount += end - start + 1
	}

	if err := cli.Reshard(ctx, host, port, targetNode.ID, slotCount, sourceNode.ID); err != nil {
		return fmt.Errorf("failed to reshard: %w", err)
	}

	return nil
}
