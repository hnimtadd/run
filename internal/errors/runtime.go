package errors

import (
	"errors"
	"fmt"
)

var ErrInvalidRuntime = errors.New("given runtime is not valid")

func New(msg string) error {
	return errors.New(msg)
}

func Newf(msg string, args ...any) error {
	return fmt.Errorf(msg, args...)
}
