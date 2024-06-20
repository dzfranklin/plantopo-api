package tracks

import (
	"context"
	"encoding/json"
	"github.com/dzfranklin/plantopo-api/db"
	"github.com/dzfranklin/plantopo-api/testsupport"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/rivertype"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func sampleGPX() []byte {
	return []byte(`
<?xml version="1.0"?>
<gpx xmlns="http://www.topografix.com/GPX/1/1" xmlns:gpxx="http://www.garmin.com/xmlschemas/GpxExtensions/v3" creator="CALTOPO" version="1.1">
	<metadata>
		<name><![CDATA[export]]></name>
	</metadata>
	<trk>
		<name>6/12/2024</name>
		<trkseg>
			<trkpt lat="56.70437094" lon="-4.00387147"><ele>207.0</ele><time>2024-06-12T09:03:59Z</time></trkpt>
			<trkpt lat="56.70436426" lon="-4.00383705"><ele>208.0</ele><time>2024-06-12T09:04:00Z</time></trkpt>
			<trkpt lat="56.70432306" lon="-4.00371947"><ele>209.0</ele><time>2024-06-12T09:04:06Z</time></trkpt>
		</trkseg>
	</trk>
</gpx>
`)
}

func sampleGeojson() json.RawMessage {
	return []byte(`{"type":"FeatureCollection","features":[{"type":"Feature","properties":{"_gpxType":"trk","name":"6/12/2024","time":"2024-06-12T09:03:59Z","coordinateProperties":{"times":["2024-06-12T09:03:59Z","2024-06-12T09:04:00Z","2024-06-12T09:04:06Z"]}},"geometry":{"type":"LineString","coordinates":[[-4.00387147,56.70437094,207],[-4.00383705,56.70436426,208],[-4.00371947,56.70432306,209]]}}]}`)
}

type MockToGeoJSON struct{}

func (MockToGeoJSON) Convert(_ context.Context, _ string, _ []byte) (json.RawMessage, error) {
	return sampleGeojson(), nil
}

type MockAnalyzer struct {
	called bool
}

func (a *MockAnalyzer) HydrateTrack(_ context.Context, f geojson.Feature) (geojson.Feature, error) {
	a.called = true
	return f, nil
}

func TestImportWorker(t *testing.T) {
	ctx := context.Background()
	pool := testsupport.NewDB(t)
	q := db.New(pool)
	analyzer := &MockAnalyzer{}
	w := &ImportWorker{db: pool, toGeoJSON: &MockToGeoJSON{}, analyzer: analyzer}

	owner := "user_1"
	_, err := q.InsertTrackImport(ctx, db.InsertTrackImportParams{
		OwnerID:  owner,
		Filename: "file.gpx",
		Data:     sampleGPX(),
		Hash:     []byte("sample_hash"),
	})
	require.NoError(t, err)

	err = w.Work(ctx, &river.Job[ImportWorkerArgs]{
		Args: ImportWorkerArgs{Id: 1},
		JobRow: &rivertype.JobRow{
			ID: 1,
		},
	})
	require.NoError(t, err)

	require.True(t, analyzer.called)

	gotTracks, err := q.ListTracksOrderByTime(ctx, &owner)
	require.NoError(t, err)
	require.Len(t, gotTracks, 1)

	got := gotTracks[0]
	require.Equal(t, owner, *got.OwnerID)
	require.Equal(t, "6/12/2024", *got.Name)
	require.Equal(t, "2024-06-12T09:03:59Z", got.Time.Time.Format(time.RFC3339))
	require.Equal(t, 3, len(got.Geojson.Geometry.(orb.LineString)))
}

func TestImportName(t *testing.T) {
	cases := []struct {
		name     string
		filename string
		track    string
		expected string
	}{
		{
			"prefers property",
			"filename.gpx",
			`{"type": "Feature", "properties":{"name":"From Property"}}`,
			"From Property",
		},
		{
			"ignores malformed property",
			"filename.gpx",
			`{"type": "Feature", "properties":{"name":42}}`,
			"filename",
		},
		{
			"falls back to filename stripped of ext",
			"filename.gpx",
			`{"type": "Feature"}`,
			"filename",
		},
		{
			"handles filename without ext",
			"filename",
			`{"type": "Feature"}`,
			"filename",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			track, err := geojson.UnmarshalFeature([]byte(c.track))
			require.NoError(t, err)
			require.Equal(t, c.expected, importName(c.filename, track))
		})
	}
}
