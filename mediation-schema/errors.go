package mediationschema

import (
	"fmt"
)

type MediationSchemaError struct {
	Code    string
	Message string
}

func (e MediationSchemaError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e MediationSchemaError) Is(err error) bool {
	if merr, ok := err.(MediationSchemaError); ok {
		return merr.Code == e.Code
	}
	return false
}

func (e *MediationSchemaError) WithMessagef(description string, vals ...interface{}) MediationSchemaError {
	return MediationSchemaError{
		Code:    e.Code,
		Message: fmt.Sprintf(description, vals...),
	}
}

func NewMediationSchemaError(msg, description string) MediationSchemaError {
	return MediationSchemaError{
		Code:    msg,
		Message: description,
	}
}

var ErrDuplicateDefinition MediationSchemaError = NewMediationSchemaError("duplicate definition", "the object was defined more than once")
