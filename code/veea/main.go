package main

import (
	"time"
	"flag"
	"net/http"

	"github.com/gpahal/veea/resources"
	"github.com/gin-gonic/gin"
	"github.com/gpahal/veea/db"
	log "github.com/Sirupsen/logrus"
	"github.com/gpahal/veea/conf"
)

func CheckPeriodically() {
	for {
		time.Sleep(5 * 60 * time.Second)
		err := db.UpdateVideoDuration()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("Error updating video duration")
		}
	}
}

func main() {
	go CheckPeriodically()

	router := gin.Default()

	router.LoadHTMLGlob(conf.BasePath + "templates/*")
	router.Static("/static", conf.BasePath + "static")

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	videoRouter := router.Group("/video/:videoId", resources.VideoMiddleware)
	{
		videoRouter.GET("/", resources.GetIndexHandler)

		videoRouter.GET("/register", resources.GetRegisterHandler)
		videoRouter.POST("/register", resources.RegisterHandler)

		videoRouter.GET("/login", resources.GetLoginHandler)
		videoRouter.POST("/login", resources.LoginHandler)

		videoRouter.GET("/logout", resources.GetLogoutHandler)

		videoRouter.GET("/watch", resources.AuthMiddleware, resources.GetVideoHandler)

		videoRouter.POST("/data", resources.AuthMiddleware, resources.GetDataHandler)
	}

	otherPtr := flag.Bool("other", false, "choose the other port")

	flag.Parse()

	if *otherPtr {
		router.Run(":8081")
	} else {
		router.Run(":8080")
	}
}
