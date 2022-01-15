package workflowsupport_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

type NoopWorker struct {
}

func (n *NoopWorker) Run(interface{}) error {
	log.Printf("doing work.\n")
	return nil
}

type NoopWorkFinder struct {
	Results chan bool
}

func NewNoopWorkFinder() NoopWorkFinder {
	return NoopWorkFinder{Results: make(chan bool)}
}

func (n *NoopWorkFinder) MarkErroneous() {
	n.Results <- false
	log.Println("non completed task")
}

func (n *NoopWorkFinder) MarkCompleted() {
	n.Results <- true
	log.Println("completed task")
}

func (n *NoopWorkFinder) Stop() {
	close(n.Results)
}

func (n NoopWorkFinder) FindRequested() []interface{} {
	log.Printf("finding work.\n")

	return []interface{}{
		"someInfo",
		"someMoreInfo",
		"andSomeMoreInfo",
	}
}

type ErroneousWorker struct {
}

func (n *ErroneousWorker) Run(interface{}) error {
	log.Printf("doing work.\n")
	return errors.New("oops")
}

func TestWorkflow(t *testing.T) {
	var worker NoopWorker
	finder := NewNoopWorkFinder()

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()

	for i := 0; i < 3; i++ {
		assert.True(t, <-finder.Results)
	}

	scheduler.Stop()
}

func TestErroneousWorkflow(t *testing.T) {
	var worker ErroneousWorker
	finder := NewNoopWorkFinder()

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()

	for i := 0; i < 3; i++ {
		assert.False(t, <-finder.Results)
	}

	scheduler.Stop()
}
