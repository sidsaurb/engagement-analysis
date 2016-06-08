package resources

import (
	"fmt"
	"time"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gpahal/veea/conf"
	"github.com/gpahal/veea/db"
	"github.com/vincent-petithory/dataurl"
	"io/ioutil"
	"net/url"
	"strings"
	"encoding/json"
)

type Data struct {
	ViewId    string `json:"viewId" binding:"required"`
	Time      float64 `json:"time" binding:"required"`
	State     int `json:"state" binding:"required"`
	Quality   string `json:"quality" binding:"required"`
	ImageNull bool `json:"imageNull"`
	ImageData string `json:"imageData" binding:"required"`
}

type DataResult struct {
	Status int
	PeopleCount int
	PeopleSuccessCount int
}

func GetIndexHandler(c *gin.Context) {
	path := GetPath(c)

	c.Redirect(http.StatusFound, path + "/watch")
}

func GetVideoHandler(c *gin.Context)  {
	video := GetVideo(c)
	path := GetPath(c)
	user := GetUser(c)

	viewId, err := db.AddVideoView(user.Id, video.VideoId)
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

	c.HTML(http.StatusOK, "video.html", gin.H{
		"Path"   : path,
		"VideoId": video.VideoId,
		"ViewId" : viewId,
		"EndTime": (time.Now().Unix() + conf.ViewExpireTime) * 1000,
	})
}

func GetDataHandler(c *gin.Context) {
	video := GetVideo(c)
	var form Data

	dr := &DataResult{0, 0, 0}

	if c.BindJSON(&form) == nil {
		viewTime := &db.ViewTime{
			ViewId: form.ViewId,
			Time: form.Time,
			State: form.State,
			Quality: form.Quality,
		}

		viewTimeId, err := db.AddViewTime(video.VideoId, viewTime)
		if err != nil {
			SendDataResultJSON(c, http.StatusInternalServerError, dr)
			return
		}

		if form.ImageNull {
			dr.Status = 1
			SendDataResultJSON(c, http.StatusOK, dr)
			return
		}

		dr.Status = 2

		var viewStatsList []*db.ViewStats

		imageData, err := dataurl.DecodeString(form.ImageData)
		if err != nil {
			SendDataResultJSON(c, http.StatusBadRequest, dr)
			return
		}

		requestUrl := "http://52.77.220.121:9999"
		form := url.Values{}

		form.Add("file", string(imageData.Data))

		resp, err := http.Post(requestUrl, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
		if err != nil {
			SendDataResultJSON(c, http.StatusInternalServerError, dr)
			return
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			SendDataResultJSON(c, http.StatusInternalServerError, dr)
			return
		}

		var errorJson map[string]interface{}
		var respJson []map[string]interface{}

		err = json.Unmarshal(respBody, &errorJson)
		if err == nil {
			dr.Status = 3
			SendDataResultJSON(c, http.StatusOK, dr)
			return
		}

		err = json.Unmarshal(respBody, &respJson)
		if err != nil {
			SendDataResultJSON(c, http.StatusInternalServerError, dr)
			return
		}

		for _, singlePerson := range respJson {
			viewStats := &db.ViewStats{}
			viewStats.ViewTimeId = viewTimeId

			genderString := singlePerson["gender"].(string)
			if genderString == "male" {
				viewStats.Gender = -1;
			} else if genderString == "female" {
				viewStats.Gender = 1;
			} else {
				viewStats.Gender = 0;
			}

			viewStats.Age = int(singlePerson["age"].(float64))
			viewStats.Mood = singlePerson["mood"].(float64)

			headPose := singlePerson["headpose"].([]interface{})

			viewStats.HeadX = headPose[0].(float64)
			viewStats.HeadY = headPose[1].(float64)
			viewStats.HeadZ = headPose[2].(float64)
			viewStats.HeadYaw = headPose[3].(float64)
			viewStats.HeadPitch = headPose[4].(float64)
			viewStats.HeadRoll = headPose[5].(float64)

			headGaze := singlePerson["headgaze"].([]interface{})

			viewStats.HeadGazeX = headGaze[0].(float64)
			viewStats.HeadGazeY = headGaze[1].(float64)

			emotions := singlePerson["emotions"].([]interface{})

			viewStats.Happy = emotions[0].(float64)
			viewStats.Surprised = emotions[1].(float64)
			viewStats.Angry = emotions[2].(float64)
			viewStats.Disgusted = emotions[3].(float64)
			viewStats.Afraid = emotions[4].(float64)
			viewStats.Sad = emotions[5].(float64)

			viewStats.Engagement = Norm2(
				viewStats.HeadYaw / 0.2,
				viewStats.HeadPitch / 0.2,
				viewStats.HeadGazeX / 300,
				viewStats.HeadGazeY / 300,
			)

			viewStatsList = append(viewStatsList, viewStats)
		}

		dr.Status = 3

		count := 0
		for _, viewStats := range viewStatsList {
			viewStats.ViewTimeId = viewTimeId
			_, err = db.AddViewStats(viewStats)
			if err == nil {
				count += 1
			}
		}

		dr.PeopleCount = len(viewStatsList)
		dr.PeopleSuccessCount = count

		SendDataResultJSON(c, http.StatusOK, dr)
	} else {
		SendDataResultJSON(c, http.StatusBadRequest, dr)
	}
}

func VideoMiddleware(c *gin.Context) {
	videoId := c.Param("videoId")
	path := fmt.Sprintf("/video/%s", videoId)

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