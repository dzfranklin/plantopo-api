package routes

import "github.com/gin-gonic/gin"

func registerHealthRoute(r gin.IRouter) {
	r.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})
}
