-- name: GetTrack :one
SELECT *
FROM tracks
WHERE id = $1;

-- name: DeleteTrack :exec
DELETE
FROM tracks
WHERE id = $1;

-- name: GetTrackOwner :one
SELECT owner_id
FROM tracks
WHERE id = $1;

-- name: ListTracksOrderByTime :many
SELECT *
FROM tracks
WHERE owner_id = $1
ORDER BY time DESC;


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
SET failed_at = NOW(),
    error     = $2
WHERE id = $1;

-- name: GetTrackImport :one
SELECT *
FROM track_imports
WHERE id = $1;

-- name: ListMyPendingOrRecentImports :many
SELECT id,
       owner_id,
       inserted_at,
       completed_at,
       failed_at,
       error,
       filename,
       length(data) as byte_size
FROM track_imports
WHERE owner_id = $1
  AND (completed_at IS NULL OR
       completed_at > NOW() - INTERVAL '1 DAY' OR
       failed_at > NOW() - INTERVAL '1 DAY')
ORDER BY inserted_at DESC;

-- name: GetTrackImportStatus :one
SELECT id,
       owner_id,
       inserted_at,
       completed_at,
       failed_at,
       error,
       filename,
       length(data) as byte_size
FROM track_imports
WHERE id = $1;

-- name: InsertImportedTrack :one
INSERT INTO tracks
    (owner_id, name, upload_time, time, geojson, import_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
