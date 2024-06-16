package tracks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type InvalidConversionInputError struct {
	Message string
}

func (e InvalidConversionInputError) Error() string {
	return e.Message
}

type ToGeoJSONService struct {
	endpoint string
	client   *http.Client
}

func NewToGeoJSONService(endpoint string) *ToGeoJSONService {
	if strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint[:len(endpoint)-1]
	}
	client := &http.Client{}
	return &ToGeoJSONService{endpoint: endpoint, client: client}
}

func (c *ToGeoJSONService) Convert(ctx context.Context, filename string, data []byte) (json.RawMessage, error) {
	if !strings.Contains(filename, ".") {
		return nil, InvalidConversionInputError{"missing file extension"}
	}
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	if ext != "gpx" {
		return nil, InvalidConversionInputError{"unsupported file extension: " + ext}
	}

	u, err := url.Parse(c.endpoint + "/gpx")
	if err != nil {
		return nil, err
	}
	req := &http.Request{
		Method: http.MethodPost,
		URL:    u,
		Body:   io.NopCloser(bytes.NewReader(data)),
	}
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			return nil, InvalidConversionInputError{"Invalid GPX file"}
		}
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var geojson json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&geojson); err != nil {
		return nil, err
	}
	return geojson, nil
}
