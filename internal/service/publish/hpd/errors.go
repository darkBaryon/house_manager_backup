package hpd

import (
	"fmt"

	"house-manager/pkg/errcode"
)

func databasef(format string, args ...any) error {
	return errcode.DatabaseError.WithError(fmt.Errorf(format, args...))
}
