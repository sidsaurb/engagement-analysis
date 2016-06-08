package resources

import (
	"time"
	"net/http"

	"github.com/gpahal/veead/db"
	"github.com/gin-gonic/gin"
)

func GetLoginHandler(c *gin.Context) {
	sid, err := GetCookie("sid", c)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Message": "",
		})
		return
	}
	if sid == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Message": "",
		})
		return
	}

	_, successful, err := db.IsLoggedInSessionIdAdmin(sid)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Message": "",
		})
		return
	}
	if !successful {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Message": "",
		})
		return
	}

	c.Redirect(http.StatusFound, "/admin")
}

func LoginHandler(c *gin.Context) {
	var form struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if c.Bind(&form) == nil {
		sid, successful, err := db.LoginAdmin(form.Username, form.Password)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Message": ErrorPrefix(err) + ": " + err.Error(),
			})
			return
		} else if !successful {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Message": "Input Error: Login unsuccessful - check username and password",
			})
			return
		}

		SetCookieOneMonth("sid", sid, c)
		c.Redirect(http.StatusFound, "/admin")
	} else {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Message": "Input Error: invalid input entries",
		})
		return
	}
}

func GetLogoutHandler(c *gin.Context) {
	sid, err := GetCookie("sid", c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	err = db.LogoutAdmin(sid)
	if err != nil {
		c.HTML(http.StatusOK, "logout.html", gin.H{
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	SetCookie("sid", "", time.Unix(0, 0), c)
	c.Redirect(http.StatusFound, "/login")
}

func AuthMiddleware(c *gin.Context) {
	sid, err := GetCookie("sid", c)
	if err != nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	userId, successful, err := db.IsLoggedInSessionIdAdmin(sid)
	if err != nil {
		c.HTML(http.StatusOK, "auth_error.html", gin.H{})
		return
	}
	if !successful {
		c.Redirect(http.StatusFound, "/login")
		return
	}

	user, err := db.GetUser(userId, userId)
	if err != nil || user == nil {
		c.HTML(http.StatusOK, "auth_error.html", gin.H{})
		return
	}

	SetUser(c, user)
	c.Next()
}