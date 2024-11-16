package switchbot

import (
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
	"github.com/nasa9084/go-switchbot/v3"
	"golang.org/x/sync/errgroup"
)

//go:embed sample.conf
var sampleConfig string

type Plugin struct {
	client *switchbot.Client
	Log    telegraf.Logger `toml:"-"`

	SwitchBotOpenToken string `toml:"-" env:"SWITCHBOT_OPEN_TOKEN"`
	SwitchBotSecretKey string `toml:"-" env:"SWITCHBOT_SECRET_KEY"`
}

func init() {
	inputs.Add("switchbot", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	if err := env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	if p.SwitchBotOpenToken == "" || p.SwitchBotSecretKey == "" {
		return errors.New("open token and secret key are required")
	}

	p.client = switchbot.New(p.SwitchBotOpenToken, p.SwitchBotSecretKey)
	return nil
}

func (p *Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Gather(accumulator telegraf.Accumulator) error {
	ctx := context.Background()

	devices, err := p.queryDevices(ctx)
	if err != nil {
		return fmt.Errorf("failed to query devices: %w", err)
	}

	eg, egctx := errgroup.WithContext(ctx)
	for _, device := range devices {
		eg.Go(func() error {
			metrics := SupportedMetrics[device.Type]
			if len(metrics) == 0 {
				return nil
			}

			status, err := p.client.Device().Status(egctx, device.ID)
			if err != nil {
				return fmt.Errorf("failed to get status for %s: %w", device.ID, err)
			}

			fields := map[string]any{}
			for _, m := range metrics {
				fields[m.Key] = m.Value(&status)
			}

			accumulator.AddFields("switchbot", fields, map[string]string{
				"device_id":   device.ID,
				"device_name": device.Name,
				"device_type": string(device.Type),
				"hub_id":      device.Hub,
			})
			return nil
		})
	}

	if err = eg.Wait(); err != nil {
		return fmt.Errorf("failed to gather metrics: %w", err)
	}

	return nil
}

func (p *Plugin) queryDevices(ctx context.Context) ([]switchbot.Device, error) {
	// NOTE: InfraredDevice not supported
	devices, _, err := p.client.Device().List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list devices: %w", err)
	}

	return devices, nil
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
