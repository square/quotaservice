// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package test

import (
	"fmt"
	"testing"
)

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

