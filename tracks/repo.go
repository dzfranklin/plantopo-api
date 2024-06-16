package tracks

import (
	"context"
	"fmt"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

const maxImportSize = 10 * 1024 * 1024

type Repo struct {
	pool  *pgxpool.Pool
	river *river.Client[pgx.Tx]
	q     *db.Queries
}

func NewRepo(pool *pgxpool.Pool, river *river.Client[pgx.Tx]) *Repo {
	return &Repo{pool: pool, q: db.New(pool), river: river}
}

func (r *Repo) ListOrderByTime(ctx context.Context, userId string) ([]db.Track, error) {
	return r.q.ListTracksOrderByTime(ctx, &userId)
}

func (r *Repo) Import(ctx context.Context, ownerID string, filename string, data []byte) (string, error) {
	if len(data) > maxImportSize {
		slog.Warn("import too large", "size", len(data), "max", maxImportSize)
		return "", fmt.Errorf("import too large")
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func(tx pgx.Tx, ctx context.Context) {
		err := tx.Rollback(ctx)
		if err != nil {
			slog.Error("rollback", "error", err)
		}
	}(tx, ctx)
	q := r.q.WithTx(tx)

	id, err := q.InsertTrackImport(ctx, db.InsertTrackImportParams{
		OwnerID:  ownerID,
		Filename: filename,
		Data:     data,
	})
	if err != nil {
		return "", err
	}

	_, err = r.river.InsertTx(ctx, tx, &ImportWorkerArgs{Id: id}, nil)
	if err != nil {
		return "", err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("trackimport_%d", id), nil
}

type ImportStatus struct {
	CreatedAt   time.Time  `json:"createdAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	FailedAt    *time.Time `json:"failedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
}

func (r *Repo) ImportStatus(ctx context.Context, id string) (ImportStatus, error) {
	importId, err := parseImportId(id)
	if err != nil {
		return ImportStatus{}, err
	}

	data, err := r.q.GetTrackImportStatus(ctx, importId)
	if err != nil {
		return ImportStatus{}, err
	}

	var completedAt *time.Time
	if data.CompletedAt.Valid {
		completedAt = &data.CompletedAt.Time
	}
	var failedAt *time.Time
	if data.FailedAt.Valid {
		failedAt = &data.FailedAt.Time
	}

	var importError string
	if data.Error != nil {
		importError = *data.Error
	}

	return ImportStatus{
		CreatedAt:   data.InsertedAt.Time,
		CompletedAt: completedAt,
		FailedAt:    failedAt,
		Error:       importError,
	}, nil
}

func parseImportId(id string) (int64, error) {
	if !strings.HasPrefix(id, "trackimport_") {
		return 0, fmt.Errorf("invalid id")
	}
	parsed, err := strconv.ParseInt(id[len("trackimport_"):], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id")
	}
	return parsed, nil
}
