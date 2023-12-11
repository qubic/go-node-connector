# Qubic network SDK

This SDK is used to interact with the Qubic network and it currently supports fetching information related to identity balances, with ongoing development to add additional functionalities exposed by Qubic nodes.

### Basic usage

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/0xluk/go-qubic"
)

var nodeIP = "65.21.10.217"
var nodePort = "21841"

func main() {
	client, err := qubic.NewClient(nodeIP, nodePort)
	if err != nil {
		log.Fatalf("creating qubic sdk: err: %s", err.Error())
	}
	// releasing tcp connection related resources
	defer client.Close()

	res, err := client.GetBalance(context.Background(), "PKXGRCNOEEDLEGTLAZOSXMEYZIEDLGMSPNTJJJBHIBJISHFFYBBFDVGHRJQF")
	if err != nil {
		log.Fatalf("Getting identity info. err: %s", err.Error())
	}

	fmt.Println(res.Entity.IncomingAmount - res.Entity.OutgoingAmount)
}
```