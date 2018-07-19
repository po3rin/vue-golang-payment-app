package infrastructure

import (
	"github.com/gin-contrib/cors"
	gin "github.com/gin-gonic/gin"
)

// Router - router api server
var Router *gin.Engine

func init() {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3030"},
		AllowMethods: []string{"GET", "POST", "DELETE"},
		AllowHeaders: []string{"Origin"},
	}))

	router.POST("/api/v1/items", func(c *gin.Context) {})
	router.GET("/api/v1/items", func(c *gin.Context) {})
	router.GET("/api/v1/items/:id", func(c *gin.Context) {})
	router.DELETE("/api/v1/items/:id", func(c *gin.Context) {})

	Router = router
}
