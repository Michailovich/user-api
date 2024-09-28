package userPack

import (
	"fmt"
	"regexp"
)

func isValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func validateUser(user User) error {
	if user.Firstname == "" || user.Lastname == "" || user.Email == "" {
		return fmt.Errorf("firstname, lastname and email are required")
	}
	if !isValidEmail(user.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}
