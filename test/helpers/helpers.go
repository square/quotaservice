package helpers

import (
	"fmt"
	"runtime"
	"testing"
)

// ExpectingPanic indicates that a function passed in should panic. If it does, no errors are
// thrown. If not, the test fails.
func ExpectingPanic(t *testing.T, f func()) {
	t.Helper()

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
	t.Helper()

	if e != nil {
		_, file, line, _ := runtime.Caller(1)
		t.Fatalf("Not expecting error %+v from (%s:%d)", e, file, line)
	}
}

// PanicError tests if the error is not nil, and panics if so.
func PanicError(e error) {
	if e != nil {
		panic(fmt.Sprintf("Not expecting error %+v", e))
	}
}

// ExpectingError checks that an error passed in is, in fact, an error. If not, it will fail the test.
func ExpectingError(t *testing.T, e error) {
	if e == nil {
		t.Fatal("Expecting error, got nil")
	}
}
