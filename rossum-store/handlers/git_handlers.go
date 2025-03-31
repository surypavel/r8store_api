package handlers

import (
	"net/http"
	"rossum-store/services"

	"github.com/gin-gonic/gin"
)

type Settings struct {
	Repositories []string `json:"repositories"`
}

type WebhookPayload struct {
	Payload  map[string]interface{} `json:"payload"`
	Name     string                 `json:"rossum_authorization_token"`
	Settings Settings               `json:"settings"`
	BaseUrl  string                 `json:"base_url"`
	Hook     string                 `json:"hook"`
	Secrets  map[string]interface{} `json:"secrets"`
}

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

func PostWebhook(c *gin.Context) {
	var payload WebhookPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var store = payload.Settings.Repositories[0]

	if payload.Payload["name"] == "get_extension_list" {
		content, err := services.GetStoreHandler(payload.Settings.Repositories[0], "", "")

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, content)
		return
	}

	if payload.Payload["name"] == "get_extension_version" {
		content, err := services.GetVersions(store, payload.Payload["extension"].(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, content)
		return
	}

	if payload.Payload["name"] == "checkout_extension" {
		// TODO: actually add extension
		content, err := services.GetFileByVersion(store, payload.Payload["extension"].(string), payload.Payload["version"].(string), "", "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, content)
		return
	}
}
