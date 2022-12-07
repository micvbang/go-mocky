package adder

import "fmt"

// DummyAddUser is a dummy function which uses the Adder interface that we
// wish to mock.
func DummyAddUser(a Adder) error {
	err := a.Add()
	if err != nil {
		return fmt.Errorf("failed to add: %w", err)
	}

	return nil
}
