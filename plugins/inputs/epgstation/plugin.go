package epgstation

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"golang.org/x/sync/errgroup"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "epgstation"

type Plugin struct {
	client *EPGStationClient

	EPGStationBaseURL string `toml:"-" env:"EPGSTATION_BASE_URL" envDefault:"http://localhost:8888"`
}

func init() {
	inputs.Add("epgstation", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	if err := env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	p.client = NewEPGStationClient(p.EPGStationBaseURL)
	return nil
}

func (p *Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Gather(accumulator telegraf.Accumulator) error {
	var eg errgroup.Group
	ctx := context.Background()

	getherFuncs := []func(context.Context, telegraf.Accumulator) error{
		p.gatherStreamMetrics,
		p.gatherReserveCountsMetrics,
		p.gatherRecordingMetrics,
		p.gatherEncodeMetrics,
		p.gatherStoragesMetrics,
	}
	for _, f := range getherFuncs {
		eg.Go(func() error {
			return f(ctx, accumulator)
		})
	}
	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to gather metrics: %w", err)
	}

	return nil
}

func (p *Plugin) gatherStreamMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	streams, err := p.client.GetStreams(ctx)
	if err != nil {
		return fmt.Errorf("failed to get streams: %w", err)
	}

	var (
		liveStreams     float64
		liveHLS         float64
		recordedStreams float64
		recordedHLS     float64
	)
	for _, stream := range streams.Items {
		switch stream.Type {
		case "LiveStream":
			liveStreams++
		case "LiveHLS":
			liveHLS++
		case "RecordedStream":
			recordedStreams++
		case "RecordedHLS":
			recordedHLS++
		}
	}

	accumulator.AddFields(measurement, map[string]any{
		"stream_live":         liveStreams,
		"stream_live_hls":     liveHLS,
		"stream_recorded":     recordedStreams,
		"stream_recorded_hls": recordedHLS,
	}, nil)
	return nil
}

func (p *Plugin) gatherReserveCountsMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	counts, err := p.client.GetReserveCounts(ctx)
	if err != nil {
		return fmt.Errorf("failed to get reserve counts: %w", err)
	}

	accumulator.AddFields(measurement, map[string]any{
		"reserve_normal":    counts.Normal,
		"reserve_conflicts": counts.Conflicts,
		"reserve_skips":     counts.Skips,
		"reserve_overlaps":  counts.Overlaps,
	}, nil)
	return nil
}

func (p *Plugin) gatherRecordingMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	recording, err := p.client.GetRecording(ctx)
	if err != nil {
		return fmt.Errorf("failed to get recording: %w", err)
	}

	accumulator.AddFields(measurement, map[string]any{
		"recording": recording.Total,
	}, nil)
	return nil
}

func (p *Plugin) gatherEncodeMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	encode, err := p.client.GetEncode(ctx)
	if err != nil {
		return fmt.Errorf("failed to get encodes: %w", err)
	}

	accumulator.AddFields(measurement, map[string]any{
		"encode_running": len(encode.RunningItems),
		"encode_waiting": len(encode.WaitItems),
	}, nil)
	return nil
}

func (p *Plugin) gatherStoragesMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	storages, err := p.client.GetStorages(ctx)
	if err != nil {
		return fmt.Errorf("failed to get storages: %w", err)
	}

	for _, storage := range storages.Items {
		accumulator.AddFields(measurement, map[string]any{
			"storage_total":     storage.Total,
			"storage_used":      storage.Used,
			"storage_available": storage.Available,
		}, map[string]string{
			"storage_name": storage.Name,
		})
	}
	return nil
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
