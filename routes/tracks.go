package routes

import (
	"context"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
)

type TracksRepo interface {
	ListOrderByTime(ctx context.Context, userId string) ([]db.Track, error)
	Import(ctx context.Context, ownerID string, filename string, data []byte) (string, error)
}

func registerTracksRoutes(
	r gin.IRouter,
	repo TracksRepo,
) {
	r.GET("/tracks/my", getMyTracks(repo))
	r.POST("/tracks/import", postImportTrack(repo))
}

func getMyTracks(repo TracksRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := getUserID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		orderBy := c.Query("orderBy")
		if orderBy != "time" {
			c.JSON(400, gin.H{"error": "Invalid orderBy parameter"})
			return
		}

		tracks, err := repo.ListOrderByTime(c.Request.Context(), userId)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{
			"data": tracks,
		})
	}
}

func postImportTrack(repo TracksRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := getUserID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(400, gin.H{"error": "Invalid form data"})
			return
		}
		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(400, gin.H{"error": "No files uploaded"})
			return
		}

		var ids []string
		for _, file := range files {
			id, err := importTrackFile(c.Request.Context(), repo, userId, *file)
			if err != nil {
				c.JSON(500, gin.H{"error": "Internal server error"})
				return
			}
			ids = append(ids, id)
		}

		c.JSON(200, gin.H{
			"data": gin.H{
				"imports": ids,
			},
		})
	}
}

func importTrackFile(ctx context.Context, repo TracksRepo, userId string, file multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", err
	}
	id, err := repo.Import(ctx, userId, file.Filename, data)
	if err != nil {
		return "", err
	}
	return id, nil
}
