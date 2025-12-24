package errs

import (
	"errors"
	"fmt"
	"maps"
	"strings"
)

// ErrorType represents the error types of the domain
type ErrorType string

const (
	ErrorTypeValidation  ErrorType = "validation"
	ErrorTypeNotFound    ErrorType = "not_found"
	ErrorTypeProcessing  ErrorType = "processing"
	ErrorTypeConversion  ErrorType = "conversion"
	ErrorTypeCreation    ErrorType = "creation"
	ErrorTypeMissing     ErrorType = "missing"
	ErrorTypeInvalid     ErrorType = "invalid"
	ErrorTypeUnsupported ErrorType = "unsupported"
)

// Error represents a typed domain error
type Error struct {
	Type    ErrorType
	Code    string
	Message string
	Details map[string]any
	Cause   error
}

// NewError creates a new domain error
func New(errType ErrorType, code string, message string, context ...any) *Error {
	e := &Error{
		Type:    errType,
		Code:    code,
		Message: message,
	}

	if len(context) > 0 {
		e = e.WithContext(context[0])
	}

	return e
}

func (e *Error) Error() string {
	msg := e.Message

	// Add details if they exist
	if len(e.Details) > 0 {
		var details []string
		for key, value := range e.Details {
			details = append(details, fmt.Sprintf("%s=%v", key, value))
		}
		msg = fmt.Sprintf("%s [%s]", msg, strings.Join(details, ", "))
	}

	// Add cause if it exists
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}

	return msg
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func (e *Error) Is(target error) bool {
	if t, ok := target.(*Error); ok {
		return e.Type == t.Type && e.Code == t.Code
	}
	return false
}

func (e *Error) WithCause(cause error) *Error {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

func (e *Error) WithDetails(details map[string]any) *Error {
	newErr := *e
	if newErr.Details == nil {
		newErr.Details = make(map[string]any)
	}
	maps.Copy(newErr.Details, details)
	return &newErr
}

func (e *Error) WithDetail(key string, value any) *Error {
	return e.WithDetails(map[string]any{key: value})
}

func (e *Error) WithContext(context any) *Error {
	return e.WithDetail("contexto", context)
}

func (e *Error) WithObject(obj any) *Error {
	return e.WithDetail("objeto", obj)
}

func (e *Error) WithExpected(exp any) *Error {
	return e.WithDetail("esperado", exp)
}

func (e *Error) WithFound(found any) *Error {
	return e.WithDetail("encontrado", found)
}

func (e *Error) WithResourceID(id uint64) *Error {
	return e.WithDetail("id_recurso", id)
}

// IsValidationError checks if the error is a validation error
func IsValidationError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeValidation
}

// IsNotFoundError checks if the error is a not found error
func IsNotFoundError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeNotFound
}

// IsMissingError checks if the error is a missing field error
func IsMissingError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeMissing
}

// IsInvalidError checks if the error is an invalid format error
func IsInvalidError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeInvalid
}

// IsProcessingError checks if the error is a processing error
func IsProcessingError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeProcessing
}

// IsCreationError checks if the error is a creation error
func IsCreationError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeCreation
}

// IsConversionError checks if the error is a conversion error
func IsConversionError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeConversion
}

// IsUnsupportedError checks if the error is an unsupported operation error
func IsUnsupportedError(err error) bool {
	var eErr *Error
	return errors.As(err, &eErr) && eErr.Type == ErrorTypeUnsupported
}

// GetErrorCode returns the error code if it's a Error
func GetErrorCode(err error) string {
	var eErr *Error
	if errors.As(err, &eErr) {
		return eErr.Code
	}
	return ""
}

// GetErrorType returns the error type if it's a Error
func GetErrorType(err error) ErrorType {
	var eErr *Error
	if errors.As(err, &eErr) {
		return eErr.Type
	}
	return ""
}

// CombineErrors combines multiple errors into a single error
func CombineErrors(mainErr *Error, errs ...error) error {
	if mainErr == nil && len(errs) == 0 {
		return nil
	}

	if mainErr == nil {
		mainErr = &Error{
			Type:    ErrorTypeProcessing,
			Code:    "MULTIPLE_ERRORS",
			Message: "Múltiplos erros ocorreram",
		}
	}

	var validErrs []error
	for _, err := range errs {
		if err != nil {
			validErrs = append(validErrs, err)
		}
	}

	if len(validErrs) == 0 {
		return mainErr
	}

	if len(validErrs) == 1 {
		return mainErr.WithCause(validErrs[0])
	}

	// For multiple errors, add as detail
	errorMessages := make([]string, len(validErrs))
	for i, err := range validErrs {
		errorMessages[i] = err.Error()
	}

	return mainErr.WithDetail("additional_errors", errorMessages)
}

// WrapError wraps a common error into a Error
func WrapError(err error, errType ErrorType, code string, message string) *Error {
	if err == nil {
		return nil
	}

	return &Error{
		Type:    errType,
		Code:    code,
		Message: message,
		Cause:   err,
	}
}
