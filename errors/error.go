package errors

import (
	"fmt"
)

type appError struct {
	Code string
	Msg  string
	Err  error
	// Op is the operation which caused the error, if any.
	Op   string
	Type string
}

const (
	EUnauthorized       = "unauthorized"
	EInternal           = "internal_error"
	EInvalidInput       = "invalid_input"
	EMalformedData      = "malformed_data"
	EUnexpectedDataType = "unexpected_data_type"
	EResourceNotFound   = "resource_not_found"
	ENotFound           = "not_found"
	EForbidden          = "forbidden"
	EConflict           = "conflict"

	TBadRequest      = "bad_request"
	TNotFound        = "not_found"
	TConflict        = "conflict"
	TInternal        = "internal"
	TForbidden       = "forbidden"
	TUnauthorized    = "unauthorized"
	TTooManyRequests = "too_many_requests"
)

type ErrModifier func(*appError)

func (e *appError) Error() string {
	if e.Msg != "" && e.Err != nil && e.Code != "" && e.Op != "" {
		return fmt.Sprintf("<%s> %s: %s (%s)", e.Code, e.Msg, e.Err.Error(), e.Op)
	}
	if e.Msg == "" && e.Err != nil && e.Code != "" {
		return fmt.Sprintf("<%s> %s: %s", e.Code, e.Msg, e.Err.Error())
	}
	if e.Msg != "" && e.Err != nil {
		return fmt.Sprintf("%s: %s", e.Msg, e.Err.Error())
	}
	if e.Msg != "" && e.Code != "" {
		return fmt.Sprintf("<%s> %s", e.Code, e.Msg)
	}
	if e.Msg != "" {
		return e.Msg
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return "unknown error"
}

func Is(err error, code string) bool {
	appErr, ok := err.(*appError)
	if !ok {
		return false
	}
	return appErr.Code == code || Is(appErr.Err, code)
}

func IsOneOf(err error, codes ...string) bool {
	for _, code := range codes {
		if Is(err, code) {
			return true
		}
	}
	return false
}

func Wrap(err error, m ...ErrModifier) error {
	appE := &appError{Code: EInternal, Msg: err.Error()}
	for _, mod := range m {
		mod(appE)
	}
	return appE
}

func WithCode(code string) ErrModifier {
	return func(e *appError) {
		e.Code = code
	}
}

func WithMsg(msg string) ErrModifier {
	return func(e *appError) {
		e.Msg = msg
	}
}

func WithOp(op string) ErrModifier {
	return func(e *appError) {
		e.Op = op
	}
}

func WithType(t string) ErrModifier {
	return func(e *appError) {
		e.Type = t
	}
}

func IsAppError(err error) bool {
	_, ok := err.(*appError)
	return ok
}

func RootErrCode(err error) string {
	if err == nil {
		return ""
	}
	appErr, ok := err.(*appError)
	if !ok {
		return EInternal
	}
	if appErr == nil {
		return ""
	}
	if appErr.Code != "" {
		return appErr.Code
	}
	if appErr.Err != nil {
		return RootErrCode(appErr.Err)
	}
	return EInternal
}

func GetErrorType(err error) string {
	if err == nil {
		return ""
	}
	appErr, ok := err.(*appError)
	if !ok {
		return TInternal
	}
	if appErr == nil {
		return ""
	}
	if appErr.Type != "" {
		return appErr.Type
	}
	if appErr.Err != nil {
		return GetErrorType(appErr.Err)
	}
	return TInternal
}

func ErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	appErr, ok := err.(*appError)
	if !ok {
		return "An internal error occurred."
	}
	if appErr == nil {
		return ""
	}
	if appErr.Msg != "" {
		return appErr.Msg
	}
	if appErr.Err != nil {
		return ErrorMessage(appErr.Err)
	}
	return ""
}
