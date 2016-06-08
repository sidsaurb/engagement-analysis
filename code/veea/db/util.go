package db

import (
	"time"
	"errors"
	"math/rand"

	"github.com/gpahal/veea/conf"
	"golang.org/x/crypto/bcrypt"
)

var src rand.Source

func init() {
	src = rand.NewSource(time.Now().UnixNano())
}

func generateHash(str string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(str), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func compareHash(hash string, str string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(str))
}

func errorFold(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func errorOr(errs ...error) error {
	var finalErr error = nil

	for _, err := range errs {
		if err == nil {
			return nil
		}
		finalErr = err
	}

	return finalErr
}

func UserIdExists(id int64) error {
	rows, err := query("SELECT * FROM user WHERE id = ?", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return nil
	}
	return errors.New("User id does not exist")
}

func UserIdAdminExists(id int64) error {
	rows, err := query("SELECT * FROM user WHERE id = ? AND is_admin > 0", id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return nil
	}
	return errors.New("Admin user id does not exist")
}

func VideoIdExists(videoId string) error {
	rows, err := query("SELECT * FROM video WHERE video_id = ?", videoId)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return nil
	}
	return errors.New("Video id does not exist")
}

func ViewIdNotExpiredExists(videoId string, viewId string) error {
	rows, err := query("SELECT * FROM video_view WHERE video_id = ? AND view_id = ? AND created_at >= ?", videoId, viewId, time.Unix(time.Now().Unix() - conf.ViewExpireTime, 0))
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return nil
	}
	return errors.New("View id does not exist")
}

func ViewTimeIdExists(viewTimeId int64) error {
	rows, err := query("SELECT * FROM video_view_time WHERE id = ?", viewTimeId)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return nil
	}
	return errors.New("View time id does not exist")
}

func UsernameNotExists(username string) error {
	rows, err := query("SELECT * FROM user WHERE username = ?", username)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return errors.New("Username already exists")
	}
	return nil
}

func UsernameOtherNotExists(id int64, username string) error {
	rows, err := query("SELECT * FROM user WHERE username = ? AND id <> ?", username, id)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return errors.New("Username already exists")
	}
	return nil
}

func sessionIdNotExists(sessionId string) error {
	rows, err := query("SELECT * FROM session WHERE session_id = ?", sessionId)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return errors.New("Session id already exists")
	}
	return nil
}

func viewIdUniqueNotExists(viewId string) error {
	rows, err := query("SELECT * FROM video_view WHERE view_id = ?", viewId)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return errors.New("View id already exists")
	}
	return nil
}

func sessionExpired(createdAt time.Time) bool {
	return (createdAt.Unix() + conf.SessionExpireTime) < time.Now().Unix()
}

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                        // 6 bits to represent a letter index
	letterIdxMask = (1 << letterIdxBits) - 1 // binary number with (letterIdxBits) digits, all 1
	letterIdxMax  = 63 / letterIdxBits       // number of letter indices fitting in 63 bits
)

func randomString(n int) string {
	b := make([]byte, n)

	// src.Int63() generates 63 random bits, enough for letterIdxMax characters!

	for i, cache, remaining := (n - 1), src.Int63(), letterIdxMax; i >= 0; {
		if remaining == 0 {
			cache, remaining = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i -= 1
		}
		cache >>= letterIdxBits
		remaining -= 1
	}

	return string(b)
}

func GenerateSessionId() (string, error) {
	tries := 5
	for tries > 0 {
		sessionId := randomString(conf.SessionIdLength)
		if err := sessionIdNotExists(sessionId); err == nil {
			return sessionId, nil
		}
		tries -= 1
	}

	return "", errors.New("Unable to generate session id")
}

func GenerateViewId() (string, error) {
	tries := 5
	for tries > 0 {
		viewId := randomString(conf.ViewIdLength)
		if err := viewIdUniqueNotExists(viewId); err == nil {
			return viewId, nil
		}
		tries -= 1
	}

	return "", errors.New("Unable to generate view id")
}