// Package pkgerror provides convenience error wrapper for our error to make it easier to identify the source of
// errors.
package pkgerror

import (
	"encoding/json"
	"errors"
	"net/http"
)

type ErrorType int

const (
	// ValidationError should indicate an error related to request validation.
	ValidationError ErrorType = iota

	// BusinessError should represent an error related to a false business process.
	BusinessError

	// ServerError should represent an error related to technical issues that occurred on the server.
	ServerError

	// PartnerError should represent an error related to error from partner side
	PartnerError
)

// Error represents wrapped errors which can be differentiated from the error type whether it's validation,
// business, or server errors.
type Error struct {
	// Message of the error will be displayed if the original error is nil.
	Message string

	// Original is the original error that wrapped by this Error.
	Original error

	// ErrorType will refer what is the current error type.
	ErrorType ErrorType

	// Code is the standardized business error code that can be used across domains.
	Code Code

	// ResponseCode is error response code from partner
	ResponseCode string
}

// Error will show the wrapped original error Error() if the original is not nil, otherwise will show the Error's
// message.
func (err *Error) Error() string {
	if err.Original != nil {
		return err.Original.Error()
	}

	return err.Message
}

// Unwrap will return the original error as stated in Wrapper interface.
func (err *Error) Unwrap() error {
	return err.Original
}

// Innermost get the innermost-wrapped error.
func (err *Error) Innermost() error {
	innerMost := false
	current := err.Original

	for !innerMost {
		wrapped := errors.Unwrap(current)

		if wrapped != nil {
			current = wrapped

			continue
		}

		innerMost = true
	}

	return current
}

func (err *Error) isValidationError() bool {
	return err.ErrorType == ValidationError
}

func (err *Error) isBusinessError() bool {
	return err.ErrorType == BusinessError
}

func (err *Error) isServerError() bool {
	return err.ErrorType == ServerError
}
func (err *Error) isPartnerError() bool {
	return err.ErrorType == PartnerError
}

func NewValidationError(message string) error {
	return &Error{
		Message:   message,
		ErrorType: ValidationError,
	}
}

func ValidationErrorFrom(err error) *Error {
	return &Error{
		Message:   "",
		Original:  err,
		ErrorType: ValidationError,
	}
}

func IsValidationError(err error) bool {
	var eval *Error

	if errors.As(err, &eval) {
		return eval.isValidationError()
	}

	return false
}

func NewBusinessError(message string) *Error {
	return &Error{
		Message:   message,
		ErrorType: BusinessError,
	}
}

func NewBusinessErrorCode(code Code) *Error {
	return &Error{
		Message:   code.String(),
		ErrorType: BusinessError,
		Code:      code,
	}
}

func NewBusinessErrorCodeWithCustomMessage(code Code, customMessage string) *Error {
	return &Error{
		Message:   customMessage,
		ErrorType: BusinessError,
		Code:      code,
	}
}

func BusinessErrorFrom(err error) *Error {
	return &Error{
		Message:   "",
		Original:  err,
		ErrorType: BusinessError,
	}
}

func IsBusinessError(err error) bool {
	var eval *Error

	if errors.As(err, &eval) {
		return eval.isBusinessError()
	}

	return false
}

func NewServerError(message string) *Error {
	return &Error{
		Message:   message,
		ErrorType: ServerError,
	}
}

func ServerErrorFrom(err error) *Error {
	return &Error{
		Message:   "",
		Original:  err,
		ErrorType: ServerError,
	}
}

func IsServerError(err error) bool {
	var eval *Error

	if errors.As(err, &eval) {
		return eval.isServerError()
	}

	return false
}

func NewPartnerError(responseCode string, message string) *Error {
	return &Error{
		Message:      message,
		ErrorType:    PartnerError,
		ResponseCode: responseCode,
	}
}

func AsValidationError(err error) (*Error, bool) {
	var eval *Error

	if errors.As(err, &eval) {
		return eval, eval.isValidationError()
	}

	return &Error{}, false
}

func AsBusinessError(err error) (*Error, bool) {
	var eval *Error

	if errors.As(err, &eval) {
		return eval, eval.isBusinessError()
	}

	return &Error{}, false
}

func AsServerError(err error) (*Error, bool) {
	var eval *Error

	if errors.As(err, &eval) {
		return eval, eval.isServerError()
	}

	return &Error{}, false
}

func AsPartnerError(err error) (*Error, bool) {
	var eval *Error

	if errors.As(err, &eval) {
		return eval, eval.isPartnerError()
	}

	return &Error{}, false
}

// WriteError writes an error response to the HTTP response writer
func WriteError(w http.ResponseWriter, err Error) {
	w.Header().Set("Content-Type", "application/json")

	// Set appropriate status code based on error type
	statusCode := http.StatusInternalServerError
	switch err.ErrorType {
	case ValidationError:
		statusCode = http.StatusBadRequest
	case BusinessError:
		statusCode = http.StatusForbidden
	case PartnerError:
		statusCode = http.StatusBadGateway
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(err.Message)
}
