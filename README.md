# Qubic network SDK

This node connector is used to interact with any Qubic node directly.

### Basic usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/qubic/go-node-connector"
)

var nodeIP = "65.21.10.217"
var nodePort = "21841"

func main() {
	client, err := qubic.NewClient(context.Background(), nodeIP, nodePort)
	if err != nil {
		log.Fatalf("creating qubic sdk: err: %s", err.Error())
	}
	// releasing tcp connection related resources
	defer client.Close()

	res, err := client.GetIdentity(context.Background(), "PKXGRCNOEEDLEGTLAZOSXMEYZIEDLGMSPNTJJJBHIBJISHFFYBBFDVGHRJQF")
	if err != nil {
		log.Fatalf("Getting identity info. err: %s", err.Error())
	}

	fmt.Println(res.Entity.IncomingAmount - res.Entity.OutgoingAmount)
}
```