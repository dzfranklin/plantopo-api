package analysis

import (
	"context"
	"fmt"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/geojson"
	"math"
)

type Analyzer struct {
	elevation ElevationQuerier
}

type ElevationQuerier interface {
	QueryElevations(ctx context.Context, points orb.LineString) ([]int32, error)
}

func NewAnalyzer(elevation ElevationQuerier) *Analyzer {
	return &Analyzer{elevation: elevation}
}

func (a *Analyzer) HydrateTrack(ctx context.Context, f geojson.Feature) (geojson.Feature, error) {
	geom, ok := f.Geometry.(orb.LineString)
	if !ok {
		return geojson.Feature{}, fmt.Errorf("expected LineString, got %s", f.Geometry.GeoJSONType())
	}

	props := f.Properties
	coordProps := props.CoordinateProperties()

	length := roundPlaces(geo.LengthHaversine(geom), 6)
	props["lengthMeters"] = length

	durationSecs, ok := TrackDuration(f)
	if ok {
		props["durationSecs"] = durationSecs
	}

	elevations, err := a.elevation.QueryElevations(ctx, geom)
	if err != nil {
		return geojson.Feature{}, fmt.Errorf("query elevations: %w", err)
	}
	coordProps["elevationMeters"] = elevations

	return f, nil
}

func roundPlaces(f float64, places int) float64 {
	scale := float64(1)
	for i := 0; i < places; i++ {
		scale *= 10
	}
	return math.Round(f*scale) / scale
}
