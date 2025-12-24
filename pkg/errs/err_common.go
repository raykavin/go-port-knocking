package errs

import (
	"strings"
)

var (
	ErrInvalidBodyFormat = &Error{
		Type:    ErrorTypeInvalid,
		Code:    "ERR_INVALID_BODY_FORMAT",
		Message: "Corpo da solicitação inválido",
	}
	ErrResourceNotFound = &Error{
		Type:    ErrorTypeNotFound,
		Code:    "ERR_RESOURCE_NOT_FOUND",
		Message: "O recurso solicitado não foi encontrado",
	}
	ErrCreateResourceFailed = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_CREATE_RESOURCE_FAILED",
		Message: "Falha ao criar o recurso solicitado",
	}
	ErrUpdateResourceFailed = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_UPDATE_RESOURCE_FAILED",
		Message: "Falha ao atualizar o recurso solicitado",
	}
	ErrGetResource = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_GET_RESOURCE_ERROR",
		Message: "Falha ao obter recurso solicitado",
	}
	ErrFilterFailed = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_FILTER_FAILED",
		Message: "Falha ao realizar filtragem do(s) recurso(s) solicitado(s)",
	}

	ErrEmptyToken = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_EMPTY_TOKEN",
		Message: "Token de autorização vazio",
	}
	ErrForbidden = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_FORBIDDEN",
		Message: "Acesso proibido",
	}
	ErrUnauthorized = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_UNAUTHORIZED",
		Message: "Acesso não autorizado",
	}
	ErrUnknown = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_UNKNOWN_ERROR",
		Message: "Erro desconhecido ou não tratado",
	}
	ErrEmptyListItems = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_EMPTY_LIST_ITEMS",
		Message: "A lista de items não pode ser vazia",
	}
	ErrEntityIsNil = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_ENTITY_IS_NIL",
		Message: "A entidade não pode ser nula",
	}
	ErrDTOIsNil = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_DTO_IS_NIL",
		Message: "O objeto de transferência de dados não pode ser nulo",
	}
	ErrNilObject = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_NIL_OBJECT",
		Message: "O objeto não pode ser nulo",
	}
	ErrDTOConversionFailed = &Error{
		Type:    ErrorTypeProcessing,
		Code:    "ERR_DTO_CONVERSION_FAILED",
		Message: "Falha ao converter objeto de transferência de dados",
	}
	ErrValueIsZero = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_ZERO_VALUE",
		Message: "O valor informado não pode ser zero",
	}
	ErrGetFields = &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_GET_BY_FIELDS",
		Message: "Ao menos um campo de busca deve ser fornecido",
	}
)

func ErrMissingRequiredField(fieldName string, context string) *Error {
	return &Error{
		Type:    ErrorTypeMissing,
		Code:    "ERR_MISSING_REQUIRED_FIELD",
		Message: "Campo obrigatório ausente",
		Details: map[string]any{
			"campo":    fieldName,
			"contexto": context,
		},
	}
}

// Preset errs
func ErrMissingRequiredFields(context string, fields ...string) *Error {
	return &Error{
		Type:    ErrorTypeMissing,
		Code:    "ERR_MISSING_REQUIRED_FIELDS",
		Message: "Um ou mas campos obrigatórios estão ausentes",
		Details: map[string]any{
			"campos":   strings.Join(fields, ","),
			"contexto": context,
		},
	}
}

func ErrMissingRequiredDependency(dependencyName string, context string) *Error {
	return &Error{
		Type:    ErrorTypeMissing,
		Code:    "ERR_MISSING_REQUIRED_DEPENDENCY",
		Message: "Uma dependência obrigatória não foi fornecida",
		Details: map[string]any{
			"dependencia": dependencyName,
			"contexto":    context,
		},
	}
}

// Validation preset functions
func ErrValidationFailed(field string, value any, reason string) *Error {
	return &Error{
		Type:    ErrorTypeValidation,
		Code:    "ERR_VALIDATION_FAILED",
		Message: "Falha na validação",
		Details: map[string]any{
			"campo":  field,
			"valor":  value,
			"motivo": reason,
		},
	}
}

func ErrInvalidStartDate(field string, value any) *Error {
	return ErrValidationFailed(field, value, "Formato de data de início inválido, esperado YYYY-MM-DD")
}

func ErrInvalidEndDate(field string, value any) *Error {
	return ErrValidationFailed(field, value, "Formato de data de final inválido, esperado YYYY-MM-DD")
}
