package routes

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"strings"
)

type Authenticator interface {
	Verify(token string) (string, error)
}

const userIDKey = "userID"

func Auth(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		l := slog.With("requestID", RequestID(c))

		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			doBearerAuth(l, authenticator, authHeader, c)
		}
	}
}

func doBearerAuth(l *slog.Logger, authenticator Authenticator, header string, c *gin.Context) {
	if !strings.HasPrefix(header, "Bearer ") {
		c.AbortWithStatusJSON(401, gin.H{"error": "Authorization header must be a Bearer token"})
		return
	}
	token := strings.TrimPrefix(header, "Bearer ")

	userID, err := authenticator.Verify(token)
	if err != nil {
		l.Info("auth check failed", "error", err)
		c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
		return
	}

	l.Info("authenticated", "userID", userID)
	c.Set(userIDKey, userID)
}

func getUserID(c *gin.Context) (string, bool) {
	userID, ok := c.Get(userIDKey)
	if !ok {
		return "", false
	} else {
		return userID.(string), true
	}
}
