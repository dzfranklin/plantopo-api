package tracks

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dzfranklin/plantopo-api/analysis"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/paulmach/orb/geojson"
	"github.com/riverqueue/river"
	"log/slog"
	"strings"
	"time"
)

// TODO: This isn't properly responding to soft stops

type ToGeoJSON interface {
	Convert(ctx context.Context, filename string, data []byte) (json.RawMessage, error)
}

type Analyzer interface {
	HydrateTrack(ctx context.Context, input geojson.Feature) (geojson.Feature, error)
}

type ImportWorkerArgs struct {
	Id int64
}

func (ImportWorkerArgs) Kind() string { return "tracks_import" }

func (ImportWorkerArgs) InsertOps() river.InsertOpts {
	return river.InsertOpts{
		UniqueOpts: river.UniqueOpts{
			ByArgs: true,
		},
	}
}

type ImportWorker struct {
	db        *pgxpool.Pool
	toGeoJSON ToGeoJSON
	analyzer  Analyzer
	river.WorkerDefaults[ImportWorkerArgs]
}

func AddImportWorker(workers *river.Workers, db *pgxpool.Pool, toGeoJSON ToGeoJSON, analyzer Analyzer) {
	river.AddWorker[ImportWorkerArgs](workers, &ImportWorker{db: db, toGeoJSON: toGeoJSON, analyzer: analyzer})
}

func (w *ImportWorker) Work(ctx context.Context, job *river.Job[ImportWorkerArgs]) error {
	importId := job.Args.Id
	uploadTime := job.CreatedAt
	l := slog.With("job", job.ID, "import", importId, "created_at", job.CreatedAt)

	q := db.New(w.db)
	data, err := q.GetTrackImport(ctx, importId)
	if err != nil {
		l.Error("get track import", "error", err)
		return err
	}

	if data.CompletedAt.Valid || data.FailedAt.Valid {
		l.Info("already done")
		return nil
	}

	rawGeojson, err := w.toGeoJSON.Convert(ctx, data.Filename, data.Data)
	if err != nil {
		invalidConversionErr := InvalidConversionInputError{}
		if errors.As(err, &invalidConversionErr) {
			l.Info("invalid conversion input", "error", invalidConversionErr.Message)
			return q.MarkTrackImportFailed(ctx, db.MarkTrackImportFailedParams{
				ID:    importId,
				Error: &invalidConversionErr.Message,
			})
		}
		l.Error("convert import to geojson", "error", err)
		return err
	}

	trackFeatures, err := geojson.UnmarshalFeatureCollection(rawGeojson)
	if err != nil {
		err = fmt.Errorf("invalid geojson: %w", err)
		l.Error("unmarshal geojson", "error", err)
		return err
	}

	var tracks []db.InsertImportedTrackParams
	for i, rawFeature := range trackFeatures.Features {
		if rawFeature.Geometry.GeoJSONType() != "LineString" {
			l.Info("skipping non-line feature", "i", i, "type", rawFeature.Geometry.GeoJSONType())
			continue
		}

		feature, err := w.analyzer.HydrateTrack(ctx, *rawFeature)
		if err != nil {
			l.Error("hydrate track", "i", i, "error", err)
			return err
		}

		name := importName(data.Filename, &feature)
		trackTime := importTrackTime(&feature, uploadTime)
		track := db.InsertImportedTrackParams{
			OwnerID:    &data.OwnerID,
			Name:       &name,
			UploadTime: pgtype.Timestamp{Time: uploadTime, Valid: true},
			Time:       pgtype.Timestamp{Time: trackTime, Valid: true},
			Geojson:    feature,
			ImportID:   &importId,
		}
		tracks = append(tracks, track)
	}

	completeTx, err := w.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer completeTx.Rollback(ctx)

	err = q.MarkTrackImportCompleted(ctx, importId)
	if err != nil {
		return err
	}

	for _, track := range tracks {
		_, err = q.InsertImportedTrack(ctx, track)
		if err != nil {
			return err
		}
	}

	return completeTx.Commit(ctx)
}

func importName(filename string, track *geojson.Feature) string {
	if prop, ok := track.Properties["name"].(string); ok {
		return prop
	}

	if !strings.Contains(filename, ".") {
		return filename
	}
	return filename[:strings.LastIndex(filename, ".")]
}

func importTrackTime(track *geojson.Feature, fallback time.Time) time.Time {
	t, ok := analysis.ParseSloppyRecentTime(track.Properties["time"])
	if ok {
		return t
	} else {
		return fallback
	}
}
