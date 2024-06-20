package routes

import (
	"github.com/dzfranklin/plantopo-api/analysis"
	"github.com/gin-gonic/gin"
)

func Router(
	authenticator Authenticator,
	tracks TracksRepo,
	elevation analysis.ElevationQuerier,
) *gin.Engine {
	r := gin.New()
	r.Use(assignRequestID())
	r.Use(loggerMiddleware())
	r.Use(gin.Recovery())
	r.Use(Auth(authenticator))
	base := r.Group("/api/v1")
	registerTracksRoutes(base, tracks)
	registerElevationRoute(base, elevation)
	return r
}
