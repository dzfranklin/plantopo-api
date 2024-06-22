package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff/v4"
	"github.com/dzfranklin/plantopo-api/meta"
	"github.com/paulmach/orb"
	"io"
	"log/slog"
	"net/http"
)

// Consider using https://elevationapi.com for areas we don't support

type ElevationService struct {
	http     *http.Client
	endpoint string
}

func NewElevationService(endpoint string) *ElevationService {
	return &ElevationService{
		http:     &http.Client{},
		endpoint: endpoint,
	}
}

// QueryElevations queries the elevation service for the given points.
//
// The input points are a list of [longitude, latitude] pairs.
func (s *ElevationService) QueryElevations(ctx context.Context, points orb.LineString) ([]float64, error) {
	var elevations []float64
	err := backoff.Retry(func() error {
		var err error
		elevations, err = doElevationLookup(ctx, s.http, s.endpoint+"/elevation", points)
		if errors.Is(err, context.Canceled) {
			return backoff.Permanent(err)
		} else {
			slog.Warn("Error querying elevations", "error", err)
			return err
		}
	}, backoff.NewExponentialBackOff())
	return elevations, err
}

func doElevationLookup(ctx context.Context, client *http.Client, url string, line orb.LineString) ([]float64, error) {
	reqData := struct {
		Coordinates orb.LineString `json:"coordinates"`
	}{line}
	reqBody, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	meta.SetUserAgent(req)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("%s: %s", resp.Status, string(body))
	}
	var result struct {
		Elevations []float64 `json:"elevations"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Elevations, nil
}
