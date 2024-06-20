package routes

import (
	"context"
	"errors"
	"github.com/dzfranklin/plantopo-api/tracks"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
)

type TracksRepo interface {
	Get(ctx context.Context, id string) (tracks.Track, error)
	Delete(ctx context.Context, id string) error
	IsOwner(ctx context.Context, userId string, trackId string) (bool, error)
	ListMyTracksOrderByTime(ctx context.Context, userId string) ([]tracks.Track, error)
	Import(ctx context.Context, ownerID string, filename string, data []byte) (string, error)
	ListMyPendingOrRecentImports(ctx context.Context, userID string) ([]tracks.Import, error)
}

const maxImportSize = 10 * 1024 * 1024 // 10MB

func registerTracksRoutes(
	r gin.IRouter,
	repo TracksRepo,
) {
	r.GET("/tracks/:id", getTrack(repo))
	r.DELETE("/tracks/:id", deleteTrack(repo))
	r.GET("/tracks/my", getMyTracks(repo))
	r.GET("/tracks/import/my/pending-or-recent", getMyPendingOrRecentImports(repo))
	r.POST("/tracks/import", postImportTrack(repo))
}

func getTrack(repo TracksRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := getUserID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		trackId := c.Param("id")
		track, err := repo.Get(c.Request.Context(), trackId)
		if err != nil {
			if errors.Is(err, tracks.ErrTrackNotFound) {
				c.JSON(404, gin.H{"error": "Track not found"})
				return
			}
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		isOwner, err := repo.IsOwner(c.Request.Context(), userId, trackId)
		if err != nil {
			if errors.Is(err, tracks.ErrTrackNotFound) {
				c.JSON(404, gin.H{"error": "Track not found"})
				return
			}
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if !isOwner {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.JSON(200, gin.H{
			"data": track,
		})
	}
}

func deleteTrack(repo TracksRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := getUserID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		trackId := c.Param("id")
		isOwner, err := repo.IsOwner(c.Request.Context(), userId, trackId)
		if err != nil {
			if errors.Is(err, tracks.ErrTrackNotFound) {
				c.JSON(404, gin.H{"error": "Track not found"})
				return
			}
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		if !isOwner {
			c.JSON(403, gin.H{"error": "Forbidden"})
			return
		}

		err = repo.Delete(c.Request.Context(), trackId)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{"data": gin.H{}})
	}
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

		data, err := repo.ListMyTracksOrderByTime(c.Request.Context(), userId)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{
			"data": data,
		})
	}
}

func getMyPendingOrRecentImports(repo TracksRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		userId, ok := getUserID(c)
		if !ok {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}

		data, err := repo.ListMyPendingOrRecentImports(c.Request.Context(), userId)
		if err != nil {
			c.JSON(500, gin.H{"error": "Internal server error"})
			return
		}

		c.JSON(200, gin.H{
			"data": data,
		})
	}
}

type successfulTrackImportRequest struct {
	ID       string `json:"id"`
	Filename string `json:"filename"`
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
			slog.Info("error parsing form", "error", err)
			c.JSON(400, gin.H{"error": "Invalid form data"})
			return
		}
		files := form.File["files"]
		if len(files) == 0 {
			c.JSON(400, gin.H{"error": "No files uploaded"})
			return
		}

		var imports []successfulTrackImportRequest
		for _, file := range files {
			id, err := importTrackFile(c.Request.Context(), c.Writer, repo, userId, *file)
			if err != nil {
				c.JSON(500, gin.H{"error": "Internal server error"})
				return
			}
			imports = append(imports, successfulTrackImportRequest{
				ID:       id,
				Filename: file.Filename,
			})
		}

		c.JSON(200, gin.H{
			"data": gin.H{
				"imports": imports,
			},
		})
	}
}

func importTrackFile(ctx context.Context, w http.ResponseWriter, repo TracksRepo, userId string, file multipart.FileHeader) (string, error) {
	f, err := file.Open()
	if err != nil {
		return "", err
	}
	defer f.Close()
	data, err := io.ReadAll(http.MaxBytesReader(w, f, maxImportSize))
	if err != nil {
		return "", err
	}
	id, err := repo.Import(ctx, userId, file.Filename, data)
	if err != nil {
		return "", err
	}
	return id, nil
}
