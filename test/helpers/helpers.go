package helpers

import (
	"testing"
	"fmt"
)

// ExpectingPanic indicates that a function passed in should panic. If it does, no errors are
// thrown. If not, the test fails.
func ExpectingPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("Did not panic()")
		} else {
			fmt.Print(r)
		}
	}()

	f()
}
