package db

import (
	"time"
	"sync"
	"database/sql"
)

var (
	sessionIdLock sync.Mutex
)

func IsLoggedInUserId(userId int64) (string, int, error) {
	rows, err := query("SELECT session_id, created_at, is_active FROM session WHERE user_id = ?", userId)
	if err != nil {
		return "", 0, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var sessionId string
		var createdAt time.Time
		var is_active int

		err = rows.Scan(&sessionId, &createdAt, &is_active)
		if err != nil {
			return "", 0, &InternalError{error: err}
		}

		if sessionExpired(createdAt) || is_active < 1 {
			return "", 0, nil
		}

		return sessionId, 1, nil
	}

	return "", -1, nil
}

func IsLoggedInSessionId(sessionId string) (int64, bool, error) {
	rows, err := query("SELECT user_id, created_at, is_active FROM session WHERE session_id = ?", sessionId)
	if err != nil {
		return 0, false, &InternalError{error: err}
	}
	defer rows.Close()

	if rows.Next() {
		var userId int64
		var createdAt time.Time
		var is_active int

		err = rows.Scan(&userId, &createdAt, &is_active)
		if err != nil {
			return 0, false, &InternalError{error: err}
		}

		if sessionExpired(createdAt) || is_active < 1 {
			return 0, false, nil
		}

		return userId, true, nil
	}

	return 0, false, nil
}

func Login(username string, password string) (string, bool, error) {
	id, successful, err := Authenticate(username, password)
	if err != nil {
		return "", false, err
	}
	if !successful {
		return "", false, nil
	}

	sessionId, success, err := IsLoggedInUserId(id)
	if err != nil {
		return "", false, err
	}
	if success < 1 {
		sessionIdLock.Lock()
		defer sessionIdLock.Unlock()

		sessionId, err := GenerateSessionId()
		if err != nil {
			return "", false, &InternalError{error: err}
		}

		var res sql.Result

		if success < 0 {
			res, err = exec("INSERT INTO session (user_id, session_id, is_active) VALUES (?, ?, ?)", id, sessionId, 1)
		} else {
			res, err = exec("UPDATE session SET session_id = ?, is_active = ? WHERE user_id = ?", sessionId, 1, id)
		}

		if err != nil {
			return "", false, &InternalError{error: err}
		}

		_, err = res.LastInsertId()
		if err != nil {
			return "", false, &InternalError{error: err}
		}

		return sessionId, true, nil
	}

	return sessionId, true, nil
}

func Logout(sessionId string) error {
	userId, successful, err := IsLoggedInSessionId(sessionId)
	if err != nil {
		return err
	}

	if successful {
		res, err := exec("UPDATE session SET is_active = 0 WHERE user_id = ?", userId)

		if err != nil {
			return &InternalError{error: err}
		}

		_, err = res.LastInsertId()
		if err != nil {
			return &InternalError{error: err}
		}

		return nil
	}

	return nil
}
