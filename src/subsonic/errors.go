package subsonic

import (
	"errors"
	"fmt"
)

var (
	ErrUnreachable     = errors.New("server not reachable")
	ErrInvalidResponse = errors.New("server returned invalid response")
)

type SubsonicError struct {
	Code    int
	Message string
}

func NewSubsonicError(code int, message string) *SubsonicError {
	return &SubsonicError{Code: code, Message: message}
}

func (err *SubsonicError) Is(other error) bool {
	casted, ok := other.(*SubsonicError)
	if !ok {
		return false
	}

	return err.Code == casted.Code
}

func (err *SubsonicError) Error() string {
	return fmt.Sprintf("subsonic error %d %s", err.Code, err.Message)
}
