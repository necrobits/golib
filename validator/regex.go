package validator

import (
	"errors"
	"regexp"
)

func ValidateRegex(str string, regex string) error {
	if !regexp.MustCompile(regex).MatchString(str) {
		return errors.New("string does not match regex")
	}
	return nil
}
