package resources

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gpahal/veead/db"
)

func GetIndexHandler(c *gin.Context) {
	path := GetPath(c)

	c.Redirect(http.StatusFound, path + "/dashboard")
}

func GetVideosHandler(c *gin.Context)  {
	account := GetUser(c)

	videos, err := db.GetVideos(account.Id)
	if err != nil {
		c.HTML(http.StatusOK, "videos.html", gin.H{
			"Account": account,
			"Message": ErrorPrefix(err) + ": " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "videos.html", gin.H{
		"Account": account,
		"Message": c.Query("msg"),
		"Videos": videos,
	})
}

func AddVideoHandler(c *gin.Context) {
	account := GetUser(c)
	var form struct {
		VideoId string `form:"videoid" binding:"required"`
		Name    string `form:"name"`
	}

	if c.Bind(&form) == nil {
		err := db.AddVideo(account.Id, form.VideoId, form.Name)
		if err != nil {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/videos?msg=%s", url.QueryEscape("Unable to add video (" + ErrorString(err) + ")")))
			return
		}

		c.Redirect(http.StatusFound, "/admin/videos")
	} else {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/videos?msg=%s", url.QueryEscape("Unable to add video (input error)")))
	}
}

func UpdateVideoHandler(c *gin.Context) {
	account := GetUser(c)
	videoId := c.Param("videoId")
	var form struct {
		Name    string `form:"name"`
	}

	if c.Bind(&form) == nil {
		err := db.UpdateVideo(account.Id, videoId, form.Name)
		if err != nil {
			c.Redirect(http.StatusFound, fmt.Sprintf("/admin/videos?msg=%s", url.QueryEscape("Unable to update video (" + ErrorString(err) + ")")))
			return
		}

		c.Redirect(http.StatusFound, "/admin/videos")
	} else {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/videos?msg=%s", url.QueryEscape("Unable to update video (input error)")))
	}
}

func DeleteVideoHandler(c *gin.Context) {
	account := GetUser(c)
	videoId := c.Param("videoId")

	err := db.DeleteVideo(account.Id, videoId)
	if err != nil {
		c.Redirect(http.StatusFound, fmt.Sprintf("/admin/videos?msg=%s", url.QueryEscape("Unable to delete video (" + ErrorString(err) + ")")))
		return
	}

	c.Redirect(http.StatusFound, "/admin/videos")
}

func VideoMiddleware(c *gin.Context) {
	videoId := c.Param("videoId")
	path := fmt.Sprintf("/admin/video/%s", videoId)

	video, err := db.GetVideo(videoId)
	if err != nil {
		switch err.(type) {
		case *db.UserError:
			c.HTML(http.StatusOK, "user_video_error.html", gin.H{
				"Path"   : path,
			})
			return
		default:
			c.HTML(http.StatusOK, "internal_error.html", gin.H{
				"Path"   : path,
			})
			return
		}
	}
	if video == nil {
		c.HTML(http.StatusOK, "internal_error.html", gin.H{
			"Path"   : path,
		})
		return
	}

	SetVideo(c, video)
	SetPath(c, path)
	c.Next()
}