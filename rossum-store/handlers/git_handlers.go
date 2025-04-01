package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"rossum-store/services"

	"github.com/gin-gonic/gin"
)

type Settings struct {
	Repositories []string `json:"repositories"`
}

type WebhookPayload struct {
	PayloadAction string                  `json:"payload_action"`
	Payload       map[string]interface{}  `json:"payload"`
	Token         string                  `json:"rossum_authorization_token"`
	Settings      Settings                `json:"settings"`
	BaseUrl       string                  `json:"base_url"`
	Hook          string                  `json:"hook"`
	Command       string                  `json:"command"`
	Secrets       map[string]interface{}  `json:"secrets"`
	Form          *map[string]interface{} `json:"form,omitempty"`
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

	if payload.PayloadAction == "add_repository" {
		if payload.Form != nil {
			var url = (*payload.Form)["url"].(string)

			postBody, _ := json.Marshal(gin.H{
				"settings": gin.H{
					"repositories": []string{url},
				},
			})

			responseBody := bytes.NewBuffer(postBody)

			resp, err := http.NewRequest(http.MethodPatch, payload.Hook, responseBody)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			resp.Header.Add("Authorization", fmt.Sprintf("Bearer %s", payload.Token))
			resp.Header.Add("Accept", "application/json")
			resp.Header.Add("Content-Type", "application/json")

			response, err := http.DefaultClient.Do(resp)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			if response.Status == "200 OK" {
				jsonData := []byte(`{"intent":{"form": null,"info":{"message":"Repository was added."}}}`)
				c.Data(http.StatusOK, "application/json", jsonData)
				return
			}
		}

		jsonData := []byte(fmt.Sprintf(`{"intent":{"form":{"command": "%s", "schema":{"properties":{"url":{"type":"string"}}}}}}`, payload.Command))
		c.Data(http.StatusOK, "application/json", jsonData)
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
		content, err := services.GetFileByVersion(store, payload.Payload["extension"].(string), payload.Payload["version"].(string), "", "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, content)
		return
	}
}
