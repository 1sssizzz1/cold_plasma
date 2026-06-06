package service

import "errors"

var (
	ErrValidation   = errors.New("validation")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrNotFound     = errors.New("not_found")
)

// ServiceError позволяет хранить человеко-читаемое сообщение,
// сохраняя при этом "тип" ошибки для errors.Is().
type ServiceError struct {
	Kind    error
	Message string
}

func (e ServiceError) Error() string {
	return e.Message
}

func (e ServiceError) Unwrap() error {
	return e.Kind
}

func se(kind error, msg string) error {
	return ServiceError{Kind: kind, Message: msg}
}
