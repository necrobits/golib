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
	WhiteListDomains  []string `json:"whiteListDomains"`
	BlockOtherDomains bool     `json:"blockOtherDomains"`
}

func ValidateEmail(email string, config EmailValidationConfig) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return errors.B().
			Code(EInvalidEmail).
			Msg("invalid email format").
			Build()
	}

	if len(config.WhiteListDomains) > 0 {
		if slices.Contains(config.WhiteListDomains, "*") {
			return nil
		}

		domain := strings.Split(email, "@")[1]
		if !slices.Contains(config.WhiteListDomains, domain) && config.BlockOtherDomains {
			return errors.B().
				Code(EUnallowedEmail).
				Msg("email is not allowed").
				Build()
		}
	}

	return nil
}
