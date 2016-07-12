package helpers

import (
	"fmt"
	"testing"
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

// CheckError tests if the error is not nil, and fails the test if so.
func CheckError(t *testing.T, e error) {
	if e != nil {
		t.Fatal("Not expecting error ", e)
	}
}

// PanicError tests if the error is not nil, and panics if so.
func PanicError(e error) {
	if e != nil {
		panic(fmt.Sprintf("Not expecting error %+v", e))
	}
}
