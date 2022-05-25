package adder_test

import (
	"fmt"
	"testing"

	"github.com/micvbang/go-mocky/examples/adder"
	"github.com/stretchr/testify/require"
)

// TestReturnAddErrorErrors verifies that GiveValueToAdder returns an error
// when Add fails, and returns nil Add does not fail.
func TestReturnAddErrorErrors(t *testing.T) {
	tests := map[string]struct {
		expected error
	}{
		"no error":      {expected: nil},
		"failed to add": {expected: fmt.Errorf("failed to add")},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			mockAdder := adder.NewMockAdder(t)
			// Here we're defining the functionality of the mock.
			mockAdder.AddMock = func() error {
				return test.expected
			}

			// Verify that ReturnAddError returns the error returned by Add
			got := adder.ReturnAddError(mockAdder)
			require.ErrorIs(t, got, test.expected)

			// If we want to be fancy, we can even see the history of the calls made
			// to Add, including which arguments were given and which values were returned:
			require.Equal(t, 1, len(mockAdder.AddCalls))

			// Add has no arguments, but if it had had, they would be available on `call` as well
			call := mockAdder.AddCalls[0]

			// Return value
			require.ErrorIs(t, call.Out0, test.expected)
		})
	}
}
