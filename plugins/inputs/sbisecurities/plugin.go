package sbisecurities

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v11"
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/inputs"
)

//go:embed sample.conf
var sampleConfig string

type Plugin struct {
	client *SBISecuritiesClient

	Username     string `toml:"-" env:"SBI_SECURITIES_USERNAME"`
	Password     string `toml:"-" env:"SBI_SECURITIES_PASSWORD"`
	DeviceCookie string `toml:"-" env:"SBI_SECURITIES_DEVICE_COOKIE"`
}

func init() {
	inputs.Add("sbi_securities", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	var err error
	if err = env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	if p.Username == "" || p.Password == "" || p.DeviceCookie == "" {
		return errors.New("missing required environment variables")
	}

	p.client, err = NewSBISecuritiesClient()
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

	assets, err := p.client.GetFundAssets(ctx)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			if err := p.client.Login(ctx, p.Username, p.Password, p.DeviceCookie); err != nil {
				return err
			}
			return p.gather(ctx, accumulator, loop+1)
		}
		return fmt.Errorf("failed to get assets: %w", err)
	}

	p.gatherFundSummary(accumulator, assets)

	return nil
}

func (p *Plugin) gatherFundSummary(accumulator telegraf.Accumulator, assets *FundAssets) {
	accumulator.AddFields("sbi_securities_fund_summary", map[string]any{
		"cost":                      parseInt(assets.Data.CostSummary),               // 合計取得金額 (円)
		"estimate_amount":           parseInt(assets.Data.EstimateAmountSummary),     // 合計評価額 (円)
		"estimate_profit_loss_rate": assets.Data.EstimateProfitLossRateSummary,       // 合計評価損益 (%)
		"estimate_profit_loss":      parseInt(assets.Data.EstimateProfitLossSummary), // 合計評価損益 (円)
		"previous_change":           parseInt(assets.Data.PreviousChangeSummary),     // 合計評価額の前日比 (円)
		"previous_ratio":            assets.Data.PreviousRatioSummary,                // 合計評価額の前日比 (%)
		"funds_count":               assets.Data.TotalCount,                          // 保有ファンドの数
	}, nil)

	for _, holdings := range assets.Holdings() {
		for _, deposit := range holdings.Deposits() {
			p.gatherFundDeposit(accumulator, deposit, holdings.HoldingType)
		}
	}
}

func (p *Plugin) gatherFundDeposit(accumulator telegraf.Accumulator, deposit *FundDeposit, holdingType HoldingType) {
	if deposit == nil {
		return
	}

	accumulator.AddFields("sbi_securities_fund_deposits", map[string]any{
		"total_cost":               parseInt(deposit.CostTotal),             // 合計取得金額 (円)
		"total_estimate_amount":    parseInt(deposit.EstimateAmountTotal),   // 合計評価額 (円)
		"funds_count":              deposit.HitCount,                        // 保有ファンドの数
		"previous_change":          parseInt(deposit.PreviousChange),        // 合計評価額の前日比 (円)
		"previous_ratio":           deposit.PreviousRatio,                   // 合計評価額の前日比 (%)
		"total_profit_loss_amount": parseInt(deposit.ProfitLossAmountTotal), // 合計評価損益 (円)
		"total_profit_loss_rate":   deposit.ProfitLossRateTotal,             // 合計評価損益 (%)
	}, map[string]string{
		"holding_type":  string(holdingType),
		"holding_label": holdingType.Label(),
		"deposit_type":  string(deposit.DepositType),
		"deposit_label": deposit.DepositType.Label(),
	})

	for _, fund := range deposit.FundInfos {
		p.gatherFundInfo(accumulator, fund, holdingType, deposit.DepositType)
	}
}

func (p *Plugin) gatherFundInfo(accumulator telegraf.Accumulator, fund *FundInfo, holdingType HoldingType, depositType DepositType) {
	accumulator.AddFields("sbi_securities_funds", map[string]any{
		"cost":                               fund.Cost,                                 // 取得金額 (円)
		"estimate_amount":                    parseInt(fund.EstimateAmount),             // 評価額 (円)
		"estimate_amount_previous":           parseInt(fund.EstimateAmountPrevious),     // 前日の評価額 (円)
		"estimate_change":                    parseInt(fund.EstimateChange),             // 評価額の前日比 (円)
		"estimate_change_rate":               fund.EstimateChangeRate,                   // 評価額の前日比 (%)
		"estimate_profit_loss":               parseInt(fund.EstimateProfitLoss),         // 評価損益 (円)
		"estimate_profit_loss_previous":      parseInt(fund.EstimateProfitLossPrevious), // 前日の評価損益 (円)
		"estimate_profit_loss_rate":          fund.EstimateProfitLossRate,               // 評価損益 (%)
		"estimate_profit_loss_rate_previous": fund.EstimateProfitLossRatePrevious,       // 前日の評価損益 (%)
		"position":                           parseInt(fund.Position),                   // 保有口数 (口)
		"price":                              fund.Price,                                // 取得単価 (円)
		"standard_price":                     fund.StandardPrice,                        // 基準価額 (円)
		"standard_price_previous":            fund.StandardPricePrevious,                // 前日の基準価額 (円)
	}, map[string]string{
		"holding_type":  string(holdingType),
		"holding_label": holdingType.Label(),
		"deposit_type":  string(depositType),
		"deposit_label": depositType.Label(),
		"fund_name":     fund.FundName,
		"fund_code":     fund.AssociationCode,
	})
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
