package otp

import "errors"

var (
	ErrCodeExpired = errors.New("otp code is expired")
	ErrCodeInvalid = errors.New("otp code is invalid")
)
