package http

import (
	"PROJECT_NAME/pkg/errs"
	"errors"
	"fmt"
	"net/http"
	"reflect"

	"github.com/go-playground/validator/v10"
)

// Response is the standard API response structure
type Response struct {
	Success bool       `json:"success"`
	Message string     `json:"message,omitempty"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
}

// ErrorInfo contains detailed error information
type ErrorInfo struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Details     any    `json:"details,omitzero"`
	TotalErrors *int   `json:"total_errors,omitempty"`
}

// Default messages for common responses
const (
	okMessage             = "A solicitação foi processada com sucesso"
	acceptedMessage       = "A solicitação foi aceita para processamento"
	createdMessage        = "O recurso foi criado com sucesso"
	unauthorizedMessage   = "Acesso não autorizado"
	forbiddenMessage      = "Acesso proibido"
	invalidRequestMessage = "Corpo da solicitação inválido"
	internalErrorMessage  = "Ocorreu um erro interno do servidor"
)

// ResponseHandler handles all API responses (success and errors)
type ResponseHandler struct{}

// NewResponseHandler creates a new response handler
func NewResponseHandler() *ResponseHandler {
	return &ResponseHandler{}
}

// SuccessResponse creates a standardized success response
func (hdr *ResponseHandler) SuccessResponse(data any, message string) Response {
	return Response{
		Success: true,
		Message: message,
		Data:    data,
	}
}

// Ok sends a 200 OK response with optional custom message
func (hdr *ResponseHandler) Ok(ctx RequestContext, data any, message ...string) {
	msg := hdr.getMsgOrDefault(message, okMessage)
	ctx.JSON(http.StatusOK, hdr.SuccessResponse(data, msg))
}

// Accepted sends a 202 Accepted response with optional custom message
func (hdr *ResponseHandler) Accepted(ctx RequestContext, data any, message ...string) {
	msg := hdr.getMsgOrDefault(message, acceptedMessage)
	ctx.JSON(http.StatusAccepted, hdr.SuccessResponse(data, msg))
}

// Created sends a 201 Created response with optional custom message
func (hdr *ResponseHandler) Created(ctx RequestContext, data any, message ...string) {
	msg := hdr.getMsgOrDefault(message, createdMessage)
	ctx.JSON(http.StatusCreated, hdr.SuccessResponse(data, msg))
}

// ErrorResponse creates a standardized error response
func (hdr *ResponseHandler) ErrorResponse(code, message string, details any) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
}

// ErrorResponseWithTotal creates a standardized error response with total error count
func (hdr *ResponseHandler) ErrorResponseWithTotal(code, message string, details any, totalErrors int) Response {
	return Response{
		Success: false,
		Error: &ErrorInfo{
			Code:        code,
			Message:     message,
			Details:     details,
			TotalErrors: &totalErrors,
		},
	}
}

// InvalidRequest sends a 400 Bad Request response and aborts the request
func (hdr *ResponseHandler) InvalidRequest(ctx RequestContext, msg string, err ...error) {
	message := hdr.getMsgOrDefault([]string{msg}, invalidRequestMessage)
	ctx.JSON(http.StatusBadRequest,
		hdr.ErrorResponse(errs.ErrInvalidBodyFormat.Code, message, hdr.parseValidationErrors(err...)))
	ctx.Abort()
}

// Unauthorized sends a 401 Unauthorized response and aborts the request
func (hdr *ResponseHandler) Unauthorized(ctx RequestContext, msg string, err ...error) {
	message := hdr.getMsgOrDefault([]string{msg}, unauthorizedMessage)
	ctx.JSON(http.StatusUnauthorized, hdr.ErrorResponse(
		errs.ErrUnauthorized.Code,
		message,
		hdr.parseValidationErrors(err...),
	))
	ctx.Abort()
}

// Forbidden sends a 403 Forbidden response and aborts the request
func (hdr *ResponseHandler) Forbidden(ctx RequestContext, msg string, err ...error) {
	message := hdr.getMsgOrDefault([]string{msg}, forbiddenMessage)
	ctx.JSON(http.StatusForbidden,
		hdr.ErrorResponse(errs.ErrForbidden.Code, message, hdr.parseValidationErrors(err...)))
	ctx.Abort()
}

// InternalErr sends a 500 Internal Server Error response and aborts the request
func (hdr *ResponseHandler) InternalErr(ctx RequestContext, code string, err error) {
	if code == "" {
		code = "ERR_INTERNAL_SERVER"
	}

	ctx.JSON(http.StatusInternalServerError, hdr.ErrorResponse(
		code,
		http.StatusText(http.StatusInternalServerError),
		hdr.parseValidationErrors(err),
	))
	ctx.Abort()
}

// ServiceUnavailable sends a 503 service unavailable response and aborts the request
func (hdr *ResponseHandler) ServiceUnavailable(ctx RequestContext, code string, err error) {
	ctx.JSON(http.StatusServiceUnavailable, hdr.ErrorResponse(
		code,
		http.StatusText(http.StatusInternalServerError),
		hdr.parseValidationErrors(err),
	))
	ctx.Abort()
}

// Error processes error and sends appropriate HTTP responses
func (hdr *ResponseHandler) Error(ctx RequestContext, err error) {
	if err == nil {
		return
	}

	// Check if it's a ErrCollection first
	var errCollection *errs.ErrCollection
	if errors.As(err, &errCollection) {
		hdr.handleMultiError(ctx, errCollection)
		return
	}

	// Check if it's a Error
	var eError *errs.Error
	if errors.As(err, &eError) {
		hdr.handleError(ctx, eError)
		return
	}

	// Handle as generic error
	hdr.handleGenericError(ctx, err)
}

// handleError processes a single domain error
func (hdr *ResponseHandler) handleError(ctx RequestContext, err *errs.Error) {
	statusCode := hdr.getStatusCodeForErrorType(err.Type)

	details := make(map[string]any)
	if err.Details != nil {
		details = err.Details
	}

	if err.Cause != nil {
		details["causa"] = err.Cause.Error()
	}

	// Send response
	ctx.JSON(statusCode, hdr.ErrorResponse(err.Code, err.Message, details))
	ctx.Abort()
}

// handleMultiError processes multiple errors
func (hdr *ResponseHandler) handleMultiError(ctx RequestContext, errCollection *errs.ErrCollection) {
	statusCode := hdr.getStatusCodeForErrorType(errCollection.Err.Type)

	// Convert errors to detailed format
	errorDetails := hdr.formatMultipleErrors(errCollection)

	// Send response with total count
	ctx.JSON(statusCode, hdr.ErrorResponseWithTotal(
		errCollection.Err.Code,
		errCollection.Err.Message,
		errorDetails,
		errCollection.Count(),
	))
	ctx.Abort()
}

// handleGenericError processes non-domain errors
func (hdr *ResponseHandler) handleGenericError(ctx RequestContext, err error) {
	ctx.JSON(http.StatusInternalServerError, hdr.ErrorResponse(
		errs.ErrUnknown.Code,
		internalErrorMessage,
		map[string]string{"error": err.Error()},
	))
	ctx.Abort()
}

// getStatusCodeForErrorType maps domain error types to HTTP status codes
func (hdr *ResponseHandler) getStatusCodeForErrorType(errorType errs.ErrorType) int {
	switch errorType {
	case errs.ErrorTypeValidation,
		errs.ErrorTypeMissing,
		errs.ErrorTypeInvalid:
		return http.StatusBadRequest
	case errs.ErrorTypeNotFound:
		return http.StatusNotFound
	case errs.ErrorTypeConversion:
		return http.StatusUnprocessableEntity
	case errs.ErrorTypeCreation:
		return http.StatusInternalServerError
	case errs.ErrorTypeUnsupported:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// formatMultipleErrors formats multiple errors for response
func (hdr *ResponseHandler) formatMultipleErrors(errCollection *errs.ErrCollection) map[string]any {
	result := make(map[string]any)

	// Group errors by type
	errorsByType := make(map[errs.ErrorType][]map[string]any)
	var genericErrors []string

	for i, err := range errCollection.GetErrors() {
		var eError *errs.Error
		if errors.As(err, &eError) {
			errorInfo := map[string]any{
				"code":    eError.Code,
				"message": eError.Message,
				"index":   i,
			}

			if eError.Details != nil {
				errorInfo["details"] = eError.Details
			}

			errorsByType[eError.Type] = append(errorsByType[eError.Type], errorInfo)
		} else {
			genericErrors = append(genericErrors, err.Error())
		}
	}

	// Add grouped errors to result
	if len(errorsByType) > 0 {
		result["errors_by_type"] = errorsByType
	}

	// Add generic errors if any
	if len(genericErrors) > 0 {
		result["generic_errors"] = genericErrors
	}

	// Add summary
	result["summary"] = map[string]int{
		"total":       errCollection.Count(),
		"validation":  len(errorsByType[errs.ErrorTypeValidation]),
		"not_found":   len(errorsByType[errs.ErrorTypeNotFound]),
		"processing":  len(errorsByType[errs.ErrorTypeProcessing]),
		"conversion":  len(errorsByType[errs.ErrorTypeConversion]),
		"creation":    len(errorsByType[errs.ErrorTypeCreation]),
		"missing":     len(errorsByType[errs.ErrorTypeMissing]),
		"invalid":     len(errorsByType[errs.ErrorTypeInvalid]),
		"unsupported": len(errorsByType[errs.ErrorTypeUnsupported]),
	}

	return result
}

// ErrOrOk handles error or returns success response if no error
func (hdr *ResponseHandler) ErrOrOk(ctx RequestContext, data any, err error, successMsg ...string) {
	if err != nil {
		hdr.Error(ctx, err)
		return
	}

	msg := okMessage
	if len(successMsg) > 0 {
		msg = successMsg[0]
	}

	hdr.Ok(ctx, data, msg)
}

// ErrWithFallback handles structured errors with fallback
func (hdr *ResponseHandler) ErrWithFallback(ctx RequestContext, err error, context string, fallbackErr *errs.Error) {
	// If it's already a structured error, just pass it on
	var eErr *errs.Error
	if errors.As(err, &eErr) {
		hdr.Error(ctx, err)
		return
	}

	// Otherwise, use the error fallback with context
	hdr.Error(ctx, fallbackErr.
		WithContext(context).
		WithCause(err))
}

// NotFound handles not found errors with custom resource name
func (hdr *ResponseHandler) NotFound(ctx RequestContext, resourceType string, identifier any) {
	err := errs.ErrResourceNotFound.
		WithDetail("resource_type", resourceType).
		WithDetail("identifier", identifier)
	hdr.Error(ctx, err)
}

// IsClientError checks if the domain error should result in a 4xx response
func (hdr *ResponseHandler) IsClientError(err error) bool {
	var eError *errs.Error
	if !errors.As(err, &eError) {
		return false
	}

	switch eError.Type {
	case errs.ErrorTypeValidation,
		errs.ErrorTypeMissing,
		errs.ErrorTypeInvalid,
		errs.ErrorTypeNotFound:
		return true
	default:
		return false
	}
}

// IsServerError checks if the domain error should result in a 5xx response
func (hdr *ResponseHandler) IsServerError(err error) bool {
	var eError *errs.Error
	if !errors.As(err, &eError) {
		return false
	}

	switch eError.Type {
	case errs.ErrorTypeProcessing,
		errs.ErrorTypeConversion,
		errs.ErrorTypeCreation:
		return true
	default:
		return false
	}
}

// GetHTTPStatusCode gets the appropriate HTTP status code for a domain error
func (hdr *ResponseHandler) GetHTTPStatusCode(err error) int {
	var eError *errs.Error
	if errors.As(err, &eError) {
		return hdr.getStatusCodeForErrorType(eError.Type)
	}

	var errCollection *errs.ErrCollection
	if errors.As(err, &errCollection) {
		return hdr.getStatusCodeForErrorType(errCollection.Err.Type)
	}

	return http.StatusInternalServerError
}

// getMsgOrDefault returns the first message if provided, otherwise returns the default
func (hdr *ResponseHandler) getMsgOrDefault(messages []string, defaultMsg string) string {
	if len(messages) > 0 && messages[0] != "" {
		return messages[0]
	}
	return defaultMsg
}

// parseValidationErrors converts validation errors into a structured map
func (hdr *ResponseHandler) parseValidationErrors(err ...error) map[string]any {
	if len(err) == 0 || err[0] == nil {
		return nil
	}

	// Pre-allocate map with reasonable capacity to reduce allocations
	errorMap := make(map[string]any, 10)

	// Handle validator.ValidationErrors specifically
	if validationErrors, ok := err[0].(validator.ValidationErrors); ok {
		return hdr.parseValidatorErrors(validationErrors, errorMap)
	}

	// Handle generic errors
	errorMap["error"] = err[0].Error()
	return errorMap
}

// parseValidatorErrors processes validator.ValidationErrors into user-friendly messages
func (hdr *ResponseHandler) parseValidatorErrors(validationErrors validator.ValidationErrors, errorMap map[string]any) map[string]any {
	for _, fieldError := range validationErrors {
		field := fieldError.Field()
		tag := fieldError.Tag()
		param := fieldError.Param()
		fieldType := fieldError.Type()

		errorMap[field] = hdr.getValidationErrorMessage(tag, param, fieldType)
	}

	return errorMap
}

// getValidationErrorMessage returns user-friendly validation error messages in Portuguese
func (hdr *ResponseHandler) getValidationErrorMessage(tag, param string, fieldType reflect.Type) string {
	switch tag {
	case "required":
		return "Este campo é obrigatório"
	case "email":
		return "Formato de e-mail inválido"
	case "min":
		if fieldType.Kind() == reflect.String {
			return fmt.Sprintf("Deve ter pelo menos %s caracteres", param)
		}
		return fmt.Sprintf("Deve ser pelo menos %s", param)
	case "max":
		if fieldType.Kind() == reflect.String {
			return fmt.Sprintf("Não deve exceder %s caracteres", param)
		}
		return fmt.Sprintf("Não deve exceder %s", param)
	case "oneof":
		return fmt.Sprintf("Deve ser um dos seguintes: %s", param)
	case "len":
		if fieldType.Kind() == reflect.String {
			return fmt.Sprintf("Deve ter exatamente %s caracteres", param)
		}
		return fmt.Sprintf("Deve ter exatamente %s itens", param)
	case "gte":
		return fmt.Sprintf("Deve ser maior ou igual a %s", param)
	case "lte":
		return fmt.Sprintf("Deve ser menor ou igual a %s", param)
	case "gt":
		return fmt.Sprintf("Deve ser maior que %s", param)
	case "lt":
		return fmt.Sprintf("Deve ser menor que %s", param)
	case "alpha":
		return "Deve conter apenas caracteres alfabéticos"
	case "alphanum":
		return "Deve conter apenas caracteres alfanuméricos"
	case "numeric":
		return "Deve ser um número válido"
	case "url":
		return "Deve ser uma URL válida"
	case "uuid":
		return "Deve ser um UUID válido"
	default:
		return fmt.Sprintf("Falha na validação: %s", tag)
	}
}
