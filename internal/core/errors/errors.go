package errors

import (
	stderrors "errors"
	"fmt"
)

type Code string

const (
	CodeInvalidArgument  Code = "invalid_argument"
	CodeMissingConfig    Code = "missing_configuration"
	CodeUnauthenticated  Code = "unauthenticated"
	CodePermissionDenied Code = "permission_denied"
	CodeNotFound         Code = "not_found"
	CodeConflict         Code = "conflict"
	CodeUnavailable      Code = "unavailable"
	CodeInternal         Code = "internal"
)

type Error struct {
	Code    Code
	Op      string
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	switch {
	case e.Op != "" && e.Message != "":
		return fmt.Sprintf("%s: %s", e.Op, e.Message)
	case e.Message != "":
		return e.Message
	case e.Op != "":
		return e.Op
	default:
		return string(e.Code)
	}
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(code Code, op, message string) error {
	return &Error{Code: code, Op: op, Message: message}
}

func Wrap(code Code, op, message string, err error) error {
	if err == nil {
		return New(code, op, message)
	}
	return &Error{Code: code, Op: op, Message: message, Err: err}
}

func MissingConfiguration(field string) error {
	return New(CodeMissingConfig, "config.validate", fmt.Sprintf("%s is required", field))
}

func CodeOf(err error) Code {
	var appErr *Error
	if stderrors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}

func IsCode(err error, code Code) bool {
	return CodeOf(err) == code
}
