package routes

import (
	"github.com/dzfranklin/plantopo-api/analysis"
	"github.com/gin-gonic/gin"
	"github.com/paulmach/orb"
)

func registerElevationRoute(
	r gin.IRouter,
	querier analysis.ElevationQuerier,
) {
	r.POST("/elevation", getElevation(querier))
}

func getElevation(querier analysis.ElevationQuerier) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Points orb.LineString `json:"points" binding:"required"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		elevations, err := querier.QueryElevations(c.Request.Context(), payload.Points)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{
			"data": gin.H{
				"elevations": elevations,
			},
		})
	}
}
