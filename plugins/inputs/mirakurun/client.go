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
	Process       *MirakurunStatusProcess       `json:"process"`
	EPG           *MirakurunStatusEPG           `json:"epg"`
	RPCCount      *int                          `json:"rpcCount"`
	StreamCount   *MirakurunStatusStreamCount   `json:"streamCount"`
	ErrorCount    *MirakurunStatusErrorCount    `json:"errorCount"`
	TimerAccuracy *MirakurunStatusTimerAccuracy `json:"timerAccuracy"`
}

type MirakurunStatusProcess struct {
	MemoryUsage *MirakurunStatusMemoryUsage `json:"memoryUsage"`
}

type MirakurunStatusMemoryUsage struct {
	RSS       int `json:"rss"`
	HeapTotal int `json:"heapTotal"`
	HeapUsed  int `json:"heapUsed"`
	// external と arrayBuffers は Node.js 実装の Mirakurun 固有であり、Mahiron などの互換実装は返さない
	External     *int `json:"external"`
	ArrayBuffers *int `json:"arrayBuffers"`
}

type MirakurunStatusEPG struct {
	StoredEvents int `json:"storedEvents"`
}

type MirakurunStatusStreamCount struct {
	TunerDevice int `json:"tunerDevice"`
	TSFilter    int `json:"tsFilter"`
	Decoder     int `json:"decoder"`
}

type MirakurunStatusErrorCount struct {
	UncaughtException  int `json:"uncaughtException"`
	UnhandledRejection int `json:"unhandledRejection"`
	BufferOverflow     int `json:"bufferOverflow"`
	TunerDeviceRespawn int `json:"tunerDeviceRespawn"`
	DecoderRespawn     int `json:"decoderRespawn"`
}

type MirakurunStatusTimerAccuracy struct {
	Last float64                           `json:"last"`
	M1   *MirakurunStatusTimerAccuracyStat `json:"m1"`
	M5   *MirakurunStatusTimerAccuracyStat `json:"m5"`
	M15  *MirakurunStatusTimerAccuracyStat `json:"m15"`
}

type MirakurunStatusTimerAccuracyStat struct {
	Avg float64 `json:"avg"`
	Min float64 `json:"min"`
	Max float64 `json:"max"`
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

	defer func() { _ = response.Body.Close() }()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return err
	}

	return nil
}
