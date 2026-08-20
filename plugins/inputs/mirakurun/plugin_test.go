package mirakurun

import (
	"maps"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/stretchr/testify/require"
)

// testAccumulator は AddFields で記録されたメトリクスを検証するための最小実装
// telegraf/testutil は testcontainers に依存し依存関係が重いため、ここでは埋め込みで interface を満たす
type testAccumulator struct {
	telegraf.Accumulator

	mu      sync.Mutex
	metrics []testMetric
}

type testMetric struct {
	measurement string
	fields      map[string]any
}

func (a *testAccumulator) AddFields(measurement string, fields map[string]any, _ map[string]string, _ ...time.Time) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.metrics = append(a.metrics, testMetric{measurement: measurement, fields: fields})
}

var _ telegraf.Accumulator = new(testAccumulator)

// Mirakurun (Node.js 実装) が返すレスポンス
const mirakurunStatusResponse = `{
  "time": 1787131323141,
  "version": "3.9.0-rc.4",
  "process": {
    "arch": "x64",
    "platform": "linux",
    "pid": 1,
    "memoryUsage": { "rss": 190581048, "heapTotal": 176062464, "heapUsed": 22377608, "external": 3145728, "arrayBuffers": 1048576 }
  },
  "epg": { "storedEvents": 17614 },
  "rpcCount": 42,
  "streamCount": { "tunerDevice": 1, "tsFilter": 11, "decoder": 11 },
  "errorCount": { "uncaughtException": 1, "unhandledRejection": 2, "bufferOverflow": 3, "tunerDeviceRespawn": 4, "decoderRespawn": 5 },
  "timerAccuracy": {
    "last": 1234.5,
    "m1": { "avg": 1.0, "min": 0.5, "max": 1.5 },
    "m5": { "avg": 2.0, "min": 1.5, "max": 2.5 },
    "m15": { "avg": 3.0, "min": 2.5, "max": 3.5 }
  }
}`

// Mahiron 5 系が返すレスポンス (https://github.com/SlashNephy/telegraf-plugins/issues/169)
const mahironStatusResponse = `{
  "time": 1787131323141,
  "version": "5.0.15",
  "process": {
    "arch": "amd64",
    "platform": "linux",
    "pid": 1,
    "memoryUsage": { "rss": 190581048, "heapTotal": 176062464, "heapUsed": 22377608 }
  },
  "epg": {
    "gatheringNetworks": [32375, 4],
    "storedEvents": 17614,
    "staleServices": 65,
    "failedServices": 0,
    "lastUpdatedAt": 1787131297000
  },
  "streamCount": { "tunerDevice": 1, "tsFilter": 11, "decoder": 11 }
}`

const channelsResponse = `[
  { "type": "GR", "services": [{}, {}] },
  { "type": "BS", "services": [{}] },
  { "type": "CS", "services": [] },
  { "type": "SKY", "services": [{}] },
  { "type": "NW", "services": [{}, {}, {}] }
]`

const tunersResponse = `[
  { "isAvailable": true, "isRemote": false, "isFree": true, "isUsing": false, "isFault": false },
  { "isAvailable": true, "isRemote": true, "isFree": false, "isUsing": true, "isFault": false }
]`

func newTestPlugin(t *testing.T, statusResponse string) *Plugin {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(statusResponse))
	})
	mux.HandleFunc("/api/channels", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(channelsResponse))
	})
	mux.HandleFunc("/api/tuners", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(tunersResponse))
	})

	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	t.Setenv("MIRAKURUN_BASE_URL", server.URL)

	plugin := new(Plugin)
	require.NoError(t, plugin.Init())
	return plugin
}

func TestPluginGatherStatusMetrics(t *testing.T) {
	tests := []struct {
		name           string
		statusResponse string
		expected       map[string]any
	}{
		{
			name:           "mirakurun",
			statusResponse: mirakurunStatusResponse,
			expected: map[string]any{
				"memory_rss":                 190581048,
				"memory_heap_total":          176062464,
				"memory_heap_used":           22377608,
				"memory_external":            3145728,
				"memory_array_buffers":       1048576,
				"epg_stored_events":          17614,
				"rpc_count":                  42,
				"stream_total":               23,
				"stream_tuner_device":        1,
				"stream_ts_filter":           11,
				"stream_decoder":             11,
				"error_total":                15,
				"error_uncaught_exception":   1,
				"error_unhandled_rejection":  2,
				"error_buffer_overflow":      3,
				"error_tuner_device_respawn": 4,
				"error_decoder_respawn":      5,
				"timer_accuracy":             1234.5,
				"timer_accuracy_m1_avg":      1.0,
				"timer_accuracy_m1_min":      0.5,
				"timer_accuracy_m1_max":      1.5,
				"timer_accuracy_m5_avg":      2.0,
				"timer_accuracy_m5_min":      1.5,
				"timer_accuracy_m5_max":      2.5,
				"timer_accuracy_m15_avg":     3.0,
				"timer_accuracy_m15_min":     2.5,
				"timer_accuracy_m15_max":     3.5,
			},
		},
		{
			// errorCount / timerAccuracy / rpcCount / external / arrayBuffers は欠落するため出力しない
			name:           "mahiron",
			statusResponse: mahironStatusResponse,
			expected: map[string]any{
				"memory_rss":          190581048,
				"memory_heap_total":   176062464,
				"memory_heap_used":    22377608,
				"epg_stored_events":   17614,
				"stream_total":        23,
				"stream_tuner_device": 1,
				"stream_ts_filter":    11,
				"stream_decoder":      11,
			},
		},
		{
			name:           "empty",
			statusResponse: `{}`,
			expected:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := newTestPlugin(t, test.statusResponse)

			var accumulator testAccumulator
			require.NoError(t, plugin.gatherStatusMetrics(t.Context(), &accumulator))

			if test.expected == nil {
				require.Empty(t, accumulator.metrics)
				return
			}

			require.Len(t, accumulator.metrics, 1)
			require.Equal(t, measurement, accumulator.metrics[0].measurement)
			require.Equal(t, test.expected, accumulator.metrics[0].fields)
		})
	}
}

func TestPluginGather(t *testing.T) {
	tests := []struct {
		name           string
		statusResponse string
		expected       map[string]any
	}{
		{
			name:           "mirakurun",
			statusResponse: mirakurunStatusResponse,
		},
		{
			name:           "mahiron",
			statusResponse: mahironStatusResponse,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			plugin := newTestPlugin(t, test.statusResponse)

			var accumulator testAccumulator
			require.NoError(t, plugin.Gather(&accumulator))

			// status / channels / tuners の 3 つが揃って収集される
			require.Len(t, accumulator.metrics, 3)

			fields := make(map[string]any)
			for _, metric := range accumulator.metrics {
				require.Equal(t, measurement, metric.measurement)
				maps.Copy(fields, metric.fields)
			}

			require.Subset(t, fields, map[string]any{
				"memory_rss":   190581048,
				"stream_total": 23,
				"channels":     5,
				"channels_gr":  1,
				"channels_bs":  1,
				"channels_cs":  1,
				"channels_sky": 1,
				"services":     7,
				"services_gr":  2,
				"services_bs":  1,
				"services_cs":  0,
				"services_sky": 1,

				"tuner_available": 2,
				"tuner_remote":    1,
				"tuner_free":      1,
				"tuner_using":     1,
				"tuner_fault":     0,
			})
		})
	}
}
