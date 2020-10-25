package err

import (
	"fmt"
	"strings"
)

type Const string

func (c Const) Error() string {
	return string(c)
}

func (c Const) Is(t error) bool {
	ts := t.Error()
	cs := c.Error()
	return ts == cs || strings.HasPrefix(ts, cs+": ")
}

func (err Const) wrap(inner error) error {
	return wrapError{msg: string(err), err: inner}
}

type wrapError struct {
	msg string
	err error
}

func (w wrapError) Error() string {
	if w.err != nil {
		return fmt.Sprintf("%s: %v", w.msg, w.err)
	}
	return w.msg
}

func (err wrapError) Unwrap() error {
	return err.err
}
func (err wrapError) Is(target error) bool {
	return Const(err.msg).Is(target)
}
