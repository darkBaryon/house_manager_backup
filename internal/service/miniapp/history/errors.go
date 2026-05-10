package history

import (
	"fmt"

	"house-manager/pkg/errcode"
)

func invalidParamf(format string, args ...any) error {
	return errcode.InvalidParam.WithError(fmt.Errorf(format, args...))
}

func notFoundf(format string, args ...any) error {
	return errcode.NotFound.WithError(fmt.Errorf(format, args...))
}

func databasef(format string, args ...any) error {
	return errcode.DatabaseError.WithError(fmt.Errorf(format, args...))
}
