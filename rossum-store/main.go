package main

import (
	"rossum-store/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/versions/:extension", handlers.GetVersionsHandler)
	r.GET("/checkout/:extension/:version", handlers.GetCheckoutHandler)
	r.GET("/extensions", handlers.GetStoreHandler)
	r.POST("/webhook", handlers.PostWebhook)

	r.Run(":8080") // Run server on port 8080
}
