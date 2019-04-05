package redis

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
)

func TestIsRedisClientClosedError(t *testing.T) {
	tests := []struct {
		input        error
		isCloseError bool
	}{
		{
			// Test exactly the error
			input:        fmt.Errorf(redisClientClosedError),
			isCloseError: true,
		},
		{
			// Test the error wrapped
			input:        errors.Wrap(fmt.Errorf(redisClientClosedError), "obfuscate"),
			isCloseError: true,
		},
		{
			// test not the error
			input:        errors.New("just another error"),
			isCloseError: false,
		},
		{
			// test not the error wrapped with the text of the error (this should never happen)
			input:        errors.Wrap(errors.New("just another error"), redisClientClosedError),
			isCloseError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.input.Error(), func(t *testing.T) {
			result := isRedisClientClosedError(test.input)
			if result != test.isCloseError {
				t.Fatal("failed to detect error")
			}
		})
	}
}
