package routes

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestElevationWithoutAuthorization(t *testing.T) {
	r := setupTestRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/v1/elevation", strings.NewReader(`{"points": [[1, 2], [3, 4]]}`))
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var resp struct {
		Data struct {
			Elevations []int32 `json:"elevations"`
		}
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, 2, len(resp.Data.Elevations))
}
