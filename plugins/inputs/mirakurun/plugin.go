package mirakurun

import (
	"context"
	_ "embed"
	"fmt"
	"runtime/debug"

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
		eg.Go(func() (err error) {
			// 未知のレスポンス形状で panic しても execd プロセス全体を落とさない
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic while gathering metrics: %v\n%s", r, debug.Stack())
				}
			}()

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

	// Mahiron などの Mirakurun 互換実装は Node.js 実装固有の指標を返さない
	// 欠けている指標を 0 として記録すると実際に 0 だった場合と区別できないため、フィールドごと出力しない
	fields := make(map[string]any)

	if status.Process != nil && status.Process.MemoryUsage != nil {
		memory := status.Process.MemoryUsage
		fields["memory_rss"] = memory.RSS
		fields["memory_heap_total"] = memory.HeapTotal
		fields["memory_heap_used"] = memory.HeapUsed

		if memory.External != nil {
			fields["memory_external"] = *memory.External
		}
		if memory.ArrayBuffers != nil {
			fields["memory_array_buffers"] = *memory.ArrayBuffers
		}
	}

	if status.EPG != nil {
		fields["epg_stored_events"] = status.EPG.StoredEvents
	}

	if status.RPCCount != nil {
		fields["rpc_count"] = *status.RPCCount
	}

	if status.StreamCount != nil {
		stream := status.StreamCount
		fields["stream_total"] = stream.TunerDevice + stream.TSFilter + stream.Decoder
		fields["stream_tuner_device"] = stream.TunerDevice
		fields["stream_ts_filter"] = stream.TSFilter
		fields["stream_decoder"] = stream.Decoder
	}

	if status.ErrorCount != nil {
		errorCount := status.ErrorCount
		fields["error_total"] = errorCount.UncaughtException + errorCount.UnhandledRejection + errorCount.BufferOverflow + errorCount.TunerDeviceRespawn + errorCount.DecoderRespawn
		fields["error_uncaught_exception"] = errorCount.UncaughtException
		fields["error_unhandled_rejection"] = errorCount.UnhandledRejection
		fields["error_buffer_overflow"] = errorCount.BufferOverflow
		fields["error_tuner_device_respawn"] = errorCount.TunerDeviceRespawn
		fields["error_decoder_respawn"] = errorCount.DecoderRespawn
	}

	if status.TimerAccuracy != nil {
		fields["timer_accuracy"] = status.TimerAccuracy.Last

		stats := []struct {
			name string
			stat *MirakurunStatusTimerAccuracyStat
		}{
			{"m1", status.TimerAccuracy.M1},
			{"m5", status.TimerAccuracy.M5},
			{"m15", status.TimerAccuracy.M15},
		}
		for _, entry := range stats {
			if entry.stat == nil {
				continue
			}

			fields["timer_accuracy_"+entry.name+"_avg"] = entry.stat.Avg
			fields["timer_accuracy_"+entry.name+"_min"] = entry.stat.Min
			fields["timer_accuracy_"+entry.name+"_max"] = entry.stat.Max
		}
	}

	if len(fields) == 0 {
		return nil
	}

	accumulator.AddFields(measurement, fields, nil)
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
