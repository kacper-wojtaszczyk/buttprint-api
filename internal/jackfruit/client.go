package jackfruit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kacper-wojtaszczyk/buttprint-api/internal/domain"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *slog.Logger
}
type environmentalResponse struct {
	Variables []variableResponse `json:"variables"`
}
type variableResponse struct {
	Name         string           `json:"name"`
	Value        float64          `json:"value"`
	Unit         string           `json:"unit"`
	RefTimestamp time.Time        `json:"ref_timestamp"`
	ActualLat    float64          `json:"actual_lat"`
	ActualLon    float64          `json:"actual_lon"`
	Lineage      *lineageResponse `json:"lineage,omitempty"`
}

type lineageResponse struct {
	Source    string    `json:"source"`
	Dataset   string    `json:"dataset"`
	RawFileID uuid.UUID `json:"raw_file_id"`
}

var errNotFound = errors.New("environmental data not found")

func NewClient(httpClient *http.Client, baseURL string, logger *slog.Logger) *Client {
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     logger,
	}
}

func (c *Client) GetEnvironmentalData(
	ctx context.Context,
	lat, lon float64,
	timestamp time.Time,
	variables []string,
) ([]domain.VariableData, error) {
	queryParams := url.Values{
		"lat":       {strconv.FormatFloat(lat, 'f', -1, 64)},
		"lon":       {strconv.FormatFloat(lon, 'f', -1, 64)},
		"timestamp": {timestamp.Format(time.RFC3339)},
		"variables": {strings.Join(variables, ",")},
	}
	u := c.baseURL + "/v1/environmental?" + queryParams.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct environmental req: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute environmental req: %w", err)
	}
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			c.logger.Error("failed to close response body", "error", err)
		}
	}(resp.Body)

	envResp, err := decodeResponse(resp)
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil, domain.ErrNoData{Lat: lat, Lon: lon, Timestamp: timestamp}
		}

		return nil, domain.ErrUpstream{Service: "jackfruit", Cause: err}
	}

	return toDomainVariables(envResp.Variables), nil
}

func decodeResponse(resp *http.Response) (*environmentalResponse, error) {
	switch resp.StatusCode {
	case http.StatusOK:
		var parsed environmentalResponse
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON response: %w", err)
		}
		return &parsed, nil

	case http.StatusNotFound:
		return nil, errNotFound

	default:
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, body)
	}
}

func toDomainVariables(variables []variableResponse) []domain.VariableData {
	result := make([]domain.VariableData, len(variables))
	for i, variable := range variables {
		var lineage *domain.Lineage
		if variable.Lineage != nil {
			lineage = &domain.Lineage{
				Source:    variable.Lineage.Source,
				Dataset:   variable.Lineage.Dataset,
				RawFileID: variable.Lineage.RawFileID,
			}
		}
		result[i] = domain.VariableData{
			Name:         variable.Name,
			Value:        variable.Value,
			Unit:         variable.Unit,
			RefTimestamp: variable.RefTimestamp,
			ActualLat:    variable.ActualLat,
			ActualLon:    variable.ActualLon,
			Lineage:      lineage,
		}
	}

	return result
}
