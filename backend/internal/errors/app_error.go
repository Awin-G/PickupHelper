package errors

import (
	"fmt"
)

// AppError is the unified error type carrying business code, message,
// HTTP status, and an optional wrapped cause. Handlers should pass it
// to handler.Error which inspects it via errors.As.
type AppError struct {
	Code       int
	Msg        string
	HTTPStatus int
	Cause      error
}

// Error formats the error as "[code=<Code>] <Msg>"; if Cause is non-nil
// the cause's message is appended after ": ".
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[code=%d] %s: %s", e.Code, e.Msg, e.Cause.Error())
	}
	return fmt.Sprintf("[code=%d] %s", e.Code, e.Msg)
}

// Unwrap returns the wrapped cause, enabling errors.Is / errors.As.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// New returns an AppError with the given code and msg. If msg is empty,
// the default message for the code is used (see Msg).
func New(code int, msg string) *AppError {
	if msg == "" {
		msg = Msg(code)
	}
	return &AppError{
		Code:       code,
		Msg:        msg,
		HTTPStatus: HTTPStatus(code),
	}
}

// Wrap returns an AppError that wraps cause with the given code and msg.
func Wrap(err error, code int, msg string) *AppError {
	if msg == "" {
		msg = Msg(code)
	}
	return &AppError{
		Code:       code,
		Msg:        msg,
		HTTPStatus: HTTPStatus(code),
		Cause:      err,
	}
}

// Convenience constructors for the most common generic error codes.

func InvalidParam(msg string) *AppError { return New(ErrInvalidParam, msg) }
func Unauthenticated(msg string) *AppError {
	return New(ErrUnauthenticated, msg)
}
func Forbidden(msg string) *AppError   { return New(ErrForbidden, msg) }
func NotFound(msg string) *AppError    { return New(ErrNotFound, msg) }
func Conflict(msg string) *AppError    { return New(ErrConflict, msg) }
func Internal(msg string) *AppError    { return New(ErrInternal, msg) }
func RateLimited(msg string) *AppError { return New(ErrRateLimited, msg) }
