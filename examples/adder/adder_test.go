package adder_test

import (
	"fmt"
	"testing"

	"github.com/micvbang/go-mocky/examples/adder"
	"github.com/stretchr/testify/require"
)

// TestReturnAddErrorErrors verifies that ReturnAddError returns an error
// when Add fails, and returns nil Add does not fail.
func TestReturnAddErrorErrors(t *testing.T) {
	expectedErr := fmt.Errorf("oh no, all is on fire!")

	// Here we're defining the mock and its functionality.
	mockAdder := adder.NewMockAdder(t)
	mockAdder.AddMock = func() error {
		return expectedErr
	}

	// Verify that ReturnAddError returns the error returned by Add
	got := adder.DummyAddUser(mockAdder)
	require.ErrorIs(t, got, expectedErr)

	// If we want to be fancy, we can even see the history of the calls made
	// to Add, including which arguments were given and which values were returned:
	require.Equal(t, 1, len(mockAdder.AddCalls))

	// Add has no arguments, but if it had had, they would be available on `call` as well
	call := mockAdder.AddCalls[0]

	// We can check that the call to Add indeed contains the expectedErr.
	require.ErrorIs(t, call.Out0, expectedErr)
}
