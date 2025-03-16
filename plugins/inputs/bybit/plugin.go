package bybit

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
	client *BybitClient

	BybitAPIKey    string `toml:"-" env:"BYBIT_API_KEY"`
	ByBitAPISecret string `toml:"-" env:"BYBIT_API_SECRET"`
}

func init() {
	inputs.Add("bybit", func() telegraf.Input {
		return &Plugin{}
	})
}

func (p *Plugin) Init() error {
	if err := env.Parse(p); err != nil {
		return fmt.Errorf("failed to parse env: %w", err)
	}

	if p.BybitAPIKey == "" || p.ByBitAPISecret == "" {
		return errors.New("BYBIT_API_KEY and BYBIT_API_SECRET must be set")
	}

	p.client = NewBybitClient(p.BybitAPIKey, p.ByBitAPISecret)
	return nil
}

func (p *Plugin) SampleConfig() string {
	return sampleConfig
}

func (p *Plugin) Gather(accumulator telegraf.Accumulator) error {
	ctx := context.Background()
	wallets, err := p.client.GetAccountWallets(ctx)
	if err != nil {
		return err
	}

	for _, wallet := range wallets {
		p.gatherWallet(accumulator, wallet)
	}

	return nil
}

func (p *Plugin) gatherWallet(accumulator telegraf.Accumulator, wallet *AccountWallet) {
	accumulator.AddFields("bybit_wallet", map[string]any{
		"account_ltv":              parseFloat(wallet.AccountLTV),
		"account_im_rate":          parseFloat(wallet.AccountIMRate),
		"account_mm_rate":          parseFloat(wallet.AccountMMRate),
		"total_equity":             parseFloat(wallet.TotalEquity),
		"total_wallet_balance":     parseFloat(wallet.TotalWalletBalance),
		"total_margin_balance":     parseFloat(wallet.TotalMarginBalance),
		"total_available_balance":  parseFloat(wallet.TotalAvailableBalance),
		"total_perp_upl":           parseFloat(wallet.TotalPerpUPL),
		"total_initial_margin":     parseFloat(wallet.TotalInitialMargin),
		"total_maintenance_margin": parseFloat(wallet.TotalMaintenanceMargin),
	}, map[string]string{
		"account_type": wallet.AccountType,
	})

	for _, coin := range wallet.Coins {
		p.gatherCoin(accumulator, coin)
	}
}

func (p *Plugin) gatherCoin(accumulator telegraf.Accumulator, coin *Coin) {
	accumulator.AddFields("bybit_wallet_coins", map[string]any{
		"equity":                parseFloat(coin.Equity),
		"usd_value":             parseFloat(coin.UsdValue),
		"wallet_balance":        parseFloat(coin.WalletBalance),
		"locked":                parseFloat(coin.Locked),
		"spot_hedging_qty":      parseFloat(coin.SpotHedgingQty),
		"borrow_amount":         parseFloat(coin.BorrowAmount),
		"available_to_borrow":   parseFloat(coin.AvailableToBorrow),
		"available_to_withdraw": parseFloat(coin.AvailableToWithdraw),
		"accrued_interest":      parseFloat(coin.AccruedInterest),
		"total_order_im":        parseFloat(coin.TotalOrderIM),
		"total_position_im":     parseFloat(coin.TotalPositionIM),
		"total_position_mm":     parseFloat(coin.TotalPositionMM),
		"unrealised_pnl":        parseFloat(coin.UnrealisedPnl),
		"cum_realised_pnl":      parseFloat(coin.CumRealisedPnl),
		"bonus":                 parseFloat(coin.Bonus),
	}, map[string]string{
		"coin":              coin.Coin,
		"collateral_switch": strconv.FormatBool(coin.CollateralSwitch),
		"margin_collateral": strconv.FormatBool(coin.MarginCollateral),
	})
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

var (
	_ telegraf.Initializer = new(Plugin)
	_ telegraf.Input       = new(Plugin)
)
