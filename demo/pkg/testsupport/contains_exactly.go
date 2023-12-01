package testsupport

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func ContainsExactly[T any](t *testing.T, arrayToCheck []T, elementsToCheck ...T) {
	assert.Equal(t, len(arrayToCheck), len(elementsToCheck))

	for _, element := range elementsToCheck {
		assert.Contains(t, arrayToCheck, element)
	}
}
