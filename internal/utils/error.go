package utils

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

var (
	ErrAuth               = &sentinelAPIError{code: http.StatusUnauthorized, msg: "invalid token"}
	ErrNotFound           = &sentinelAPIError{code: http.StatusNotFound, msg: "not found"}
	ErrInvalidCredentials = &sentinelAPIError{code: http.StatusBadRequest, msg: "invalid credentials provided"}
	ErrDuplicate          = &sentinelAPIError{code: http.StatusBadRequest, msg: "duplicate"}
	ErrBadRequest         = &sentinelAPIError{code: http.StatusBadRequest, msg: "bad request"}
	ErrInternal           = &sentinelAPIError{code: http.StatusInternalServerError, msg: "internal error"}
)

type APIError interface {
	// APIError returns StatusError HTTP status code and an API-safe error msg.
	APIError() StatusError
}

// StatusError holds HTTP status code, an API-safe error msg, and caller details.
type StatusError struct {
	Code   int
	Msg    string
	Caller string
}

// sentinel is a computer prog. pattern of using specific value
// to signify that no further processing is possible
type sentinelAPIError struct {
	code int
	msg  string
}

func (e sentinelAPIError) Error() string {
	return e.msg
}

func (e sentinelAPIError) APIError() StatusError {
	return StatusError{Code: e.code, Msg: e.msg}
}

// this helps associates errors from elsewhere in the application
// with one of the predefined sentinel errors above
type sentinelWrappedError struct {
	error
	caller   string
	sentinel *sentinelAPIError
}

// by implementing this interface: `Is(error) bool`, we can
// compare wrapped errors with the embedded sentinel errors
func (e sentinelWrappedError) Is(err error) bool {
	return e.sentinel == err
}

func (e sentinelWrappedError) APIError() StatusError {
	se := e.sentinel.APIError()
	se.Caller = e.caller
	return se
}

func wrapError(err error, sentinel *sentinelAPIError) error {
	pc, _, line, _ := runtime.Caller(1)
	details := runtime.FuncForPC(pc)

	return sentinelWrappedError{
		error:    err,
		caller:   fmt.Sprintf("%s#%d", details.Name(), line),
		sentinel: sentinel,
	}
}

func ErrorBadRequest(err error) error {
	return wrapError(err, ErrBadRequest)
}

func ErrorAuth(err error) error {
	return wrapError(err, ErrAuth)
}

func ErrorNotFound(err error) error {
	return wrapError(err, ErrNotFound)
}

func ErrorDuplicate(err error) error {
	return wrapError(err, ErrDuplicate)
}

func ErrorInvalidCredentials(err error) error {
	return wrapError(err, ErrInvalidCredentials)
}

func ErrorInvalidUserInput(err error, m map[string]string) error {
	var b strings.Builder
	for k, v := range m {
		fmt.Fprintf(&b, "%s: %s, ", k, v)
	}
	message := b.String()
	if len(message) >= 2 {
		message = message[:len(message)-2]
	}
	return wrapError(err, &sentinelAPIError{
		code: http.StatusBadRequest,
		msg:  message,
	})
}

func ErrorInternal(err error) error {
	return wrapError(err, ErrInternal)
}

func WrapDbError(err error, sentinel *sentinelAPIError) error {
	return wrapError(err, sentinel)
}
