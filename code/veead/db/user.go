package db

import (
	"time"
	"errors"

	"github.com/gpahal/veead/conf"
)

type User struct {
	Id int64
	Username string
	FullName string
	IsAdmin bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func GetUsers(userId int64) ([]*User, error) {
	err := errorFold(
		UserIdAdminExists(userId),
	)
	if err != nil {
		return nil, &UserError{error: err}
	}

	rows, err := query("SELECT id, username, full_name, is_admin, created_at, updated_at FROM user")
	if err != nil {
		return nil, &InternalError{error: err}
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
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

		users = append(users, &user)
	}

	return users, nil
}

func GetUser(userId int64, otherUserId int64) (*User, error) {
	err := errorFold(
		UserIdAdminExists(userId),
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

func CreateAdminUserIfNotExists() error {
	username := conf.AdminUsername
	fullName := conf.AdminFullname
	password := conf.AdminPassword

	err := errorFold(
		validateUsername(username),
		validateFullname(fullName),
	)
	if err != nil {
		return &UserError{error: err}
	}

	err = AdminUserExists()
	if err == nil {
		return nil
	}

	err = errorFold(
		UsernameNotExists(password),
	)
	if err != nil {
		return &UserError{error: err}
	}

	hash, err := generateHash(password)
	if err != nil {
		return &InternalError{error: err}
	}

	res, err := exec("INSERT INTO user (username, full_name, password_hash, is_admin) VALUES (?, ?, ?, ?)", username, fullName, hash, 1)
	if err != nil {
		return &InternalError{error: err}
	}

	_, err = res.LastInsertId()
	if err != nil {
		return &InternalError{error: err}
	}

	return nil
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

func AuthenticateAdmin(username string, password string) (int64, bool, error) {
	rows, err := query("SELECT id, password_hash FROM user WHERE username = ? AND is_admin > 0", username)
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
