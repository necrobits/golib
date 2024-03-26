package validator

import (
	"net/mail"
	"slices"
	"strings"

	"github.com/necrobits/x/errors"
)

const (
	EInvalidEmail   = "invalid_email"
	EUnallowedEmail = "unallowed_email"
)

type EmailValidationConfig struct {
	WhiteListedDomains []string `json:"whiteListedDomains"`
	BlockOtherDomains  bool     `json:"blockOtherDomains"`
}

func ValidateEmail(email string, config EmailValidationConfig) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.B().
			Code(EInvalidEmail).
			Msg("invalid email format").
			Build()
	}

	if len(config.WhiteListedDomains) > 0 {
		if slices.Contains(config.WhiteListedDomains, "*") {
			return nil
		}

		domain := strings.Split(email, "@")[1]
		if !slices.Contains(config.WhiteListedDomains, domain) && config.BlockOtherDomains {
			return errors.B().
				Code(EUnallowedEmail).
				Msg("email is not allowed").
				Build()
		}
	}

	return nil
}
