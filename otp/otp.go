package otp

import (
	"time"
)

type OTP struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (otp *OTP) Verify(code string) error {
	if otp.IsExpired() {
		return ErrCodeExpired
	}
	if otp.Code != code {
		return ErrCodeInvalid
	}
	return nil
}

func (otp *OTP) IsExpired() bool {
	return time.Now().After(otp.ExpiresAt)
}
