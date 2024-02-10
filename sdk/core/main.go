package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/sdk/core/policyprovider"
)

func main() {
	fmt.Println("From core.Hello", policyprovider.Hello("saurabh"))
}
