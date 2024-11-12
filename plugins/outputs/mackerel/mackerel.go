package mackerel

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
	"github.com/mackerelio/mackerel-client-go"
)

var (
	//go:embed sample.conf
	sampleConfig                  string
	notAcceptableMetricNamePatten = regexp.MustCompile("[^a-zA-Z0-9._-]+")
)

type Plugin struct {
	APIKey       string          `toml:"api_key" env:"MACKEREL_API_KEY"`
	HostID       string          `toml:"host_id" env:"MACKEREL_HOST_ID"`
	ServiceName  string          `toml:"service_name" env:"MACKEREL_SERVICE_NAME"`
	MetricPrefix string          `toml:"metric_prefix" env:"MACKEREL_METRIC_PREFIX"`
	Log          telegraf.Logger `toml:"-"`

	client *mackerel.Client
}

func (*Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Connect() error {
	if err := env.Parse(p); err != nil {
		return err
	}

	if p.APIKey != "" {
		p.client = mackerel.NewClient(p.APIKey)
	} else {
		p.Log.Warn("No API key configured, Mackerel output plugin never sends metrics to Mackerel")
	}

	if p.HostID == "" && p.ServiceName == "" {
		p.Log.Warn("Neither No host ID nor service name configured, Mackerel output plugin never sends metrics")
	}
	if p.HostID != "" && p.ServiceName != "" {
		p.Log.Warn("Both host ID and service name configured, Mackerel output plugin always sends metrics as Host Metrics")
	}

	return nil
}

func (p *Plugin) Close() error {
	return nil
}

func (p *Plugin) Write(metrics []telegraf.Metric) error {
	if p.client == nil {
		return nil
	}

	var payloads []*mackerel.MetricValue
	for _, metric := range metrics {
		for _, field := range metric.FieldList() {
			payloads = append(payloads, &mackerel.MetricValue{
				Name:  p.buildMetricName(metric, field),
				Time:  metric.Time().Unix(),
				Value: field.Value,
			})
		}
	}

	if p.HostID != "" {
		return p.client.PostHostMetricValuesByHostID(p.HostID, payloads)
	}

	if p.ServiceName != "" {
		return p.client.PostServiceMetricValues(p.ServiceName, payloads)
	}

	return nil
}

func (p *Plugin) buildMetricName(metric telegraf.Metric, field *telegraf.Field) string {
	prefix := p.MetricPrefix
	if prefix == "" {
		prefix = "telegraf"
	}

	keys := []string{"custom", prefix, metric.Name()}

	var tagKeys []string
	for _, tag := range metric.TagList() {
		tagKeys = append(tagKeys, tag.Key+"-"+tag.Value)
	}
	if len(tagKeys) > 0 {
		keys = append(keys, strings.Join(tagKeys, "_"))
	}

	keys = append(keys, field.Key)

	name := strings.Join(keys, ".")
	return notAcceptableMetricNamePatten.ReplaceAllString(name, "_")
}

func init() {
	outputs.Add("mackerel", func() telegraf.Output {
		return &Plugin{}
	})
}

var _ telegraf.Output = &Plugin{}
