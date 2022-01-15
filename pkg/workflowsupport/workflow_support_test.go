package workflowsupport_test

import (
	"errors"
	"github.com/hexa-org/policy-orchestrator/pkg/workflowsupport"
	"github.com/stretchr/testify/assert"
	"log"
	"sync/atomic"
	"testing"
)

type NoopWorker struct {
}

func (n *NoopWorker) Run(interface{}) error {
	log.Printf("doing work.\n")
	return nil
}

type NoopWorkFinder struct {
	completed    int32
	notcompleted int32
}

func (n *NoopWorkFinder) MarkErroneous() {
	addInt64 := atomic.AddInt32(&n.notcompleted, 1)
	log.Printf("notcompleted %v tasks\n", addInt64)
}

func (n *NoopWorkFinder) MarkCompleted() {
	addInt64 := atomic.AddInt32(&n.completed, 1)
	log.Printf("completed %v tasks\n", addInt64)
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
	var finder NoopWorkFinder

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()

	for atomic.LoadInt32(&finder.completed) < 3 {
	}
	scheduler.Stop()
	assert.True(t, atomic.LoadInt32(&finder.completed) > 2)
}

func TestErroneousWorkflow(t *testing.T) {
	var worker ErroneousWorker
	var finder NoopWorkFinder

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()

	for atomic.LoadInt32(&finder.notcompleted) < 3 {
	}
	scheduler.Stop()
	assert.True(t, atomic.LoadInt32(&finder.notcompleted) > 2)
}
