package irbis

import (
	"errors"
	"fmt"
)

// Client-side network / transport failure (not an IRBIS server code).
const ErrCodeNetwork = -100000

// Error is an IRBIS or client-side failure with a numeric code.
type Error struct {
	Code    int
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Message == "" {
		e.Message = DescribeError(e.Code)
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// NewError builds an Error from an IRBIS return code.
func NewError(code int) *Error {
	return &Error{
		Code:    code,
		Message: DescribeError(code),
	}
}

// WrapError attaches a cause to an IRBIS/client error code.
func WrapError(code int, err error) *Error {
	return &Error{
		Code:    code,
		Message: DescribeError(code),
		Err:     err,
	}
}

// CodeOf returns the IRBIS/client code, or 0 if err is not *Error.
func CodeOf(err error) int {
	var e *Error
	if errors.As(err, &e) {
		return e.Code
	}
	return 0
}

// AsError extracts *Error from err.
func AsError(err error) (*Error, bool) {
	var e *Error
	if errors.As(err, &e) {
		return e, true
	}
	return nil, false
}
