package rakutensecurities

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Plugin struct {
	client *RakutenSecuritiesClient

	Username string `toml:"-" env:"RAKUTEN_SECURITIES_USERNAME"`
	Password string `toml:"-" env:"RAKUTEN_SECURITIES_PASSWORD"`
}

func init() {
	inputs.Add("rakuten_securities", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	var err error
	if err = env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	if p.Username == "" || p.Password == "" {
		return errors.New("username and password are required")
	}

	p.client, err = NewRakutenSecuritiesClient()
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	return nil
}

func (p *Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Gather(accumulator telegraf.Accumulator) error {
	ctx := context.Background()

	return p.gather(ctx, accumulator, 0)
}

func (p *Plugin) gather(ctx context.Context, accumulator telegraf.Accumulator, loop int) error {
	if loop > 3 {
		return errors.New("max retry limit exceeded")
	}

	metrics, err := p.client.GetMetrics(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to get metrics", slog.String("error", err.Error()))

		if errors.Is(err, ErrUnauthorized) {
			if err = p.client.Login(ctx, p.Username, p.Password); err != nil {
				slog.ErrorContext(ctx, "failed to login", slog.String("error", err.Error()))
				return err
			}
			return p.gather(ctx, accumulator, loop+1)
		}
		return fmt.Errorf("failed to get metrics: %w", err)
	}

	for _, summary := range metrics.Summaries {
		p.gatherAssetSummary(accumulator, summary)
	}

	for _, asset := range metrics.Assets {
		p.gatherAsset(accumulator, asset)
	}

	for _, rate := range metrics.ExchangeRates {
		p.gatherExchangeRate(accumulator, rate)
	}

	return nil
}

func (p *Plugin) gatherAssetSummary(accumulator telegraf.Accumulator, summary *AssetSummary) {
	accumulator.AddFields("rakuten_securities_asset_summaries", map[string]any{
		"estimate_amount":                   summary.EstimateAmount,
		"estimate_amount_change":            summary.EstimateAmountChange,
		"estimate_amount_change_rate":       summary.EstimateAmountChangeRate,
		"estimate_amount_month_change":      summary.EstimateAmountMonthChange,
		"estimate_amount_month_change_rate": summary.EstimateAmountMonthChangeRate,
		"estimate_profit_loss":              summary.EstimateProfitLoss,
		"estimate_profit_loss_rate":         summary.EstimateProfitLossRate,
		"realized_profit_loss":              summary.RealizedProfitLoss,
		"dividend":                          summary.Dividend,
	}, map[string]string{
		"title": summary.Title,
	})
}

func (p *Plugin) gatherAsset(accumulator telegraf.Accumulator, asset *Asset) {
	accumulator.AddFields("rakuten_securities_assets", map[string]any{
		"position":                  asset.Position,
		"average_cost":              asset.AverageCost,
		"price":                     asset.Price,
		"price_change":              asset.PriceChange,
		"estimate_amount":           asset.EstimateAmount,
		"estimate_profit_loss":      asset.EstimateProfitLoss,
		"estimate_profit_loss_rate": asset.EstimateProfitLossRate,
	}, map[string]string{
		"kind":    asset.Kind,
		"code":    asset.Code,
		"name":    asset.Name,
		"account": asset.Account,
	})
}

func (p *Plugin) gatherExchangeRate(accumulator telegraf.Accumulator, rate *ExchangeRate) {
	accumulator.AddFields("rakuten_securities_exchange_rates", map[string]any{
		"rate": rate.Rate,
	}, map[string]string{
		"currency_name": rate.CurrencyName,
		"unit":          rate.Unit,
	})
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
