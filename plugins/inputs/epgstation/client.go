package epgstation

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/goccy/go-json"
)

type EPGStationClient struct {
	baseURL string
	client  *http.Client
}

func NewEPGStationClient(baseURL string) *EPGStationClient {
	return &EPGStationClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type EPGStationError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Errors  string `json:"errors"`
}

type EPGStationStreams struct {
	Items []*struct {
		Type string `json:"type"`
	} `json:"items"`
	EPGStationError
}

func (c *EPGStationClient) GetStreams(ctx context.Context) (*EPGStationStreams, error) {
	var result EPGStationStreams
	if err := c.get(ctx, "/api/streams?isHalfWidth=false", &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get streams: %d: %s, %s", result.Code, result.Message, result.Errors)
	}

	return &result, nil
}

type EPGStationReserveCounts struct {
	Normal    int `json:"normal"`
	Conflicts int `json:"conflicts"`
	Skips     int `json:"skips"`
	Overlaps  int `json:"overlaps"`
	EPGStationError
}

func (c *EPGStationClient) GetReserveCounts(ctx context.Context) (*EPGStationReserveCounts, error) {
	var result EPGStationReserveCounts
	if err := c.get(ctx, "/api/reserves/cnts", &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get reserve counts: %d: %s, %s", result.Code, result.Message, result.Errors)
	}

	return &result, nil
}

type EPGStationRecording struct {
	Total int `json:"total"`
	EPGStationError
}

func (c *EPGStationClient) GetRecording(ctx context.Context) (*EPGStationRecording, error) {
	var result EPGStationRecording
	if err := c.get(ctx, "/api/recording?isHalfWidth=false", &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get recording: %d: %s, %s", result.Code, result.Message, result.Errors)
	}

	return &result, nil
}

type EPGStationEncode struct {
	RunningItems []*struct{} `json:"runningItems"`
	WaitItems    []*struct{} `json:"waitItems"`
	EPGStationError
}

func (c *EPGStationClient) GetEncode(ctx context.Context) (*EPGStationEncode, error) {
	var result EPGStationEncode
	if err := c.get(ctx, "/api/encode?isHalfWidth=false", &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get encode: %d: %s, %s", result.Code, result.Message, result.Errors)
	}

	return &result, nil
}

type EPGStationStorages struct {
	Items []*struct {
		Name      string `json:"name"`
		Available int    `json:"available"`
		Used      int    `json:"used"`
		Total     int    `json:"total"`
	} `json:"items"`
	EPGStationError
}

func (c *EPGStationClient) GetStorages(ctx context.Context) (*EPGStationStorages, error) {
	var result EPGStationStorages
	if err := c.get(ctx, "/api/storages", &result); err != nil {
		return nil, err
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("failed to get storages: %d: %s, %s", result.Code, result.Message, result.Errors)
	}

	return &result, nil
}

func (c *EPGStationClient) get(ctx context.Context, path string, result any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", "telegraf-input-epgstation (+https://github.com/SlashNephy/telegraf-plugins)")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}

	return nil
}
