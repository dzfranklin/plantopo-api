package analysis

import (
	"context"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/require"
	"testing"
)

type MockElevationQuerier struct{}

func (m *MockElevationQuerier) QueryElevations(_ context.Context, points orb.LineString) ([]int32, error) {
	var out []int32
	for range points {
		out = append(out, 42)
	}
	return out, nil
}

func TestAnalyzer_HydrateTrack(t *testing.T) {
	subject := NewAnalyzer(&MockElevationQuerier{})

	input := *geojson.NewFeature(orb.LineString{{0, 0}, {1, 1}})

	got, err := subject.HydrateTrack(context.Background(), input)
	require.NoError(t, err)

	props := got.Properties
	require.Equal(t, 157425.537108, props["lengthMeters"])
	require.Equal(t, []int32{42, 42}, props.CoordinateProperties()["elevationMeters"])
}
