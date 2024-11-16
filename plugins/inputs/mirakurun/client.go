package mirakurun

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type MirakurunClient struct {
	baseURL string
}

func NewMirakurunClient(baseURL string) *MirakurunClient {
	return &MirakurunClient{
		baseURL: baseURL,
	}
}

type MirakurunStatus struct {
	Process *struct {
		MemoryUsage *struct {
			RSS          int `json:"rss"`
			HeapTotal    int `json:"heapTotal"`
			HeapUsed     int `json:"heapUsed"`
			External     int `json:"external"`
			ArrayBuffers int `json:"arrayBuffers"`
		} `json:"memoryUsage"`
	} `json:"process"`
	EPG *struct {
		StoredEvents int `json:"storedEvents"`
	} `json:"epg"`
	RPCCount    int `json:"rpcCount"`
	StreamCount *struct {
		TunerDevice int `json:"tunerDevice"`
		TSFilter    int `json:"tsFilter"`
		Decoder     int `json:"decoder"`
	} `json:"streamCount"`
	ErrorCount *struct {
		UncaughtException  int `json:"uncaughtException"`
		UnhandledRejection int `json:"unhandledRejection"`
		BufferOverflow     int `json:"bufferOverflow"`
		TunerDeviceRespawn int `json:"tunerDeviceRespawn"`
		DecoderRespawn     int `json:"decoderRespawn"`
	} `json:"errorCount"`
	TimerAccuracy *struct {
		Last float64 `json:"last"`
		M1   *struct {
			Avg float64 `json:"avg"`
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"m1"`
		M5 *struct {
			Avg float64 `json:"avg"`
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"m5"`
		M15 *struct {
			Avg float64 `json:"avg"`
			Min float64 `json:"min"`
			Max float64 `json:"max"`
		} `json:"m15"`
	} `json:"timerAccuracy"`
}

func (c *MirakurunClient) GetStatus(ctx context.Context) (*MirakurunStatus, error) {
	var result MirakurunStatus
	if err := c.get(ctx, "/api/status", &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type MirakurunChannel struct {
	Type     string      `json:"type"`
	Services []*struct{} `json:"services"`
}

func (c *MirakurunClient) GetChannels(ctx context.Context) ([]*MirakurunChannel, error) {
	var results []*MirakurunChannel
	if err := c.get(ctx, "/api/channels", &results); err != nil {
		return nil, err
	}

	return results, nil
}

type MirakurunTuner struct {
	IsAvailable bool `json:"isAvailable"`
	IsRemote    bool `json:"isRemote"`
	IsFree      bool `json:"isFree"`
	IsUsing     bool `json:"isUsing"`
	IsFault     bool `json:"isFault"`
}

func (c *MirakurunClient) GetTuners(ctx context.Context) ([]*MirakurunTuner, error) {
	var results []*MirakurunTuner
	if err := c.get(ctx, "/api/tuners", &results); err != nil {
		return nil, err
	}

	return results, nil
}

func (c *MirakurunClient) get(ctx context.Context, path string, result any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}

	request.Header.Set("User-Agent", "telegraf-input-mirakurun (+https://github.com/SlashNephy/telegraf-plugins)")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
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
