package testsupport_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/testsupport"
	"github.com/stretchr/testify/assert"
	"testing"
)

type TestData struct {
	data string
}

func (t *TestData) SetUp() {
	t.data = "aTest"
}

func (t *TestData) TearDown() {
}

func TestWithSetUp(t *testing.T) {
	testsupport.WithSetUp(&TestData{}, func(d *TestData) {
		assert.Equal(t, "aTest", d.data)
	})
}
