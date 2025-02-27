# KeyCL

KeyCL is a simple valkey/redis manager for operating.

## Pre-requirements

- Go 1.24 or later
- redis-cli or valkey-cli

## Installation

```bash
go get github.com/snowmerak/keycl
```

## Usage

### init

```go
package main

import (
	"context"
	
	"github.com/snowmerak/keycl/lib/cli"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	c := cli.New(cli.Valkey)
	// c := cli.New(cli.Redis)
}
```

### create cluster

```go
if err := c.CreateCluster(ctx, 0, "127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003"); err != nil {
	panic(err)
}
```

### add node

```go
if err := c.AddNode(ctx, "127.0.0.1", 7005, "127.0.0.1", 7002); err != nil {
	panic(err)
}
```

### remove node

#### forget node

```go
if err := c.ForgetNode(ctx, "127.0.0.1", 7005, "2b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
    panic(err)
}
```

#### delete node

```go
if err := c.DeleteNode(ctx, "127.0.0.1", 7005, "2b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
    panic(err)
}
```

### reshard

#### reshard all empty node

```go
if err := c.ReshardAll(ctx, "127.0.0.1", 7001); err != nil {
	panic(err)
}
```

#### move slot to neighbor node

```go
if err := c.ExceptNode(ctx, "127.0.0.1", 7001, "2b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
	panic(err)
}
```

#### merge slot to other node

```go
if err := c.MergeNode(ctx, "127.0.0.1", 7001, "2b6a441e4cd32fe88ddb460338a76479e4875a6b", "4b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
    panic(err)
}
```

#### rebalance slot

```go
if err := c.Rebalance(ctx, "127.0.0.1", 7001); err != nil {
	panic(err)
}
```

### replicate node

```go
if err := c.ReplicateNode(ctx, "127.0.0.1", 7003, "2b6a441e4cd32fe88ddb460338a76479e4875a6b"); err != nil {
    panic(err)
}
```
