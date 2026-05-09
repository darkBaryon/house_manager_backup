package hmd

import (
	"fmt"
	"strings"

	"house-manager/pkg/errcode"
)

func invalidParamf(format string, args ...any) error {
	return errcode.InvalidParam.WithError(fmt.Errorf(format, args...))
}

func notFoundf(format string, args ...any) error {
	return errcode.NotFound.WithError(fmt.Errorf(format, args...))
}

func alreadyExistsf(format string, args ...any) error {
	return errcode.AlreadyExists.WithError(fmt.Errorf(format, args...))
}

func databasef(format string, args ...any) error {
	return errcode.DatabaseError.WithError(fmt.Errorf(format, args...))
}

func mutationError(action string, err error) error {
	if err == nil {
		return nil
	}
	if errcode.FromError(err) != nil {
		return err
	}

	wrapped := fmt.Errorf("%s: %w", action, err)
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "required") ||
		strings.Contains(message, "invalid") ||
		strings.Contains(message, "not allowed") ||
		strings.Contains(message, "must be") ||
		strings.Contains(message, "is nil") {
		return errcode.InvalidParam.WithError(wrapped)
	}
	return errcode.DatabaseError.WithError(wrapped)
}
