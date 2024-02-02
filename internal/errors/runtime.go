package errors

import "errors"

var ErrInvalidRuntime = errors.New("given runtime is not valid")

func New(msg string) error {
	return errors.New(msg)
}
