package db

import (
	"time"
	"errors"
)

type User struct {
	Id int64
	Username string
	FullName string
	IsAdmin bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetUser(userId int64, otherUserId int64) (*User, error) {
	err := errorFold(
		UserIdExists(userId),
		UserIdExists(otherUserId),
	)
	if err != nil {
		return nil, &UserError{error: err}
	}

	rows, err := query("SELECT id, username, full_name, is_admin, created_at, updated_at FROM user WHERE id = ? LIMIT 1", otherUserId)
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var isAdmin int
		var user User
		err = rows.Scan(
			&user.Id,
			&user.Username,
			&user.FullName,
			&isAdmin,
			&user.CreatedAt,
			&user.UpdatedAt,
		)

		if err != nil {
			return nil, &InternalError{error: err}
		}

		if isAdmin > 0 {
			user.IsAdmin = true
		} else {
			user.IsAdmin = false
		}

		return &user, nil
	}

	return nil, &UserError{error: errors.New("User does not exist")}
}

func CreateUser(username string, password string, fullName string) (int64, error) {
	err := errorFold(
		validateUsername(username),
		validateFullname(fullName),
		UsernameNotExists(username),
	)
	if err != nil {
		return 0, &UserError{error: err}
	}

	hash, err := generateHash(password)
	if err != nil {
		return 0, &InternalError{error: err}
	}

	res, err := exec("INSERT INTO user (username, full_name, password_hash) VALUES (?, ?, ?)", username, fullName, hash)
	if err != nil {
		return 0, &InternalError{error: err}
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, &InternalError{error: err}
	}

	return id, nil
}

func UpdateUsernameAndFullName(id int64, username string, fullName string) error {
	err := errorFold(
		validateUsername(username),
		validateFullname(fullName),
		UserIdExists(id),
		UsernameOtherNotExists(id, username),
	)
	if err != nil {
		return &UserError{error: err}
	}

	_, err = exec("UPDATE user SET username = ?, full_name = ? WHERE id = ?", username, fullName, id)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}

func UpdatePassword(id int64, password string) error {
	err := errorFold(
		UserIdExists(id),
	)
	if err != nil {
		return &UserError{error: err}
	}

	hash, err := generateHash(password)
	if err != nil {
		return &InternalError{error: err}
	}

	_, err = exec("UPDATE user SET password_hash = ? WHERE id = ?", hash, id)
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
}

func Authenticate(username string, password string) (int64, bool, error) {
	rows, err := query("SELECT id, password_hash FROM user WHERE username = ?", username)
	if err != nil {
		return 0, false, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var id int64
		var hash string

		err = rows.Scan(&id, &hash)
		if err != nil {
			return 0, false, &InternalError{error: err}
		}

		if compareHash(hash, password) != nil {
			return 0, false, nil
		}

		return id, true, nil
	}

	return 0, false, nil
}
