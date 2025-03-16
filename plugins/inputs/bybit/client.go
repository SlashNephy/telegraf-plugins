package bybit

import (
	"context"
	"encoding/json"
	"fmt"

	bybit_connector "github.com/bybit-exchange/bybit.go.api"
)

type BybitClient struct {
	client *bybit_connector.Client
}

func NewBybitClient(apiKey, apiSecret string) *BybitClient {
	return &BybitClient{
		client: bybit_connector.NewBybitHttpClient(
			apiKey,
			apiSecret,
			bybit_connector.WithBaseURL(bybit_connector.MAINNET),
		),
	}
}

type AccountWalletResponse struct {
	Lists []*AccountWallet `json:"list"`
}

type AccountWallet struct {
	AccountIMRate          string  `json:"accountIMRate"`
	AccountLTV             string  `json:"accountLTV"`
	AccountMMRate          string  `json:"accountMMRate"`
	AccountType            string  `json:"accountType"`
	Coins                  []*Coin `json:"coin"`
	TotalAvailableBalance  string  `json:"totalAvailableBalance"`
	TotalEquity            string  `json:"totalEquity"`
	TotalInitialMargin     string  `json:"totalInitialMargin"`
	TotalMaintenanceMargin string  `json:"totalMaintenanceMargin"`
	TotalMarginBalance     string  `json:"totalMarginBalance"`
	TotalPerpUPL           string  `json:"totalPerpUPL"`
	TotalWalletBalance     string  `json:"totalWalletBalance"`
}

type Coin struct {
	AccruedInterest     string `json:"accruedInterest"`
	AvailableToBorrow   string `json:"availableToBorrow"`
	AvailableToWithdraw string `json:"availableToWithdraw"`
	Bonus               string `json:"bonus"`
	BorrowAmount        string `json:"borrowAmount"`
	Coin                string `json:"coin"`
	CollateralSwitch    bool   `json:"collateralSwitch"`
	CumRealisedPnl      string `json:"cumRealisedPnl"`
	Equity              string `json:"equity"`
	Locked              string `json:"locked"`
	MarginCollateral    bool   `json:"marginCollateral"`
	SpotHedgingQty      string `json:"spotHedgingQty"`
	TotalOrderIM        string `json:"totalOrderIM"`
	TotalPositionIM     string `json:"totalPositionIM"`
	TotalPositionMM     string `json:"totalPositionMM"`
	UnrealisedPnl       string `json:"unrealisedPnl"`
	UsdValue            string `json:"usdValue"`
	WalletBalance       string `json:"walletBalance"`
}

func (c *BybitClient) GetAccountWallets(ctx context.Context) ([]*AccountWallet, error) {
	params := c.client.NewUtaBybitServiceWithParams(map[string]any{
		"accountType": "UNIFIED",
	})
	rawResponse, err := params.GetAccountWallet(ctx)
	if err != nil {
		return nil, err
	}

	rawResult, err := json.Marshal(rawResponse.Result)
	if err != nil {
		return nil, err
	}

	if rawResponse.RetMsg != "OK" {
		return nil, fmt.Errorf("error: %s", rawResponse.RetMsg)
	}

	var response AccountWalletResponse
	if err = json.Unmarshal(rawResult, &response); err != nil {
		return nil, err
	}

	return response.Lists, nil
}
