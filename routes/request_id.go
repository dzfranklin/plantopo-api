package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
)

const requestIDKey = "requestID"

func assignRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = ulid.Make().String()
		}
		c.Set(requestIDKey, requestID)
	}
}

func getRequestID(c *gin.Context) string {
	requestID, ok := c.Get(requestIDKey)
	if !ok {
		panic("requestID not set in context")
	}
	return requestID.(string)
}
