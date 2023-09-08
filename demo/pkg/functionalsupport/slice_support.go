package functionalsupport

import (
	"golang.org/x/exp/slices"
	"sort"
	"strings"
)

func DiffUnique(arr1, arr2 []string) (diff1 []string, intersection []string, diff2 []string) {
	sorted1 := SortCompact(arr1)
	sorted2 := SortCompact(arr2)

	map1 := make(map[string]bool)
	for _, ro := range sorted1 {
		if strings.TrimSpace(ro) != "" {
			map1[ro] = false
		}
	}

	diff1 = make([]string, 0)
	intersection = make([]string, 0)
	diff2 = make([]string, 0)

	//arr1 = [a, b]
	//arr2 = [a, c]
	// iter1 'a' - found=true, map1=[b], intersect=[a], diff2 =[]
	// iter2 'c' - found=false,map1=[b], intersect=[a], diff2=[c]

	for _, val2 := range sorted2 {
		if strings.TrimSpace(val2) == "" {
			continue
		}
		_, foundIn1 := map1[val2]
		if foundIn1 {
			map1[val2] = true
		} else {
			diff2 = append(diff2, val2)
		}
	}

	for ro, foundIn2 := range map1 {
		if foundIn2 {
			intersection = append(intersection, ro)
		} else {
			diff1 = append(diff1, ro)
		}
	}

	// diff1 - elements from arr1 that don't exist in arr2
	// intersection - elements that exist in both
	// diff2 = elements from arr2, that don't exist in arr1
	return diff1, intersection, diff2
}

// SortCompact - sorts slice, eliminates duplicates.
// does not remove whitespace
func SortCompact(arr []string) []string {
	sorted := make([]string, 0)
	sorted = append(sorted, arr...)
	// sort so duplicates are consecutive
	// compact eliminates consecutive dups
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] <= sorted[j]
	})
	compacted := slices.Compact(sorted)
	startIdx := 0
	for startIdx < len(compacted) {
		if strings.TrimSpace(compacted[startIdx]) != "" {
			break
		}
		startIdx++
	}
	return compacted[startIdx:]
}
