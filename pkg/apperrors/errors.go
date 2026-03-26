package apperrors

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int
	Message string
}

func (e *AppError) Error() string { return e.Message }

var (
	ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrConflict     = &AppError{Code: http.StatusConflict, Message: "resource already exists"}
	ErrBadRequest   = &AppError{Code: http.StatusBadRequest, Message: "bad request"}
	ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrInternal     = &AppError{Code: http.StatusInternalServerError, Message: "internal server error"}
)

func As(err error, target any) bool { return errors.As(err, target) }