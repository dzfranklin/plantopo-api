-- name: ListTracksOrderByTime :many
SELECT *
FROM tracks
WHERE owner_id = $1
ORDER BY time;

-- name: InsertTrackImport :one
INSERT INTO track_imports (owner_id, filename, data)
VALUES ($1, $2, $3)
RETURNING id;

-- name: MarkTrackImportCompleted :exec
UPDATE track_imports
SET completed_at = NOW()
WHERE id = $1;

-- name: MarkTrackImportFailed :exec
UPDATE track_imports
SET failed_at = NOW(), error = $2
WHERE id = $1;

-- name: GetTrackImport :one
SELECT *
FROM track_imports
WHERE id = $1;

-- name: GetTrackImportStatus :one
SELECT inserted_at, completed_at, failed_at, error
FROM track_imports
WHERE id = $1;

-- name: InsertImportedTrack :one
INSERT INTO tracks
    (owner_id, name, upload_time, time, geojson, import_id)
VALUES
    ($1, $2, $3, $4, $5, $6)
RETURNING id;
