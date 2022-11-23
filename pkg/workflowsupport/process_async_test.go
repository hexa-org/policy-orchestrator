package workflowsupport

import (
	"fmt"
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"testing"
)

func TestProcessAsync(t *testing.T) {
	things := []string{"thing", "anotherThing", "evenMoreThings", "lotsOfThings", "whyDoWeNeedMoreThings"}

	responses := ProcessAsync[string, string](things, func(thing string) (string, error) {
		return fmt.Sprintf("processed:%s", thing), nil
	})

	testsupport.ContainsExactly(t, responses,
		"processed:thing",
		"processed:anotherThing",
		"processed:evenMoreThings",
		"processed:lotsOfThings",
		"processed:whyDoWeNeedMoreThings",
	)
}
