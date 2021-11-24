package mediationschema

import (
	"fmt"
)

type MediationSchemaError struct {
	Message     string
	Description string
}

func (e MediationSchemaError) Error() string {
	result := fmt.Sprintf("error: %s", e.Message)
	if e.Description != "" {
		result += fmt.Sprintf(". description: %s", e.Description)
	}
	return result
}

func (e *MediationSchemaError) WithDescriptionf(description string, vals ...interface{}) MediationSchemaError {
	return MediationSchemaError{
		Message:     e.Message,
		Description: fmt.Sprintf(description, vals...),
	}
}

func NewMediationSchemaError(msg, description string) MediationSchemaError {
	return MediationSchemaError{
		Message:     msg,
		Description: description,
	}
}

var ErrDuplicateDefinition MediationSchemaError = NewMediationSchemaError("duplicate definition", "the object was defined more than once")
