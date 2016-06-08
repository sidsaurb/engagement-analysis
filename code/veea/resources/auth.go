package resources

import (
	"time"
	"net/http"

	"github.com/gpahal/veea/db"
	"github.com/gin-gonic/gin"
)

func GetRegisterHandler(c *gin.Context) {
	path := GetPath(c)

	c.HTML(http.StatusOK, "register.html", gin.H{
		"Path"   : path,
		"Message": "",
	})
}

func RegisterHandler(c *gin.Context) {
	path := GetPath(c)
	var form struct {
		Username   string `form:"username" binding:"required"`
		FullName   string `form:"fullname" binding:"required"`
		Password   string `form:"password" binding:"required"`
		RePassword string `form:"repassword" binding:"required"`
	}

	if c.Bind(&form) == nil {
		if form.Password != form.RePassword {
			c.HTML(http.StatusOK, "register.html", gin.H{
				"Path"   : path,
				"Message": "Input Error: Password and repeat password do not match",
			})
			return
		}

		_, err := db.CreateUser(form.Username, form.Password, form.FullName)
		if err != nil {
			c.HTML(http.StatusOK, "register.html", gin.H{
				"Path"   : path,
				"Message": ErrorPrefix(err) + ": " + err.Error(),
			})
			return
		}

		c.Redirect(http.StatusFound, path + "/login")
	} else {
		c.HTML(http.StatusOK, "register.html", gin.H{
			"Path"   : path,
			"Message": "Input Error: invalid input entries",
		})
		return
	}
}

func GetLoginHandler(c *gin.Context) {
	path := GetPath(c)

	sid, err := GetCookie("sid", c)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Path"   : path,
			"Message": "",
		})
		return
	}
	if sid == "" {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Path"   : path,
			"Message": "",
		})
		return
	}

	_, successful, err := db.IsLoggedInSessionId(sid)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Path"   : path,
			"Message": "",
		})
		return
	}
	if !successful {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Path"   : path,
			"Message": "",
		})
		return
	}

	c.Redirect(http.StatusFound, path + "/watch")
}

func LoginHandler(c *gin.Context) {
	path := GetPath(c)
	var form struct {
		Username string `form:"username" binding:"required"`
		Password string `form:"password" binding:"required"`
	}

	if c.Bind(&form) == nil {
		sid, successful, err := db.Login(form.Username, form.Password)
		if err != nil {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Path"   : path,
				"Message": ErrorPrefix(err) + ": " + err.Error(),
			})
			return
		} else if !successful {
			c.HTML(http.StatusOK, "login.html", gin.H{
				"Path"   : path,
				"Message": "Input Error: Login unsuccessful - check username and password",
			})
			return
		}

		SetCookieOneMonth("sid", sid, c)
		c.Redirect(http.StatusFound, path + "/watch")
	} else {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"Path"   : path,
			"Message": "Input Error: invalid input entries",
		})
		return
	}
}

func GetLogoutHandler(c *gin.Context) {
	path := GetPath(c)

	sid, err := GetCookie("sid", c)
	if err != nil {
		c.Redirect(http.StatusFound, path + "/login")
		return
	}

	err = db.Logout(sid)
	if err != nil {
		c.HTML(http.StatusOK, "logout.html", gin.H{
			"Path"   : path,
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	SetCookie("sid", "", time.Unix(0, 0), c)
	c.Redirect(http.StatusFound, path + "/login")
}

func AuthMiddleware(c *gin.Context) {
	path := GetPath(c)

	sid, err := GetCookie("sid", c)
	if err != nil {
		c.Redirect(http.StatusFound, path + "/login")
		return
	}

	userId, successful, err := db.IsLoggedInSessionId(sid)
	if err != nil {
		c.HTML(http.StatusOK, "auth_error.html", gin.H{
			"Path"   : path,
		})
		return
	}
	if !successful {
		c.Redirect(http.StatusFound, path + "/login")
		return
	}

	user, err := db.GetUser(userId, userId)
	if err != nil || user == nil {
		c.HTML(http.StatusOK, "auth_error.html", gin.H{
			"Path"   : path,
		})
		return
	}

	SetUser(c, user)
	c.Next()
}