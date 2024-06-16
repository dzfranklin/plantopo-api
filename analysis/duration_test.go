package analysis

import (
	"github.com/paulmach/orb/geojson"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func sampleFeature() geojson.Feature {
	b := []byte(`{"type":"Feature","properties":{"_gpxType":"trk","name":"6/12/2024","time":"2024-06-12T09:03:59Z","coordinateProperties":{"times":["2024-06-12T09:03:59Z","2024-06-12T09:04:00Z","2024-06-12T09:04:06Z"]}},"geometry":{"type":"LineString","coordinates":[[-4.00387147,56.70437094,207],[-4.00383705,56.70436426,208],[-4.00371947,56.70432306,209]]}}`)
	f, err := geojson.UnmarshalFeature(b)
	if err != nil {
		panic(err)
	}
	return *f
}

func TestTrackDuration(t *testing.T) {
	f := sampleFeature()
	got, ok := TrackDuration(f)
	require.True(t, ok)
	assert.Equal(t, 7, got)
}
