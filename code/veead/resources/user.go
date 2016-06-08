package resources

import (
	"github.com/gin-gonic/gin"
	"github.com/gpahal/veead/db"
	"net/http"
)

func GetUsersHandler(c *gin.Context)  {
	account := GetUser(c)

	users, err := db.GetUsers(account.Id)
	if err != nil {
		c.HTML(http.StatusOK, "users.html", gin.H{
			"Account": account,
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "users.html", gin.H{
		"Account": account,
		"Message": "",
		"Users": users,
	})
}

func GetUserViewsHandler(c *gin.Context)  {
	account := GetUser(c)
	userId := StringToInt64Unsafe(c.Param("id"))

	user, err := db.GetUser(account.Id, userId)
	if err != nil {
		c.HTML(http.StatusOK, "user_views.html", gin.H{
			"Account": account,
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	views, err := db.GetViews(account.Id, userId)
	if err != nil {
		c.HTML(http.StatusOK, "user_views.html", gin.H{
			"Account": account,
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "user_views.html", gin.H{
		"Account": account,
		"Message": "",
		"User": user,
		"Views": views,
	})
}