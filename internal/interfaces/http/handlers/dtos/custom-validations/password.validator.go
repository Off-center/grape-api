package customvalidations

import (
	"regexp"
	"unicode/utf8"

	"github.com/go-playground/validator/v10"
)

var (
	lowercaseRegex = regexp.MustCompile(`[a-z]`)
	uppercaseRegex = regexp.MustCompile(`[A-Z]`)
	numberRegex    = regexp.MustCompile(`[0-9]`)
	specialRegex   = regexp.MustCompile(`[^a-zA-Z0-9\s]`)
)

func Password(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if utf8.RuneCountInString(password) < 8 {
		return false
	}

	return lowercaseRegex.MatchString(password) &&
		uppercaseRegex.MatchString(password) &&
		numberRegex.MatchString(password) &&
		specialRegex.MatchString(password)
}
