package validator

import (
	"errors"
	"strings"
)

type StringValidationConfig struct {
	MinLength           int    `json:"minLength" mapstructure:"min_length"`
	MaxLength           int    `json:"maxLength" mapstructure:"max_length"`
	MinDigits           int    `json:"minDigits" mapstructure:"min_digits"`
	MinUppers           int    `json:"minUppers" mapstructure:"min_uppers"`
	MinLowers           int    `json:"minLowers" mapstructure:"min_lowers"`
	MinSpecials         int    `json:"minSpecials" mapstructure:"min_specials"`
	WhitespaceAllowed   bool   `json:"whitespaceAllowed" mapstructure:"whitespace_allowed"`
	AllowedSpecialChars string `json:"allowedSpecialChars" mapstructure:"allowed_special_chars"`
}

func ValidateString(str string, config StringValidationConfig) error {
	var (
		numDigits   int
		numUppers   int
		numLowers   int
		numSpecials int
	)

	if len(str) < config.MinLength {
		return errors.New("string is too short")
	}
	if len(str) > config.MaxLength {
		return errors.New("string is too long")
	}

	for _, char := range str {
		switch {
		case char >= '0' && char <= '9':
			numDigits++
		case char >= 'A' && char <= 'Z':
			numUppers++
		case char >= 'a' && char <= 'z':
			numLowers++
		case char == ' ':
			if !config.WhitespaceAllowed {
				return errors.New("whitespace is not allowed")
			}
		case config.AllowedSpecialChars != "":
			if !strings.ContainsRune(config.AllowedSpecialChars, char) {
				return errors.New("special character is not allowed")
			}
			numSpecials++
		}
	}

	if numDigits < config.MinDigits {
		return errors.New("string does not contain enough digits")
	}
	if numUppers < config.MinUppers {
		return errors.New("string does not contain enough uppercase characters")
	}
	if numLowers < config.MinLowers {
		return errors.New("string does not contain enough lowercase characters")
	}
	if numSpecials < config.MinSpecials {
		return errors.New("string does not contain enough special characters")
	}

	return nil
}
