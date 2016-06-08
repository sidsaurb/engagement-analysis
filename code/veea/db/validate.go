package db

import (
	"fmt"
	"errors"

	"github.com/gpahal/veea/conf"
)

func validateLength(propertyName string, str string, length int) error {
	if len(str) > length {
		return errors.New(fmt.Sprintf("%s must have %d or less characters", propertyName, length))
	}

	return nil
}

func validateUsername(username string) error {
	if username == conf.AdminUsername {
		return errors.New(fmt.Sprintf("Username %s is not allowed (reserved)", username))
	}

	return validateLength("Username", username, 10)
}

func validateFullname(fullName string) error {
	return validateLength("Fullname", fullName, 80)
}
