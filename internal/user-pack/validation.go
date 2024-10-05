package userPack

import (
	"fmt"
	"regexp"
)

func IsValidEmail(email string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return re.MatchString(email)
}

func ValidateUser(user *User) error {
	if user.Firstname == "" || user.Lastname == "" || user.Email == "" {
		return fmt.Errorf("firstname, lastname and email are required")
	}
	if !IsValidEmail(user.Email) {
		return fmt.Errorf("invalid email format")
	}
	return nil
}
