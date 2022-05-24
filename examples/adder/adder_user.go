package adder

import (
	"fmt"
)

func ReturnAddError(a Adder) error {
	err := a.Add()
	if err != nil {
		return fmt.Errorf("failed to add: %w", err)
	}

	return nil
}
