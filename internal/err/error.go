package err

import (
	"fmt"
	"strings"
)

// A Const error which means that the error can't be
// changed, unlike a var error.
type Const string

// Get the error message
func (c Const) Error() string {
	return string(c)
}

// Is c the same as thet (target)
func (c Const) Is(t error) bool {
	ts := t.Error()
	cs := c.Error()
	return ts == cs || strings.HasPrefix(ts, cs+": ")
}

// Wrap the inner error inside the const error
func (c Const) Wrap(inner error) error {
	return wrapError{msg: string(c), err: inner}
}

type wrapError struct {
	msg string
	err error
}

// Get the error message
func (w wrapError) Error() string {
	if w.err != nil {
		return fmt.Sprintf("%s: %v", w.msg, w.err)
	}
	return w.msg
}

// Unwrap the inner error
func (w wrapError) Unwrap() error {
	return w.err
}

// Is w the same as thet (target)
func (w wrapError) Is(target error) bool {
	return Const(w.msg).Is(target)
}
