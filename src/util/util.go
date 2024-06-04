package util

import (
	"errors"
	"fmt"
)

func Err_fmt(msg string, args ...any) error {
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	return errors.New(msg)
}

func Err_add(list *[]error, msg string, args ...any) {
	*list = append(*list, Err_fmt(msg, args...))
}

func Err_join(errs []error) error {
	return errors.Join(errs...)
}
