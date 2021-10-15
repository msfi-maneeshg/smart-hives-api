package common

import (
	"regexp"
)

func IsEmailValid(e string) bool {
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return emailRegex.MatchString(e)
}

func IsPasswordStrong(p string) bool {
	checkNumber := regexp.MustCompile(`[0-9]`)
	if !checkNumber.MatchString(p) {
		return false
	}

	checkAlfabet := regexp.MustCompile(`[a-z]`)
	if !checkAlfabet.MatchString(p) {
		return false
	}

	checkCapitalAlfabet := regexp.MustCompile(`[A-Z]`)
	if !checkCapitalAlfabet.MatchString(p) {
		return false
	}

	checkSpecialChar := regexp.MustCompile(`[!@#\$%\^&\*]`)
	if !checkSpecialChar.MatchString(p) {
		return false
	}

	checkLength := regexp.MustCompile(`^.{8,32}$`)

	return checkLength.MatchString(p)
}
