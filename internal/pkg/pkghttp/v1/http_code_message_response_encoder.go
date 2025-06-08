package pkghttp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/JoshuaPangaribuan/payslip/internal/pkg/pkgerror"
)

// This is the default code and message combination for our external API specification.
// Prefix 0x is reserved for successful responses.
// Prefix 14xx is reserved for error from the client side.
// Prefix 15xx is reserved for error from server side.
//
//nolint:gochecknoglobals // intended to be global
var (
	RequestSuccess    = CodeMessage{"00", "Request success"}
	RequestInProgress = CodeMessage{"01", "Request in progress"}
	RequestCompleted  = CodeMessage{"02", "Request completed"}

	RequestNotFound             = CodeMessage{"1401", "Endpoint not found"}
	RequestMethodNotAllowed     = CodeMessage{"1402", "Method endpoint not allowed"}
	RequestAuthenticationFailed = CodeMessage{"1403", "Authentication failed"}
	RequestValidationFailed     = CodeMessage{"1404", "Validation failed"}
	RequestInvalid              = CodeMessage{"1405", "Invalid request"}

	RequestGenericError = CodeMessage{"1500", "Unexpected error. Please contact support"}
)

// CodeMessageAware interface should be implemented by the response struct which desires custom code and message.
type CodeMessageAware interface {
	CodeMessage() CodeMessage
}

type CodeMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CodeMessageResponse struct {
	CodeMessage
	Data any `json:"data,omitempty"`
}

func (codeMessageResponse *CodeMessageResponse) UpdateCodeMessage(
	codeMessageAware CodeMessageAware,
) {
	codeMessageResponse.Code = codeMessageAware.CodeMessage().Code
	codeMessageResponse.Message = codeMessageAware.CodeMessage().Message
}

func DefaultCodeMessageResponse(response any) CodeMessageResponse {
	return NewCodeMessageResponse(RequestSuccess, response)
}

func NewCodeMessageResponse(codeMessage CodeMessage, response any) CodeMessageResponse {
	return CodeMessageResponse{
		CodeMessage: CodeMessage{
			Code:    codeMessage.Code,
			Message: codeMessage.Message,
		},
		Data: response,
	}
}

func DefaultErrorCodeMessageResponse() CodeMessageResponse {
	return NewCodeMessageResponse(RequestGenericError, nil)
}

// CodeMessageResponseEncoder function defines the contract for encoding successful responses with additional code and
// message. This additional field usually used by the external API hit by a partner to define clearer response.
func CodeMessageResponseEncoder(_ context.Context, w http.ResponseWriter, response any) error {
	code := writeHeaderAndStatusCode(w, response)

	if code == http.StatusNoContent {
		return nil
	}

	codeMessageResponse := DefaultCodeMessageResponse(response)

	if codeMessageAware, ok := response.(CodeMessageAware); ok {
		codeMessageResponse.UpdateCodeMessage(codeMessageAware)
	}

	return json.NewEncoder(w).Encode(codeMessageResponse)
}

// CodeMessageErrorEncoder function defines the contract for encoding error responses with additional
// code and message.
func CodeMessageErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set(contentType, applicationJSON)
	statusCode := http.StatusInternalServerError
	response := DefaultErrorCodeMessageResponse()

	if pkgerror.IsValidationError(err) {
		statusCode = http.StatusBadRequest
		response = CodeMessageResponse{
			CodeMessage: RequestValidationFailed,
			Data:        err.Error(),
		}
	}

	if pkgerror.IsBusinessError(err) {
		statusCode = http.StatusBadRequest
		response = CodeMessageResponse{
			CodeMessage: RequestInvalid,
			Data:        err.Error(),
		}
	}

	if pkgerror.IsServerError(err) {
		statusCode = http.StatusInternalServerError
		response = CodeMessageResponse{
			CodeMessage: RequestGenericError,
			Data:        err.Error(),
		}
	}

	w.WriteHeader(statusCode)

	_ = json.NewEncoder(w).Encode(response) //nolint:errcheck,errchkjson // won't be an error
}
