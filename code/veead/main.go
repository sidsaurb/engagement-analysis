package main

import (
	"net/http"

	"github.com/gpahal/veead/db"
	"github.com/gpahal/veead/resources"
	"github.com/gin-gonic/gin"
	"github.com/gpahal/veead/conf"
)

func main() {
	err := db.CreateAdminUserIfNotExists()
	if err != nil {
		panic("Unable to create admin user")
	}

	router := gin.Default()

	router.LoadHTMLGlob(conf.BasePath + "templates/*")
	router.Static("/static", conf.BasePath + "static")

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	router.GET("/", func (c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})

	router.GET("/login", resources.GetLoginHandler)
	router.POST("/login", resources.LoginHandler)

	router.GET("/logout", resources.GetLogoutHandler)

	authRouter := router.Group("/admin", resources.AuthMiddleware)
	{
		authRouter.GET("/", func (c *gin.Context) {
			c.Redirect(http.StatusFound, "/admin/videos")
		})


		authRouter.GET("/users", resources.GetUsersHandler)
		authRouter.GET("/user/:id/views", resources.GetUserViewsHandler)

		authRouter.GET("/videos", resources.GetVideosHandler)

		authRouter.POST("/add_video", resources.AddVideoHandler)
		authRouter.POST("/update_video/:videoId", resources.UpdateVideoHandler)
		authRouter.POST("/delete_video/:videoId", resources.DeleteVideoHandler)

		videoRouter := authRouter.Group("/video/:videoId", resources.VideoMiddleware)
		{
			videoRouter.GET("/", resources.GetIndexHandler)

			videoRouter.GET("/all/dashboard", resources.GetDashboardHandler)
			videoRouter.POST("/all/dashboard_data", resources.DashboardDataHandler)

			videoRouter.GET("/single/:viewId/dashboard", resources.GetDashboardSingleHandler)
			videoRouter.POST("/single/:viewId/dashboard_data", resources.DashboardSingleDataHandler)
		}
	}

	router.Run(":8083")
}
