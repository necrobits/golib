package validator

import (
	"errors"
	"net/mail"
	"slices"
	"strings"
)

type EmailValidationConfig struct {
	WhiteListDomains  []string `json:"whiteListDomains"`
	BlockOtherDomains bool     `json:"blockOtherDomains"`
}

func ValidateEmail(email string, config EmailValidationConfig) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}

	if len(config.WhiteListDomains) > 0 {
		if slices.Contains(config.WhiteListDomains, "*") {
			return nil
		}

		domain := strings.Split(email, "@")[1]
		if !slices.Contains(config.WhiteListDomains, domain) && config.BlockOtherDomains {
			return errors.New("email is not allowed")
		}
	}

	return nil
}
