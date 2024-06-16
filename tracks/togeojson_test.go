package tracks

import (
	"context"
	"fmt"
	"github.com/go-jose/go-jose/v4/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestToGeoJSONService(t *testing.T) {
	t.Skip("skipping test")

	s := NewToGeoJSONService("http://togeojson.pt.svc.cluster.local")
	geojson, err := s.Convert(context.Background(), "file.gpx", sampleGPX())
	require.NoError(t, err)

	fmt.Println(string(geojson))

	var parsed struct {
		Type     string            `json:"type"`
		Features []json.RawMessage `json:"features"`
	}
	err = json.Unmarshal(geojson, &parsed)
	require.NoError(t, err)

	require.Equal(t, "FeatureCollection", parsed.Type)
	require.True(t, len(parsed.Features) > 0)
}
