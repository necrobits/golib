package otp

import (
	"time"

	"github.com/necrobits/x/errors"
)

type OTP struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (otp *OTP) Verify(code string) error {
	if otp.IsExpired() {
		return errors.B().
			Code(ErrCodeExpired).
			Msg("code is expired").Build()
	}
	if otp.Code != code {
		return errors.B().
			Code(ErrCodeInvalid).
			Msg("code is invalid").Build()
	}
	return nil
}

func (otp *OTP) IsExpired() bool {
	return time.Now().After(otp.ExpiresAt)
}
