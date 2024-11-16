package mirakurun

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
)

//go:embed sample.conf
var sampleConfig string

const measurement = "mirakurun"

type Plugin struct {
	client *MirakurunClient

	MirakurunBaseURL string `toml:"-" env:"MIRAKURUN_BASE_URL" envDefault:"http://localhost:40772"`
}

func init() {
	inputs.Add("mirakurun", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	if err := env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	p.client = NewMirakurunClient(p.MirakurunBaseURL)
	return nil
}

func (p *Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Gather(accumulator telegraf.Accumulator) error {
	var eg errgroup.Group
	ctx := context.Background()

	getherFuncs := []func(context.Context, telegraf.Accumulator) error{
		p.gatherStatusMetrics,
		p.gatherChannelsMetrics,
		p.gatherTunersMetrics,
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

func (p *Plugin) gatherStatusMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	status, err := p.client.GetStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	accumulator.AddFields(measurement, map[string]any{
		"memory_rss":                 status.Process.MemoryUsage.RSS,
		"memory_heap_total":          status.Process.MemoryUsage.HeapTotal,
		"memory_heap_used":           status.Process.MemoryUsage.HeapUsed,
		"memory_external":            status.Process.MemoryUsage.External,
		"memory_array_buffers":       status.Process.MemoryUsage.ArrayBuffers,
		"epg_stored_events":          status.EPG.StoredEvents,
		"rpc_count":                  status.RPCCount,
		"stream_total":               status.StreamCount.TunerDevice + status.StreamCount.TSFilter + status.StreamCount.Decoder,
		"stream_tuner_device":        status.StreamCount.TunerDevice,
		"stream_ts_filter":           status.StreamCount.TSFilter,
		"stream_decoder":             status.StreamCount.Decoder,
		"error_total":                status.ErrorCount.UncaughtException + status.ErrorCount.UnhandledRejection + status.ErrorCount.BufferOverflow + status.ErrorCount.TunerDeviceRespawn + status.ErrorCount.DecoderRespawn,
		"error_uncaught_exception":   status.ErrorCount.UncaughtException,
		"error_unhandled_rejection":  status.ErrorCount.UnhandledRejection,
		"error_buffer_overflow":      status.ErrorCount.BufferOverflow,
		"error_tuner_device_respawn": status.ErrorCount.TunerDeviceRespawn,
		"error_decoder_respawn":      status.ErrorCount.DecoderRespawn,
		"timer_accuracy":             status.TimerAccuracy.Last,
		"timer_accuracy_m1_avg":      status.TimerAccuracy.M1.Avg,
		"timer_accuracy_m1_min":      status.TimerAccuracy.M1.Min,
		"timer_accuracy_m1_max":      status.TimerAccuracy.M1.Max,
		"timer_accuracy_m5_avg":      status.TimerAccuracy.M5.Avg,
		"timer_accuracy_m5_min":      status.TimerAccuracy.M5.Min,
		"timer_accuracy_m5_max":      status.TimerAccuracy.M5.Max,
		"timer_accuracy_m15_avg":     status.TimerAccuracy.M15.Avg,
		"timer_accuracy_m15_min":     status.TimerAccuracy.M15.Min,
		"timer_accuracy_m15_max":     status.TimerAccuracy.M15.Max,
	}, nil)
	return nil
}

func (p *Plugin) gatherChannelsMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	channels, err := p.client.GetChannels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get channels: %w", err)
	}

	channelsByType := lo.GroupBy(channels, func(c *MirakurunChannel) string {
		return c.Type
	})

	accumulator.AddFields(measurement, map[string]any{
		"channels":     len(channels),
		"channels_gr":  len(channelsByType["GR"]),
		"channels_bs":  len(channelsByType["BS"]),
		"channels_cs":  len(channelsByType["CS"]),
		"channels_sky": len(channelsByType["SKY"]),
		"services":     lo.SumBy(channels, func(c *MirakurunChannel) int { return len(c.Services) }),
		"services_gr":  lo.SumBy(channelsByType["GR"], func(c *MirakurunChannel) int { return len(c.Services) }),
		"services_bs":  lo.SumBy(channelsByType["BS"], func(c *MirakurunChannel) int { return len(c.Services) }),
		"services_cs":  lo.SumBy(channelsByType["CS"], func(c *MirakurunChannel) int { return len(c.Services) }),
		"services_sky": lo.SumBy(channelsByType["SKY"], func(c *MirakurunChannel) int { return len(c.Services) }),
	}, nil)
	return nil
}

func (p *Plugin) gatherTunersMetrics(ctx context.Context, accumulator telegraf.Accumulator) error {
	tuners, err := p.client.GetTuners(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tuners: %w", err)
	}

	accumulator.AddFields(measurement, map[string]any{
		"tuner_available": lo.CountBy(tuners, func(t *MirakurunTuner) bool { return t.IsAvailable }),
		"tuner_remote":    lo.CountBy(tuners, func(t *MirakurunTuner) bool { return t.IsRemote }),
		"tuner_free":      lo.CountBy(tuners, func(t *MirakurunTuner) bool { return t.IsFree }),
		"tuner_using":     lo.CountBy(tuners, func(t *MirakurunTuner) bool { return t.IsUsing }),
		"tuner_fault":     lo.CountBy(tuners, func(t *MirakurunTuner) bool { return t.IsFault }),
	}, nil)
	return nil
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
