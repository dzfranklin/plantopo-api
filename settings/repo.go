package settings

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
	q  *db.Queries
}

func NewRepo(pool *pgxpool.Pool) *Repo {
	return &Repo{db: pool, q: db.New(pool)}
}

func (r *Repo) GetUnitSettings(ctx context.Context, userID string) (json.RawMessage, error) {
	value, err := r.q.GetUnitSettings(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return value, nil
}

func (r *Repo) SetUnitSettings(ctx context.Context, userID string, value json.RawMessage) error {
	return r.q.SetUnitSettings(ctx, db.SetUnitSettingsParams{
		UserID: userID,
		Value:  value,
	})
}
