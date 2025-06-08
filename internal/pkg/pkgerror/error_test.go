package pkgerror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_NewError(t *testing.T) {
	tests := []struct {
		name            string
		errProvider     func() error
		matcherFunc     func(error) bool
		expectedBool    bool
		expectedMessage string
	}{
		{
			name: "validation error",
			errProvider: func() error {
				return NewValidationError("some error")
			},
			matcherFunc: func(err error) bool {
				return IsValidationError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "business error",
			errProvider: func() error {
				return NewBusinessError("some error")
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "server error",
			errProvider: func() error {
				return NewServerError("some error")
			},
			matcherFunc: func(err error) bool {
				return IsServerError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.errProvider()
			isMatch := test.matcherFunc(err)

			assert.Equal(t, test.expectedBool, isMatch)
			assert.Equal(t, test.expectedMessage, err.Error())
		})
	}
}

func Test_ErrorFrom(t *testing.T) {
	tests := []struct {
		name            string
		errProvider     func() error
		matcherFunc     func(error) bool
		expectedBool    bool
		expectedMessage string
	}{
		{
			name: "validation error",
			errProvider: func() error {
				return ValidationErrorFrom(errors.New("some error"))
			},
			matcherFunc: func(err error) bool {
				return IsValidationError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "false validation error",
			errProvider: func() error {
				return errors.New("some error")
			},
			matcherFunc: func(err error) bool {
				return IsValidationError(err)
			},
			expectedBool:    false,
			expectedMessage: "some error",
		},
		{
			name: "business error",
			errProvider: func() error {
				return BusinessErrorFrom(errors.New("some error"))
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "false business error",
			errProvider: func() error {
				return errors.New("some error")
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    false,
			expectedMessage: "some error",
		},
		{
			name: "business error code",
			errProvider: func() error {
				return NewBusinessErrorCode(Generic)
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "Error",
		},
		{
			name: "business error code with custom message",
			errProvider: func() error {
				return NewBusinessErrorCodeWithCustomMessage(Generic, "some error")
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "unknown business error code",
			errProvider: func() error {
				return NewBusinessErrorCode(-1)
			},
			matcherFunc: func(err error) bool {
				return IsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "unknown error",
		},
		{
			name: "server error",
			errProvider: func() error {
				return ServerErrorFrom(errors.New("some error"))
			},
			matcherFunc: func(err error) bool {
				return IsServerError(err)
			},
			expectedBool:    true,
			expectedMessage: "some error",
		},
		{
			name: "false server error",
			errProvider: func() error {
				return errors.New("some error")
			},
			matcherFunc: func(err error) bool {
				return IsServerError(err)
			},
			expectedBool:    false,
			expectedMessage: "some error",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.errProvider()
			isMatch := test.matcherFunc(err)

			assert.Equal(t, test.expectedBool, isMatch)
			assert.Equal(t, test.expectedMessage, err.Error())
		})
	}
}

func Test_TypeError_Unwrap(t *testing.T) {
	err := &Error{
		Message:   "",
		Original:  errors.New("some error"),
		ErrorType: ValidationError,
	}

	assert.Equal(t, errors.New("some error"), err.Unwrap())
}

func Test_Error_InnerMost(t *testing.T) {
	tests := []struct {
		name          string
		errProvider   func() *Error
		expectedError error
	}{
		{
			name: "no original error",
			errProvider: func() *Error {
				return &Error{
					Message:   "",
					Original:  nil,
					ErrorType: 0,
				}
			},
			expectedError: nil,
		},
		{
			name: "first wrap",
			errProvider: func() *Error {
				return &Error{
					Message:   "",
					Original:  errors.New("some error"),
					ErrorType: 0,
				}
			},
			expectedError: errors.New("some error"),
		},
		{
			name: "second wrap",
			errProvider: func() *Error {
				return &Error{
					Message: "",
					Original: &Error{
						Message:   "",
						Original:  errors.New("some other error"),
						ErrorType: 0,
					},
					ErrorType: 0,
				}
			},
			expectedError: errors.New("some other error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.errProvider()

			assert.Equal(t, test.expectedError, err.Innermost())
		})
	}
}

func Test_AsError(t *testing.T) {
	tests := []struct {
		name            string
		errProvider     func() error
		matcherFunc     func(error) (*Error, bool)
		expectedBool    bool
		expectedMessage string
	}{
		{
			name: "validation error",
			errProvider: func() error {
				return ValidationErrorFrom(assert.AnError)
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsValidationError(err)
			},
			expectedBool:    true,
			expectedMessage: "assert.AnError general error for testing",
		},
		{
			name: "false validation error",
			errProvider: func() error {
				return assert.AnError
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsValidationError(err)
			},
			expectedBool:    false,
			expectedMessage: "assert.AnError general error for testing",
		},
		{
			name: "business error",
			errProvider: func() error {
				return BusinessErrorFrom(assert.AnError)
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsBusinessError(err)
			},
			expectedBool:    true,
			expectedMessage: "assert.AnError general error for testing",
		},
		{
			name: "false business error",
			errProvider: func() error {
				return assert.AnError
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsBusinessError(err)
			},
			expectedBool:    false,
			expectedMessage: "assert.AnError general error for testing",
		},
		{
			name: "server error",
			errProvider: func() error {
				return ServerErrorFrom(assert.AnError)
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsServerError(err)
			},
			expectedBool:    true,
			expectedMessage: "assert.AnError general error for testing",
		},
		{
			name: "false server error",
			errProvider: func() error {
				return assert.AnError
			},
			matcherFunc: func(err error) (*Error, bool) {
				return AsServerError(err)
			},
			expectedBool:    false,
			expectedMessage: "assert.AnError general error for testing",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.errProvider()
			_, isMatch := test.matcherFunc(err)

			assert.Equal(t, test.expectedBool, isMatch)
			assert.Equal(t, test.expectedMessage, err.Error())
		})
	}
}
