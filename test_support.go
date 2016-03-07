// Licensed under the Apache License, Version 2.0
// Details: https://raw.githubusercontent.com/maniksurtani/quotaservice/master/LICENSE

package quotaservice

import (
	"fmt"
	"testing"
)

type MockEmitter struct{}

func (d *MockEmitter) Emit(e Event) {
	// No-op
}

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

