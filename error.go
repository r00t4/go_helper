package helper

import "errors"

var (
	EmptyErr = errors.New("")
	ErrInvalidToken = errors.New("invalid token")
	ErrUnexpiredToken = errors.New("can't generate, unexpired token")
)

type HttpError interface {
	error
	Status() int
}

type MiddleHttpError struct {
	Code int
	Err  error
}

func (se MiddleHttpError) Error() string {
	return se.Err.Error()
}

func (se MiddleHttpError) Status() int {
	return se.Code
}
