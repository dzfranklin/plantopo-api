package routes

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"time"
)

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := getRequestID(c)
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		elapsed := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		slog.Info("request",
			"requestID", requestID,
			"clientIP", clientIP,
			"method", method,
			"path", path,
			"statusCode", statusCode,
			"errorMessage", errorMessage,
			"elapsed", elapsed,
			"bodySize", bodySize,
		)
	}
}
