package apperror

import "fmt"

type AppError struct {
	Code    int
	Message string
	Detail  any
}

func (e *AppError) Error() string {
	if e.Detail != nil {
		strDetail, ok := e.Detail.(string)
		if ok { return strDetail }

		return fmt.Sprintf("%s: %v", e.Message, e.Detail)
	}

	return e.Message
}

func ErrMethod(err string) *AppError {
	return &AppError{
		Code:    405,
		Message: "Method Not Allowed",
		Detail:  err,
	}
}

func ErrUnauthorized(err string) *AppError {
	return &AppError{
		Code:    401,
		Message: "Unauthorized",
		Detail:  err,
	}
}

func ErrForbidden(err string) *AppError {
	return &AppError{
		Code:    403,
		Message: "Forbidden",
		Detail:  err,
	}
}

func ErrInternal(err string) *AppError {
	return &AppError{
		Code:    500,
		Message: "Internal Server Error",
		Detail:  err,
	}
}

func ErrExternal(err string) *AppError {
	return &AppError{
		Code:    502,
		Message: "External Service Error",
		Detail:  err,
	}
}

func ErrNotFound(err string) *AppError {
	return &AppError{
		Code:    404,
		Message: "Not Found",
		Detail:  err,
	}
}

func ErrConflict(err string) *AppError {
	return &AppError{
		Code:    409,
		Message: "Conflict",
		Detail:  err,
	}
}

func ErrBadRequest(err string) *AppError {
	return &AppError{
		Code:    400,
		Message: "Bad Request",
		Detail:  err,
	}
}
