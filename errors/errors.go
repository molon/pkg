package errors

import (
	pkgerr "github.com/pkg/errors"
)

func Cause(err error) error {
	return pkgerr.Cause(err)
}

func New(message string) error {
	return pkgerr.New(message)
}

func Errorf(format string, args ...interface{}) error {
	return pkgerr.Errorf(format, args...)
}

func WithStack(err error) error {
	return pkgerr.WithStack(err)
}

func Wrap(err error, message string) error {
	return pkgerr.Wrap(err, message)
}

func Wrapf(err error, format string, args ...interface{}) error {
	return pkgerr.Wrapf(err, format, args...)
}

func WithMessage(err error, message string) error {
	return pkgerr.WithMessage(err, message)
}

func WithMessagef(err error, format string, args ...interface{}) error {
	return pkgerr.WithMessagef(err, format, args...)
}
