package sc

import (
	"errors"
	"fmt"
)

// Common errors
var (
	ErrInvalidState      = errors.New("invalid state of knowledge base")
	ErrInvalidValue      = errors.New("invalid value")
	ErrTimeout           = errors.New("timeout")
	ErrConnectionFailed  = errors.New("connection failed")
	ErrElementNotFound   = errors.New("element not found")
	ErrEventNotFound     = errors.New("event not found")
	ErrInvalidAlias      = errors.New("invalid alias")
	ErrInvalidType       = errors.New("invalid type")
	ErrInvalidParameters = errors.New("invalid parameters")
)

// CommonError creates a common error with message
func CommonError(errorType error, msg string) error {
	return fmt.Errorf("%s: %s", errorType.Error(), msg)
}

// KnowledgeBaseError creates knowledge base error
func KnowledgeBaseError(msg string) error {
	return CommonError(ErrInvalidState, msg)
}

// InvalidValueError creates invalid value error
func InvalidValueError(msg string) error {
	return CommonError(ErrInvalidValue, msg)
}
