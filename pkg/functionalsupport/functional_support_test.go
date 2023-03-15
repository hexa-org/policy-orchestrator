package functionalsupport_test

import (
	"testing"

	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
	"github.com/stretchr/testify/assert"
)

type Record struct {
	Name string
}

func TestMap(t *testing.T) {
	records := []Record{{"foo"}, {"bar"}, {"baz"}}
	names := functionalsupport.Map(records, func(record Record) string {
		return record.Name
	})
	assert.Equal(t, []string{"foo", "bar", "baz"}, names)
}
