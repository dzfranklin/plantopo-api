package routes

import (
	"github.com/dzfranklin/plantopo-api/meta"
	"github.com/gin-gonic/gin"
)

func registerMetaRoutes(r gin.IRouter) {
	r.GET("/_info", func(context *gin.Context) {
		context.String(200, meta.FullInfo)
	})
}
