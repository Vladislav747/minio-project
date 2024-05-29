package apperror

import (
	"encoding/json"
	"fmt"
)

var (
	ErrNotFound     = NewApiError("not found", "FS-00010", "")
	ErrAlreadyExist = NewApiError("already exists", "FS-00011", "")
)

type AppError struct {
	Err              error  `json:"-"`
	Message          string `json:"message, omitempty"`
	DeveloperMessage string `json:"developer_message, omitempty"`
	Code             string `json:"code, omitempty"`
}

func NewApiError(message, code, developerMessage string) *AppError {
	return &AppError{
		Err:              fmt.Errorf(message),
		Message:          message,
		DeveloperMessage: developerMessage,
		Code:             code,
	}
}

func (e *AppError) Error() string {
	return e.Err.Error()
}

func (e *AppError) Unwrap() error { return e.Err }

func (e *AppError) Marshal() []byte {
	bytes, err := json.Marshal(e)
	if err != nil {
		return nil
	}
	return bytes
}

func UnauthorizedError(msg string) *AppError {
	return NewApiError(msg, "FS-000003", "")
}

func BadRequestError(msg string) *AppError {
	return NewApiError(msg, "FS-000002", "some thing wrong with user data")
}

func systemError(developerMsg string) *AppError {
	return NewApiError("system error", "FS-000001", developerMsg)
}
