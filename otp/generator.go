package otp

import (
	"crypto/rand"
	"time"
)

const (
	_Digits  = "1234567890"
	_Charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

type CodeOpts struct {
	Length     int
	ExpiresIn  time.Duration
	DigitsOnly bool
}

func GenerateOtp(opts CodeOpts) (*OTP, error) {
	code, err := randString(opts.Length, opts.DigitsOnly)
	if err != nil {
		return nil, err
	}
	return &OTP{
		Code:      code,
		ExpiresAt: time.Now().Add(opts.ExpiresIn),
	}, nil
}

func randString(length int, digitOnly bool) (string, error) {
	buffer := make([]byte, length)
	_, err := rand.Read(buffer)
	if err != nil {
		return "", err
	}

	charsSet := _Charset
	if digitOnly {
		charsSet = _Digits
	}
	strCharsLength := len(charsSet)
	for i := 0; i < length; i++ {
		buffer[i] = charsSet[int(buffer[i])%strCharsLength]
	}

	return string(buffer), nil
}
