// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: query.sql

package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paulmach/orb/geojson"
)

const getTrackImport = `-- name: GetTrackImport :one
SELECT id, owner_id, inserted_at, completed_at, failed_at, error, filename, data
FROM track_imports
WHERE id = $1
`

func (q *Queries) GetTrackImport(ctx context.Context, id int64) (TrackImport, error) {
	row := q.db.QueryRow(ctx, getTrackImport, id)
	var i TrackImport
	err := row.Scan(
		&i.ID,
		&i.OwnerID,
		&i.InsertedAt,
		&i.CompletedAt,
		&i.FailedAt,
		&i.Error,
		&i.Filename,
		&i.Data,
	)
	return i, err
}

const getTrackImportStatus = `-- name: GetTrackImportStatus :one
SELECT inserted_at, completed_at, failed_at, error
FROM track_imports
WHERE id = $1
`

type GetTrackImportStatusRow struct {
	InsertedAt  pgtype.Timestamp `json:"insertedAt"`
	CompletedAt pgtype.Timestamp `json:"completedAt"`
	FailedAt    pgtype.Timestamp `json:"failedAt"`
	Error       *string          `json:"error"`
}

func (q *Queries) GetTrackImportStatus(ctx context.Context, id int64) (GetTrackImportStatusRow, error) {
	row := q.db.QueryRow(ctx, getTrackImportStatus, id)
	var i GetTrackImportStatusRow
	err := row.Scan(
		&i.InsertedAt,
		&i.CompletedAt,
		&i.FailedAt,
		&i.Error,
	)
	return i, err
}

const insertImportedTrack = `-- name: InsertImportedTrack :one
INSERT INTO tracks
    (owner_id, name, upload_time, time, geojson, import_id)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING id
`

type InsertImportedTrackParams struct {
	OwnerID    *string          `json:"ownerID"`
	Name       *string          `json:"name"`
	UploadTime pgtype.Timestamp `json:"uploadTime"`
	Time       pgtype.Timestamp `json:"time"`
	Geojson    geojson.Feature  `json:"geojson"`
	ImportID   *int64           `json:"importID"`
}

func (q *Queries) InsertImportedTrack(ctx context.Context, arg InsertImportedTrackParams) (int64, error) {
	row := q.db.QueryRow(ctx, insertImportedTrack,
		arg.OwnerID,
		arg.Name,
		arg.UploadTime,
		arg.Time,
		arg.Geojson,
		arg.ImportID,
	)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const insertTrackImport = `-- name: InsertTrackImport :one
INSERT INTO track_imports (owner_id, filename, data)
VALUES ($1, $2, $3)
RETURNING id
`

type InsertTrackImportParams struct {
	OwnerID  string `json:"ownerID"`
	Filename string `json:"filename"`
	Data     []byte `json:"data"`
}

func (q *Queries) InsertTrackImport(ctx context.Context, arg InsertTrackImportParams) (int64, error) {
	row := q.db.QueryRow(ctx, insertTrackImport, arg.OwnerID, arg.Filename, arg.Data)
	var id int64
	err := row.Scan(&id)
	return id, err
}

const listTracksOrderByTime = `-- name: ListTracksOrderByTime :many
SELECT id, owner_id, name, upload_time, time, geojson, import_id
FROM tracks
WHERE owner_id = $1
ORDER BY time
`

func (q *Queries) ListTracksOrderByTime(ctx context.Context, ownerID *string) ([]Track, error) {
	rows, err := q.db.Query(ctx, listTracksOrderByTime, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Track{}
	for rows.Next() {
		var i Track
		if err := rows.Scan(
			&i.ID,
			&i.OwnerID,
			&i.Name,
			&i.UploadTime,
			&i.Time,
			&i.Geojson,
			&i.ImportID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const markTrackImportCompleted = `-- name: MarkTrackImportCompleted :exec
UPDATE track_imports
SET completed_at = NOW()
WHERE id = $1
`

func (q *Queries) MarkTrackImportCompleted(ctx context.Context, id int64) error {
	_, err := q.db.Exec(ctx, markTrackImportCompleted, id)
	return err
}

const markTrackImportFailed = `-- name: MarkTrackImportFailed :exec
UPDATE track_imports
SET failed_at = NOW(), error = $2
WHERE id = $1
`

type MarkTrackImportFailedParams struct {
	ID    int64   `json:"id"`
	Error *string `json:"error"`
}

func (q *Queries) MarkTrackImportFailed(ctx context.Context, arg MarkTrackImportFailedParams) error {
	_, err := q.db.Exec(ctx, markTrackImportFailed, arg.ID, arg.Error)
	return err
}