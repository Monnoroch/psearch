package errors

import (
	goerrors "errors"
	"runtime/debug"
)

type ErrorT struct {
	Base      error
	callstack string
}

func (self ErrorT) Error() string {
	return self.Base.Error() + "\n" + self.callstack
}

func (self ErrorT) Reason() error {
	return self.Base
}

func New(s string) error {
	return ErrorT{goerrors.New(s), string(debug.Stack())}
}

func NewErr(err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(ErrorT); ok {
		return err
	}
	return ErrorT{err, string(debug.Stack())}
}
