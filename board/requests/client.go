package requests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"pips-solver/backend/board/types"
	"time"
)

// TODO move to a config
const CLIENT_TIMEOUT = 10

type NytClient struct {
	baseUrl    *url.URL
	httpClient *http.Client
	token      string
}

func NewClient(base string, token string, httpClient *http.Client) (*NytClient, error) {
	parsed, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("ERROR[NewClient] failed to initialize client: %w", err)
	}

	if httpClient == nil {
		httpClient = &http.Client{Timeout: CLIENT_TIMEOUT * time.Second}
	}

	return &NytClient{baseUrl: parsed, httpClient: httpClient, token: token}, nil
}

func (c *NytClient) GetPuzzles(ctx context.Context, date string) (*types.PuzzlePayload, error) {
	reqPayload := types.PuzzleReqPayload{AbsoluteDate: date, Game: "pips"}
	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, fmt.Errorf("ERROR[GetPuzzles] failed to create payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl.String(), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("ERROR[GetPuzzles] failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ERROR[GetPuzzles] failed to make request: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("ERROR[GetPuzzles] invalid response status: %w", err)
	}

	var payload []types.PuzzlePayload
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("ERROR[GetPuzzles] faild to parse response: %w", err)
	}

	return &payload[0], nil
}
