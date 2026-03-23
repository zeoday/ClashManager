package handlers

import (
	"net/http"

	"clash-manager/internal/repository"
	"clash-manager/internal/service"

	"github.com/gin-gonic/gin"
)

type SubHandler struct {
	Service          *service.ConfigService
	SingBoxService   *service.SingBoxConfigService
	UserRepo         *repository.UserRepository
	SubHandlerHelper *SubscriptionHandler
}

func NewSubHandler() *SubHandler {
	return &SubHandler{
		Service:          service.NewConfigService(),
		SingBoxService:   service.NewSingBoxConfigService(),
		UserRepo:         &repository.UserRepository{},
		SubHandlerHelper: NewSubscriptionHandler(),
	}
}

func (h *SubHandler) GetConfig(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.String(http.StatusBadRequest, "Missing token")
		return
	}

	// Get format parameter (clash or singbox)
	format := c.DefaultQuery("format", "clash")

	// Validate token
	user, err := h.UserRepo.FindByToken(token)
	if err != nil {
		h.SubHandlerHelper.LogSubscription(0, token, c.ClientIP(), c.GetHeader("User-Agent"), false, "Invalid token")
		c.String(http.StatusUnauthorized, "Invalid token")
		return
	}

	var configBytes []byte
	var filename string
	var contentType string
	var profileTitle string

	// Generate config based on format
	switch format {
	case "singbox", "sing-box":
		configBytes, err = h.SingBoxService.GenerateConfig()
		filename = "singbox_config.json"
		contentType = "application/json; charset=utf-8"
		profileTitle = "SingBox-" + user.Username
	default:
		configBytes, err = h.Service.GenerateConfig()
		filename = "clash_config.yaml"
		contentType = "application/yaml; charset=utf-8"
		profileTitle = "Clash-" + user.Username
	}

	if err != nil {
		h.SubHandlerHelper.LogSubscription(user.ID, token, c.ClientIP(), c.GetHeader("User-Agent"), false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log successful subscription
	h.SubHandlerHelper.LogSubscription(user.ID, token, c.ClientIP(), c.GetHeader("User-Agent"), true, "")

	// Set headers
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename="+filename)
	// Set Subscription-Userinfo header for clients to display traffic info
	c.Header("Subscription-Userinfo", "upload=0; download=0; total=10737418240; expire=0")
	// Set profile name
	c.Header("Profile-Title", profileTitle)

	c.String(http.StatusOK, string(configBytes))
}
