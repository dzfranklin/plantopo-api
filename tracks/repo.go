package tracks

import (
	"context"
	"errors"
	"fmt"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/dzfranklin/plantopo-api/ids"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paulmach/orb/geojson"
	"github.com/riverqueue/river"
	"log/slog"
	"time"
)

const (
	maxImportSize = 10 * 1024 * 1024

	trackIdPrefix  = "t"
	importIdPrefix = "ti"
)

var ErrTrackNotFound = fmt.Errorf("track not found")

type Repo struct {
	pool  *pgxpool.Pool
	river *river.Client[pgx.Tx]
	q     *db.Queries
}

func NewRepo(pool *pgxpool.Pool, river *river.Client[pgx.Tx]) *Repo {
	return &Repo{pool: pool, q: db.New(pool), river: river}
}

type Track struct {
	ID         string          `json:"id"`
	OwnerID    string          `json:"ownerID,omitempty"`
	Name       string          `json:"name,omitempty"`
	UploadTime time.Time       `json:"uploadTime"`
	Time       *time.Time      `json:"time,omitempty"`
	Geojson    geojson.Feature `json:"geojson"`
	ImportID   string          `json:"importID,omitempty"`
}

type Import struct {
	ID          string     `json:"id"`
	OwnerID     string     `json:"ownerID"`
	StartedAt   time.Time  `json:"startedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	FailedAt    *time.Time `json:"failedAt,omitempty"`
	Error       string     `json:"error,omitempty"`
	Filename    string     `json:"filename"`
	ByteSize    int        `json:"byteSize"`
}

func (r *Repo) Get(ctx context.Context, id string) (Track, error) {
	tid, err := ids.Unmarshal(trackIdPrefix, id)
	if err != nil {
		return Track{}, err
	}
	track, err := r.q.GetTrack(ctx, tid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Track{}, ErrTrackNotFound
		}
		return Track{}, err
	}
	return toTrack(track), nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	tid, err := ids.Unmarshal(trackIdPrefix, id)
	if err != nil {
		return err
	}
	return r.q.DeleteTrack(ctx, tid)
}

func (r *Repo) IsOwner(ctx context.Context, userId string, trackId string) (bool, error) {
	tid, err := ids.Unmarshal(trackIdPrefix, trackId)
	if err != nil {
		return false, err
	}

	owner, err := r.q.GetTrackOwner(ctx, tid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, ErrTrackNotFound
		}
		return false, err
	}
	if owner == nil {
		return false, nil
	}

	return *owner == userId, nil
}

func (r *Repo) ListMyTracksOrderByTime(ctx context.Context, userID string) ([]Track, error) {
	tracks, err := r.q.ListTracksOrderByTime(ctx, &userID)
	if err != nil {
		return nil, err
	}
	out := make([]Track, 0)
	for _, t := range tracks {
		out = append(out, toTrack(t))
	}
	return out, nil
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
		_ = tx.Rollback(ctx)
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

func (r *Repo) ImportStatus(ctx context.Context, id string) (Import, error) {
	importId, err := ids.Unmarshal(importIdPrefix, id)
	if err != nil {
		return Import{}, err
	}

	data, err := r.q.GetTrackImportStatus(ctx, importId)
	if err != nil {
		return Import{}, err
	}

	return Import{
		ID:          ids.Marshal(importIdPrefix, data.ID),
		OwnerID:     data.OwnerID,
		StartedAt:   data.InsertedAt.Time,
		CompletedAt: pgTimestampToNullable(data.CompletedAt),
		FailedAt:    pgTimestampToNullable(data.FailedAt),
		Error:       stringFromNullable(data.Error),
		Filename:    data.Filename,
		ByteSize:    int(data.ByteSize),
	}, nil
}

func (r *Repo) ListMyPendingOrRecentImports(ctx context.Context, userID string) ([]Import, error) {
	imports, err := r.q.ListMyPendingOrRecentImports(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]Import, 0)
	for _, i := range imports {
		out = append(out, Import{
			ID:          ids.Marshal(importIdPrefix, i.ID),
			OwnerID:     i.OwnerID,
			StartedAt:   i.InsertedAt.Time,
			CompletedAt: pgTimestampToNullable(i.CompletedAt),
			FailedAt:    pgTimestampToNullable(i.FailedAt),
			Error:       stringFromNullable(i.Error),
			Filename:    i.Filename,
			ByteSize:    int(i.ByteSize),
		})
	}
	return out, nil
}

func toTrack(data db.Track) Track {
	return Track{
		ID:         ids.Marshal(trackIdPrefix, data.ID),
		OwnerID:    stringFromNullable(data.OwnerID),
		Name:       stringFromNullable(data.Name),
		UploadTime: data.UploadTime.Time,
		Time:       pgTimestampToNullable(data.Time),
		Geojson:    data.Geojson,
		ImportID:   ids.MarshalNullable(importIdPrefix, data.ImportID),
	}
}

func pgTimestampToNullable(t pgtype.Timestamp) *time.Time {
	if t.Valid {
		return &t.Time
	}
	return nil
}

func stringFromNullable(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
