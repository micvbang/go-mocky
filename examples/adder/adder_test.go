package adder_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/micvbang/go-mocky/examples/adder"
	"github.com/stretchr/testify/require"
)

// TestGiveValueToAdderErrors verifies that GiveValueToAdder returns an error
// when Add fails, and returns nil Add does not fail.
func TestGiveValueToAdderErrors(t *testing.T) {
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
			mockAdder.AddMock = func(ctx context.Context, value string) error {
				return test.expected
			}

			const addValue = "value"
			got := adder.GiveValueToAdder(mockAdder, addValue)
			require.ErrorIs(t, got, test.expected)

			// We can look into the calls that was made to Add, including
			// which arguments were given and which values were returned:
			require.Equal(t, 1, len(mockAdder.AddCalls))

			// Named argument "Value"
			require.Equal(t, mockAdder.AddCalls[0].Value, addValue)

			// Return value 0
			require.ErrorIs(t, mockAdder.AddCalls[0].Out0, test.expected)
		})
	}
}
