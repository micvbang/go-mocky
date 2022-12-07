# go-mocky

`go-mocky` makes it easy to generate fakes/mocks from Go interfaces.

`go-mocky` is heavily inspired by <https://github.com/vektra/mockery>.

## Goals

- generate code that does not rely on `{}interface`s at all, making it possible for our tooling to help us catch bugs
- keep the generated code (and this library!) as simple as possible, in an attempt to make maintenance of both as little work as possible
- make it easy to generate mocks using `go generate`

## Usage and Example

We will generate mocks for the interface `Adder` below by instructing [go generate](https://go.dev/blog/generate) to run `mocky` with argument `-i Adder`, the name of the interface.

```go
//go:generate mocky -i Adder

type Adder interface {
	Add() error
}
```

```bash
$ go generate ./...

writing to file /home/micvbang/go-mocky/examples/adder/mock_adder.go
```

The generated code looks like this:

```go
type MockAdder struct {
	T testing.TB

	AddMock  func() error
	AddCalls []adderAddCall
}

func NewMockAdder(t testing.TB) *MockAdder {
	return &MockAdder{T: t}
}

type adderAddCall struct {
	Out0 error
}

func (_v *MockAdder) Add() error {
	if _v.AddMock == nil {
		msg := fmt.Sprintf("call to %T.Add, but MockAdd is not set", _v)
		panic(msg)
	}

	_v.AddCalls = append(_v.AddCalls, adderAddCall{})
	out0 := _v.AddMock()
	_v.AddCalls[len(_v.AddCalls)-1].Out0 = out0
	return out0
}
```

That's it! We're now ready to mock Adder.

Below there's a silly function called `DummyAddUser` which does nothing useful except returning an error when `Add()` fails.

```go
// DummyAddUser is a dummy function which uses the Adder interface that we
// wish to mock.
func DummyAddUser(a Adder) error {
	err := a.Add()
	if err != nil {
		return fmt.Errorf("failed to add: %w", err)
	}

	return nil
}
```

Since we have generated a mock for the `Adder` interface, we can easily write a test that verifies that it does indeed return the errors returned by `Add`.

```go

// TestReturnAddErrorErrors verifies that GiveValueToAdder returns an error
// when Add fails, and returns nil Add does not fail.
func TestReturnAddErrorErrors(t *testing.T) {
	expectedErr := fmt.Errorf("oh no, all is on fire!")

	// Here we're defining the mock and its functionality.
	mockAdder := adder.NewMockAdder(t)
	mockAdder.AddMock = func() error {
		return expectedErr
	}

	// Verify that ReturnAddError returns the error returned by Add
	got := adder.ReturnAddError(mockAdder)
	require.ErrorIs(t, got, expectedErr)

	// If we want to be fancy, we can even see the history of the calls made
	// to Add, including which arguments were given and which values were returned:
	require.Equal(t, 1, len(mockAdder.AddCalls))

	// Add has no arguments, but if it had had, they would be available on `call` as well
	call := mockAdder.AddCalls[0]

	// We can check that the call to Add indeed contains the expectedErr.
	require.ErrorIs(t, call.Out0, expectedErr)
}
```

The test above shows how simple it is to define a mock/fake implementation of `Add()` to be used in our test.

If you're into tracking the number of calls (or the arguments/return values), `go-mocky` helps you by keeping a log of these in the `AddCalls` slice.

## License

This project is Apache 2 licensed. See LICENSE for the full license text.
