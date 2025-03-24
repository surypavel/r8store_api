package handlers

import (
	"net/http"
	"rossum-store/services"

	"github.com/gin-gonic/gin"
)

func GetCheckoutHandler(c *gin.Context) {
	extension := c.Param("extension")
	version := c.Param("version")
	store := c.Query("store")
	username, password, err := GetBasicAuth(c.GetHeader("Authorization"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	content, err := services.GetFileByVersion(store, extension, version, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

func GetStoreHandler(c *gin.Context) {
	store := c.Query("store")

	username, password, err := GetBasicAuth(c.GetHeader("Authorization"))

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	content, err := services.GetStoreHandler(store, username, password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}

func GetVersionsHandler(c *gin.Context) {
	store := c.Query("store")
	extension := c.Param("extension")

	content, err := services.GetVersions(store, extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, content)
}
