package routes

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type SettingsRepo interface {
	GetUnitSettings(ctx context.Context, userID string) (json.RawMessage, error)
	SetUnitSettings(ctx context.Context, userID string, value json.RawMessage) error
}

func registerSettingsRoutes(r gin.IRouter, repo SettingsRepo) {
	r.GET("/settings/units", getUnitSettings(repo))
	r.POST("/settings/units", setUnitSettings(repo))
}

func getUnitSettings(repo SettingsRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		settings, err := repo.GetUnitSettings(c.Request.Context(), userID)
		if err != nil {
			slog.Error("get unit settings", "error", err)
			c.JSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		c.JSON(200, gin.H{"data": settings})
	}
}

func setUnitSettings(repo SettingsRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, ok := getUserID(c)
		if !ok {
			c.JSON(401, gin.H{"error": "unauthorized"})
			return
		}

		var value json.RawMessage
		if err := c.BindJSON(&value); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if err := repo.SetUnitSettings(c.Request.Context(), userID, value); err != nil {
			slog.Error("set unit settings", "error", err)
			c.JSON(500, gin.H{"error": "Internal Server Error"})
			return
		}
		c.JSON(200, gin.H{"data": gin.H{}})
	}
}
