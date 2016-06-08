package resources

import (
	"fmt"
	"strconv"
	"time"
	"net/http"
	"errors"

	"github.com/gpahal/veead/db"
	"github.com/gin-gonic/gin"
)

func DefaultQueryInt(c *gin.Context, key string, defaultValue int) (int, error) {
	value, successful := c.GetQuery(key)
	if !successful {
		return defaultValue, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("expected an integer but got '%s'", value))
	}

	return intValue, nil
}

func HttpError(c *gin.Context, err error, code int) {
	http.Error(c.Writer, err.Error(), code)
	c.Abort()
}

func SetCookie(name string, value string, expireTime time.Time, c *gin.Context) {
	cookie := &http.Cookie{
		Name: name,
		Value: value,
		Expires: expireTime,
	}

	http.SetCookie(c.Writer, cookie)
}

func SetCookieOneMonth(name string, value string, c *gin.Context) {
	expireTime := time.Now().AddDate(0, 1, 0)
	SetCookie(name, value, expireTime, c)
}

func GetCookie(name string, c *gin.Context) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func ErrorPrefix(err error) string {
	switch err.(type) {
	case *db.UserError:
		return "Input Error"
	default:
		return "Internal Error"
	}
}

func ErrorString(err error) string {
	switch err.(type) {
	case *db.UserError:
		return "input error"
	default:
		return "internal error"
	}
}

func SetUser(c *gin.Context, user *db.User) {
	c.Set("user", user)
}

func SetVideo(c *gin.Context, video *db.Video) {
	c.Set("video", video)
}

func SetPath(c *gin.Context, path string) {
	c.Set("path", path)
}

func GetUser(c *gin.Context) *db.User {
	user, exists := c.Get("user")
	if !exists {
		return nil
	}

	return user.(*db.User)
}

func GetVideo(c *gin.Context) *db.Video {
	video, exists := c.Get("video")
	if !exists {
		return nil
	}

	return video.(*db.Video)
}

func GetPath(c *gin.Context) string {
	path, exists := c.Get("path")
	if !exists {
		return ""
	}

	return path.(string)
}

func StringToInt64Unsafe(s string) int64 {
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}

	return val
}