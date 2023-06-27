package functionalsupport_test

import (
	"github.com/hexa-org/policy-orchestrator/pkg/functionalsupport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/slices"
	"testing"
)

func TestSortCompact_Nil(t *testing.T) {
	sorted := functionalsupport.SortCompact(nil)
	assert.NotNil(t, sorted)
	assert.Empty(t, sorted)
}

func TestSortCompact_AllWhitespace(t *testing.T) {
	sorted := functionalsupport.SortCompact([]string{"", "", " ", "  ", "", " "})
	assert.NotNil(t, sorted)
	assert.Empty(t, sorted)
}

func TestSortCompact(t *testing.T) {
	arr := []string{"hello", "", "how", "are", " ", "you", "hello", "   ", "hello", "", "how", "are", "you", " "}
	sorted := functionalsupport.SortCompact(arr)
	assert.Equal(t, []string{"are", "hello", "how", "you"}, sorted)
}

func TestDiffUnique_NoDifference(t *testing.T) {
	arr1 := []string{"a", "b", "c", "", "  "}
	arr2 := []string{"a", "b", "c"}
	diff1, intersection, diff2 := functionalsupport.DiffUnique(arr1, arr2)
	assert.NotNil(t, diff1)
	assert.Empty(t, diff1)
	assert.NotNil(t, intersection)
	slices.Sort(intersection)
	assert.Equal(t, []string{"a", "b", "c"}, intersection)
	assert.NotNil(t, diff2)
	assert.Empty(t, diff2)
}

func TestDiffUnique_AllDifferent(t *testing.T) {
	arr1 := []string{"a", "b", "c"}
	arr2 := []string{"   ", "d", "", "e", "f"}
	diff1, intersection, diff2 := functionalsupport.DiffUnique(arr1, arr2)
	slices.Sort(diff1)
	slices.Sort(diff2)

	assert.NotNil(t, diff1)
	assert.Equal(t, []string{"a", "b", "c"}, diff1)

	assert.NotNil(t, intersection)
	assert.Empty(t, intersection)

	assert.NotNil(t, diff2)
	assert.Equal(t, []string{"d", "e", "f"}, diff2)
}

func TestDiffUnique_IntersectionAndDiff(t *testing.T) {
	arr1 := []string{"val1", "b", "  ", "c", "", "d", "e", "val2"}
	arr2 := []string{"   ", "d", "", "e", "val3", "b", "c", "val4"}
	diff1, intersection, diff2 := functionalsupport.DiffUnique(arr1, arr2)
	slices.Sort(diff1)
	slices.Sort(intersection)
	slices.Sort(diff2)
	assert.Equal(t, []string{"val1", "val2"}, diff1)
	assert.Equal(t, []string{"b", "c", "d", "e"}, intersection)
	assert.Equal(t, []string{"val3", "val4"}, diff2)
}

func TestDiffUnique_EmptyInputs(t *testing.T) {
	diff1, intersection, diff2 := functionalsupport.DiffUnique([]string{}, []string{})
	assert.NotNil(t, diff1)
	assert.Empty(t, diff1)

	assert.NotNil(t, intersection)
	assert.Empty(t, intersection)

	assert.NotNil(t, diff2)
	assert.Empty(t, diff2)
}
