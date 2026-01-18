package model

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound                = errors.New("not found")
	ErrNotAuthenticated        = errors.New("not authenticated")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrMissingParameter        = errors.New("missing parameter")
)

func NewErrMissingParameter(name string) error {
	return &missingParameterError{Parameter: name}
}

type missingParameterError struct {
	Parameter string
}

func (err *missingParameterError) Is(other error) bool {
	return other == ErrMissingParameter
}

func (err *missingParameterError) Error() string {
	return fmt.Sprintf("missing parameter '%s'", err.Parameter)
}
