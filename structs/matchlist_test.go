package structs

import (
	"fmt"
	"testing"
)

// Registers an error and prints the `msg` if `cond` is false.
func failIf(cond bool, t *testing.T, msg string) {
	if cond {
		t.Error(msg)
	}
}

// Test that adding and retrieving numbers works.
func TestBasicAdd(t *testing.T) {
	idl := NewIDList()

	idl.Add(12345)
	id, real := idl.Next()

	failIf(id != 12345, t, "got unadded value back")
	failIf(!real, t, "got unexpected item back")
}

// Ensure we get numbers back in the same order we added them.
func TestRetrievalOrder(t *testing.T) {
	idl := NewIDList()

	var i RiotID
	for i = 0; i < 100; i++ {
		idl.Add(i)
	}

	for i = 0; i < 100; i++ {
		id, real := idl.Next()

		failIf(id != i, t, "got incorrect item back")
		failIf(!real, t, "queue was empty (?)")
	}

	failIf(idl.Available(), t, "items still available (bad)")
}

// Confirm that the right number of elements have been blacklisted.
func TestBlacklistCount(t *testing.T) {
	idl := NewIDList()
	addedCount := uint(0)

	for i := 0; i < 100; i++ {
		if idl.Add(RiotID(i)) {
			addedCount++
		}

		failIf(idl.blacklisted != addedCount, t, "some elements not added to list")
	}
}

// Ensure that grow() is triggered automatically and the correct number
// of times when blcap is exceeded. Also make sure the correct number of values
// are blacklisted despite the grow().
func TestGrowTrigger(t *testing.T) {
	idl := NewIDList()
	overflow := idl.blcap + 100
	addCount := uint(0)

	curr := 0
	for addCount < overflow {
		if idl.Add(RiotID(curr)) { // add
			addCount++
			idl.Next() // immediately remove
		}

		curr++
	}

	failIf(idl._growCount != 1, t, fmt.Sprintf("_growCount = %d", idl._growCount))
	failIf(idl.blacklisted != addCount, t, "incorrect blacklist count")
}

// Ensure that grow() operations are non-destructive and keep track of
// values from the smaller blacklist.
func TestGrowMerge(t *testing.T) {
	idl := NewIDList()

	for i := 0; i < int(idl.blcap)+10; i++ {
		idl.Add(RiotID(12345))

		// Should always be exactly 1; key edge case is after grow operation.
		failIf(idl.blacklisted != 1, t, "incorrect number of items blacklisted")
	}
}
