package db

import (
	"fmt"
	"errors"
)

func validateLength(propertyName string, str string, length int) error {
	if len(str) > length {
		return errors.New(fmt.Sprintf("%s must have %d or less characters", propertyName, length))
	}

	return nil
}

func validateUsername(username string) error {
	return validateLength("Username", username, 10)
}

func validateFullname(fullName string) error {
	return validateLength("Fullname", fullName, 80)
}
