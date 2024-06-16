// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/paulmach/orb/geojson"
)

type Track struct {
	ID         int64            `json:"id"`
	OwnerID    *string          `json:"ownerID"`
	Name       *string          `json:"name"`
	UploadTime pgtype.Timestamp `json:"uploadTime"`
	Time       pgtype.Timestamp `json:"time"`
	Geojson    geojson.Feature  `json:"geojson"`
	ImportID   *int64           `json:"importID"`
}

type TrackImport struct {
	ID          int64            `json:"id"`
	OwnerID     string           `json:"ownerID"`
	InsertedAt  pgtype.Timestamp `json:"insertedAt"`
	CompletedAt pgtype.Timestamp `json:"completedAt"`
	FailedAt    pgtype.Timestamp `json:"failedAt"`
	Error       *string          `json:"error"`
	Filename    string           `json:"filename"`
	Data        []byte           `json:"data"`
}