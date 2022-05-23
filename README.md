# go-mocky

`go-mocky` makes it easy to generate fakes/mocks from Go interfaces. `go-mocky` is heavily inspired by <https://github.com/vektra/mockery>.

## Goals

- generate code that does not rely on `{}interface`s at all, making it possible for our tooling to help us catch bugs
- keep the generated code (and this library!) as simple as possible, in an attempt to make maintenance of both as little work as possible
- make it easy to generate mocks using `go generate`

## Example

The interface `Adder` is given below. Using `go-mocky`, we can generate code that makes it easy to fake/mock `Adder`.

```go
type Adder interface {
    Add(ctx context.Context, value string) error
}
```

The generated code looks like this:

```go
type MockAdder struct {
	T testing.TB

	MockAdd  func(ctx context.Context, value string) error
	AddCalls []AddCall
}

func NewMockAdder(t testing.TB) *MockAdder {
	return &MockAdder{T: t}
}

type AddCall struct {
	In0 context.Context
	In1 string

	Out0 error
}

func (v *MockAdder) Add(ctx context.Context, value string) error {
	if v.MockAdd == nil {
		msg := "call to Add, but MockAdd is not set"
		if v.T == nil {
			panic(msg)
		}
		require.Fail(v.T, msg)
	}

	v.AddCalls = append(v.AddCalls, AddCall{
		In0: ctx,
		In1: value,
	})
	out0 := v.MockAdd(ctx, value)
	v.AddCalls[len(v.AddCalls)-1].Out0 = out0
	return out0
}
```

Given the function `GiveValueToAdder` below, which we want to test, we could write a test:

```go
func GiveValueToAdder(a Adder, v string) error {
	ctx := context.Background()
	err := a.Add(ctx, v)
	if err != nil {
		return fmt.Errorf("failed to add: %w", err)
	}

	return nil
}

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
			mockAdder.MockAdd = func(ctx context.Context, value string) error {
				return test.expected
			}

			const addValue = "value"
			got := adder.GiveValueToAdder(mockAdder, addValue)
			require.ErrorIs(t, got, test.expected)

			// We can look into the calls that was made to Add, including
			// which arguments were given and which values were returned:
			require.Equal(t, 1, len(mockAdder.AddCalls))

			// Argument 1
			require.Equal(t, mockAdder.AddCalls[0].In1, addValue)

			// Return value 0
			require.ErrorIs(t, mockAdder.AddCalls[0].Out0, test.expected)
		})
	}
}
```

The test shows how simple it is to define a mock/fake implementation of `Add()` to be used in our test.

If you're into tracking the number of calls (or the arguments/return values), `go-mocky` helps you by keeping a log of these in the `AddCalls` slice.
