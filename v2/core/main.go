package main

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/v2/core/policyprovider"
)

func main() {
	fmt.Println("From core.Hello", policyprovider.Hello("saurabh"))
}
