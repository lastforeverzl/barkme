package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	hub := newHub()
	go hub.run()

	router := gin.Default()

	// router.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"message": "Welcome to gin!",
	// 		"status":  http.StatusOK,
	// 	})
	// })
	router.LoadHTMLFiles("test.html")
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "test.html", nil)
	})

	router.GET("/ws", func(c *gin.Context) {
		wsHandler(hub, c.Writer, c.Request)
	})

	router.Run("localhost:8080")
}
