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
	completed    int
	notcompleted int
}

func (n *NoopWorkFinder) MarkErroneous(task interface{}) {
	n.notcompleted = n.notcompleted + 1
	log.Printf("completed %v tasks\n", n.completed+1)
}

func (n *NoopWorkFinder) MarkCompleted(task interface{}) {
	n.completed = n.completed + 1
	log.Printf("completed %v tasks\n", n.completed+1)
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

	for finder.completed < 3 {
	}
	scheduler.Stop()
	assert.Equal(t, finder.completed, 3)
	assert.Equal(t, finder.notcompleted, 0)
}

func TestErroneousWorkflow(t *testing.T) {
	var worker ErroneousWorker
	var finder NoopWorkFinder

	list := []workflowsupport.Worker{&worker}
	scheduler := workflowsupport.NewScheduler(&finder, list, 50)
	scheduler.Start()

	for finder.notcompleted < 3 {
	}
	scheduler.Stop()
	assert.Equal(t, finder.completed, 0)
	assert.Equal(t, finder.notcompleted, 3)
}
