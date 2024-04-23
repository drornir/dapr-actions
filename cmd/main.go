package main

import (
	"fmt"

	"github.com/drornir/dapr-actions/orch"
)

func main() {
	orchestrator := &orch.Application{}

	_ = orchestrator

	fmt.Printf("%#v\n", orchestrator)
}
