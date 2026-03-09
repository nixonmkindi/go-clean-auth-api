package apperror

import "fmt"

// Error is an application error with an HTTP status code and machine-readable code.
type Error struct {
	Code       string
	Message    string
	HTTPStatus int
	Details    interface{}
	Err        error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Err
}

func New(httpStatus int, code, message string, details interface{}, err error) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Details:    details,
		Err:        err,
	}
}
