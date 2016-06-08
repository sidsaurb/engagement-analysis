package resources

import (
	"strconv"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gpahal/veead/db"
	"github.com/gpahal/veea/conf"
)

type DashboardStats struct {
	TotalViews int64 `json:"totalViews"`
	UniqueVisitors int64 `json:"uniqueVisitors"`
	AvgViewDurationPresent bool `json:"avgViewDurationPresent"`
	AvgViewDuration float64 `json:"avgViewDuration"`
	MaleCount int64 `json:"maleCount"`
	FemaleCount int64 `json:"femaleCount"`
	AgeCounts []int64 `json:"ageCounts"`
	Stats []float64 `json:"stats"`
	InstantStats map[string][]float64 `json:"instantStats"`
	InstantViewedCount map[string]int64 `json:"instantViewedCount"`
}

func GetDashboardHandler(c *gin.Context) {
	account := GetUser(c)
	video := GetVideo(c)

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"Account": account,
		"Video": video,
	})
}

func DashboardDataHandler(c *gin.Context) {
	video := GetVideo(c)
	var form struct {
		VideoDuration float64 `form:"videoDuration" json:"videoDuration" binding:"required"`
	}

	var ds DashboardStats

	if c.Bind(&form) == nil {
		if form.VideoDuration < 0 || form.VideoDuration > float64(conf.ViewExpireTime) {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		totalViews, err := db.GetTotalViews(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.TotalViews = totalViews

		uniqueVisitors, err := db.GetUniqueVisitors(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.UniqueVisitors = uniqueVisitors

		avgViewDuration, successful, err := db.GetAverageViewDuration(video.VideoId, form.VideoDuration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.AvgViewDurationPresent = successful
		ds.AvgViewDuration = avgViewDuration

		maleCount, err := db.GetMaleCount(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.MaleCount = maleCount

		femaleCount, err := db.GetFemaleCount(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.FemaleCount = femaleCount

		ageCounts, err := db.GetAgeCounts(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.AgeCounts = ageCounts

		//maxEngagement, err := db.GetMaxEngagement()
		//if err != nil {
		//	fmt.Printf("err: %s\n", err.Error());
		//	c.JSON(http.StatusInternalServerError, gin.H{})
		//	return
		//}

		stats, err := db.GetStats(video.VideoId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		//if maxEngagement <= 0 {
		//	stats[7] = 0
		//} else {
		//	stats[7] = stats[7] / maxEngagement
		//}
		ds.Stats = stats

		ds.InstantStats = make(map[string][]float64)
		ds.InstantViewedCount = make(map[string]int64)

		var time float64
		for time < form.VideoDuration {
			startTime := time
			endTime := time + 5
			if endTime > form.VideoDuration {
				endTime = form.VideoDuration
			}

			midTimeString := strconv.FormatFloat((startTime + endTime) / 2, 'f', -1, 64)

			instantStats, err := db.GetInstantStats(video.VideoId, startTime, endTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			ds.InstantStats[midTimeString] = instantStats

			instantViewedCount, err := db.GetInstantViewedCount(video.VideoId, startTime, endTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			ds.InstantViewedCount[midTimeString] = instantViewedCount

			time += 5
		}

		c.JSON(http.StatusOK, ds)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}

func GetDashboardSingleHandler(c *gin.Context) {
	account := GetUser(c)
	video := GetVideo(c)
	viewId := c.Param("viewId")

	c.HTML(http.StatusOK, "dashboard_single.html", gin.H{
		"Account": account,
		"Video": video,
		"ViewId": viewId,
	})
}

func DashboardSingleDataHandler(c *gin.Context) {
	video := GetVideo(c)
	viewId := c.Param("viewId")
	var form struct {
		VideoDuration float64 `form:"videoDuration" json:"videoDuration" binding:"required"`
	}

	var ds DashboardStats

	if c.Bind(&form) == nil {
		if form.VideoDuration < 0 || form.VideoDuration > float64(conf.ViewExpireTime) {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		err := db.VideoIdViewIdExists(video.VideoId, viewId)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{})
			return
		}

		ds.TotalViews = 1

		ds.UniqueVisitors = 1

		avgViewDuration, successful, err := db.GetAverageViewDurationSingle(viewId, form.VideoDuration)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.AvgViewDurationPresent = successful
		ds.AvgViewDuration = avgViewDuration

		maleCount, err := db.GetMaleCountSingle(viewId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.MaleCount = maleCount

		femaleCount, err := db.GetFemaleCountSingle(viewId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.FemaleCount = femaleCount

		ageCounts, err := db.GetAgeCountsSingle(viewId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		ds.AgeCounts = ageCounts

		//maxEngagement, err := db.GetMaxEngagement()
		//if err != nil {
		//	fmt.Printf("err: %s\n", err.Error());
		//	c.JSON(http.StatusInternalServerError, gin.H{})
		//	return
		//}

		stats, err := db.GetStatsSingle(viewId)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{})
			return
		}
		//if maxEngagement <= 0 {
		//	stats[7] = 0
		//} else {
		//	stats[7] = stats[7] / maxEngagement
		//}
		ds.Stats = stats

		ds.InstantStats = make(map[string][]float64)
		ds.InstantViewedCount = make(map[string]int64)

		var time float64
		for time < form.VideoDuration {
			startTime := time
			endTime := time + 5
			if endTime > form.VideoDuration {
				endTime = form.VideoDuration
			}

			midTimeString := strconv.FormatFloat((startTime + endTime) / 2, 'f', -1, 64)

			instantStats, err := db.GetInstantStatsSingle(viewId, startTime, endTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			ds.InstantStats[midTimeString] = instantStats

			instantViewedCount, err := db.GetInstantViewedCountSingle(viewId, startTime, endTime)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{})
				return
			}
			ds.InstantViewedCount[midTimeString] = instantViewedCount

			time += 5
		}

		c.JSON(http.StatusOK, ds)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{})
	}
}