package errs

import (
	"errors"
	"fmt"
	"strings"
)

// ErrCollection represents a errCollection of errors
type ErrCollection struct {
	Err    *Error
	Errors []error
}

// NewErrCollection creates a new ErrCollection with a base Error
func NewErrCollection(errType ErrorType, code string, message string) *ErrCollection {
	return &ErrCollection{
		Err: &Error{
			Type:    errType,
			Code:    code,
			Message: message,
		},
		Errors: []error{},
	}
}

// NewStdErrCollection creates a new error collector for accumulating errors
func NewStdErrCollection() *ErrCollection {
	return NewErrCollection(ErrorTypeProcessing,
		"ERR_COLLECTION", "Um ou mais erros ocorreram durante a operação")
}

// Add adds an error to the errCollection
func (m *ErrCollection) Add(err error) {
	if err != nil {
		m.Errors = append(m.Errors, err)
	}
}

// AddAll adds multiple errors to the errCollection
func (m *ErrCollection) AddAll(errs ...error) {
	for _, err := range errs {
		m.Add(err)
	}
}

// HasErrors returns true if there are any errors in the errCollection
func (m *ErrCollection) HasErrors() bool {
	return len(m.Errors) > 0
}

// Count returns the number of errors
func (m *ErrCollection) Count() int {
	return len(m.Errors)
}

// Error implements the error interface
func (m *ErrCollection) Error() string {
	if !m.HasErrors() {
		return m.Err.Error()
	}

	var errorMessages []string
	for _, err := range m.Errors {
		errorMessages = append(errorMessages, err.Error())
	}

	return fmt.Sprintf("%s: [%s]", m.Err.Error(), strings.Join(errorMessages, "; "))
}

// Unwrap returns the errors as a slice (Go 1.20+ multiple error unwrapping)
func (m *ErrCollection) Unwrap() []error {
	return m.Errors
}

// ToError returns nil if no errors, otherwise returns the ErrCollection
func (m *ErrCollection) ToError() error {
	if !m.HasErrors() {
		return nil
	}
	return m.Err
}

// GetErrs returns all errors in the errCollection
func (m *ErrCollection) GetErrs() []error {
	return m.Errors
}

// GetErrors returns only Error types from the errCollection
func (m *ErrCollection) GetErrors() []*Error {
	var errsList []*Error
	for _, err := range m.Errors {
		var sError *Error
		if errors.As(err, &sError) {
			errsList = append(errsList, sError)
		}
	}
	return errsList
}

// GetByType returns errors filtered by ErrorType
func (m *ErrCollection) GetByType(errType ErrorType) []error {
	var filtered []error
	for _, err := range m.Errors {
		var sError *Error
		if errors.As(err, &sError) && sError.Type == errType {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// GetByCode returns errors filtered by error code
func (m *ErrCollection) GetByCode(code string) []error {
	var filtered []error
	for _, err := range m.Errors {
		var sError *Error
		if errors.As(err, &sError) && sError.Code == code {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// First returns the first error in the errCollection, or nil if empty
func (m *ErrCollection) First() error {
	if len(m.Errors) > 0 {
		return m.Errors[0]
	}
	return nil
}

// Last returns the last error in the errCollection, or nil if empty
func (m *ErrCollection) Last() error {
	if len(m.Errors) > 0 {
		return m.Errors[len(m.Errors)-1]
	}
	return nil
}

// WrapErrors wraps a slice of errors into a single ErrCollection
func WrapErrors(errs []error, errType ErrorType, code string, message string) error {
	if len(errs) == 0 {
		return nil
	}

	errCollection := NewErrCollection(errType, code, message)
	errCollection.AddAll(errs...)
	return errCollection.Err
}
