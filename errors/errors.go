package errors

import (
	"fmt"

	pkgerr "github.com/pkg/errors"
)

func Cause(err error) error {
	return pkgerr.Cause(err)
}

func Wrap(err error) error {
	// 如果没包过，就包一下
	// 否则也不包了，保留起源的堆栈信息
	type causer interface {
		Cause() error
	}
	if _, ok := err.(causer); !ok {
		err = pkgerr.WithStack(err)
	}

	return err
}

func Wrapf(format string, a ...interface{}) error {
	return Wrap(fmt.Errorf(format, a...))
}

func WrapRecovery(p interface{}) error {
	perr, ok := p.(error)
	if !ok {
		perr = fmt.Errorf(fmt.Sprint(p))
	}
	return Wrap(perr)
}
