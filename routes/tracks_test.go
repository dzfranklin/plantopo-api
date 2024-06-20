package routes

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/dzfranklin/plantopo-api/tracks"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/require"
	"mime/multipart"
	"net/http/httptest"
	"testing"
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

type AuthenticatorMock struct{}

func (a *AuthenticatorMock) Verify(_ string) (string, error) {
	return "user_test", nil
}

type TracksRepoMock struct{}

func (r *TracksRepoMock) ListMyTracksOrderByTime(_ context.Context, _ string) ([]tracks.Track, error) {
	return []tracks.Track{
		{ID: "track_42"},
	}, nil
}

func (r *TracksRepoMock) Import(_ context.Context, _ string, _ string, _ []byte) (string, error) {
	return "trackimport_42", nil
}

func (r *TracksRepoMock) Get(_ context.Context, id string) (tracks.Track, error) {
	return tracks.Track{
		ID: id,
	}, nil
}

func (r *TracksRepoMock) IsOwner(_ context.Context, _ string, _ string) (bool, error) {
	return true, nil
}

func (r *TracksRepoMock) Delete(_ context.Context, _ string) error {
	return nil
}

func (r *TracksRepoMock) ListMyPendingOrRecentImports(_ context.Context, _ string) ([]tracks.Import, error) {
	return []tracks.Import{}, nil
}

type ElevationMock struct{}

func (e *ElevationMock) QueryElevations(_ context.Context, points orb.LineString) ([]int32, error) {
	out := make([]int32, len(points))
	for i := range out {
		out[i] = int32(42)
	}
	return out, nil
}

func setupTestRouter() *gin.Engine {
	return Router(
		&AuthenticatorMock{},
		&TracksRepoMock{},
		&ElevationMock{},
	)
}

func TestListTracksRequiresAuth(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tracks/my?orderBy=time", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestListTracksFailsWithoutOrderBy(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tracks/my", nil)
	req.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(w, req)
	assert.Equal(t, 400, w.Code)
}

func TestListTracks(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/tracks/my?orderBy=time", nil)
	req.Header.Set("Authorization", "Bearer token")
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data []tracks.Track `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 1, len(resp.Data))
}

func TestImportTrackRequiresAuth(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/tracks/import", nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, 401, w.Code)
}

func TestImportTrack(t *testing.T) {
	var b bytes.Buffer
	reqW := multipart.NewWriter(&b)

	fw1, err := reqW.CreateFormFile("files", "Track 1.gpx")
	require.NoError(t, err)
	_, err = fw1.Write(sampleGPX())
	require.NoError(t, err)

	fw2, err := reqW.CreateFormFile("files", "Track 2.gpx")
	require.NoError(t, err)
	_, err = fw2.Write(sampleGPX())
	require.NoError(t, err)

	err = reqW.Close()
	require.NoError(t, err)

	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/tracks/import", &b)
	req.Header.Set("Authorization", "Bearer token")
	req.Header.Set("Content-Type", reqW.FormDataContentType())
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Imports []successfulTrackImportRequest `json:"imports"`
		}
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, 2, len(resp.Data.Imports))
}
