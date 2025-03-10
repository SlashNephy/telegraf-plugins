package sbisecurities

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var ErrUnauthorized = errors.New("unauthorized")

type SBISecuritiesClient struct {
	httpClient *http.Client
	csrfToken  string
}

func NewSBISecuritiesClient() (*SBISecuritiesClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &SBISecuritiesClient{
		httpClient: &http.Client{
			Jar: jar,
		},
	}, nil
}

type AccountInfo struct {
	Status string `json:"status"`
}

func (c *SBISecuritiesClient) Login(ctx context.Context, username, password string) error {
	// GET https://site1.sbisec.co.jp/ETGate/
	{
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://site1.sbisec.co.jp/ETGate/", nil)
		if err != nil {
			return err
		}

		request.Header.Set("Sec-Ch-Ua", `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", "document")
		request.Header.Set("Sec-Fetch-Mode", "navigate")
		request.Header.Set("Sec-Fetch-Site", "same-origin")
		request.Header.Set("Sec-Fetch-User", "?1")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response: %s", response.Status)
		}
	}

	// POST https://site1.sbisec.co.jp/ETGate/
	{
		formData := url.Values{
			"JS_FLG":          []string{"1"},
			"BW_FLG":          []string{"chrome,134"},
			"_ControlID":      []string{"WPLETlgR001Control"},
			"_DataStoreID":    []string{"DSWPLETlgR001Control"},
			"_PageID":         []string{"WPLETlgR001Rlgn20"},
			"_ActionID":       []string{"login"},
			"getFlg":          []string{"on"},
			"allPrmFlg":       []string{"on"},
			"_ReturnPageInfo": []string{"WPLEThmR001Control/DefaultPID/DefaultAID/DSWPLEThmR001Control"},
			"user_id":         []string{username},
			"user_password":   []string{password},
			"ACT_login":       []string{"%83%8D%83O%83C%83%93"},
		}

		body := strings.NewReader(formData.Encode())
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://site1.sbisec.co.jp/ETGate/", body)
		if err != nil {
			return err
		}

		request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		request.Header.Set("Origin", "https://site0.sbisec.co.jp")
		request.Header.Set("Referer", "https://site0.sbisec.co.jp/")
		request.Header.Set("Sec-Ch-Ua", `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", "document")
		request.Header.Set("Sec-Fetch-Mode", "navigate")
		request.Header.Set("Sec-Fetch-Site", "same-origin")
		request.Header.Set("Sec-Fetch-User", "?1")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response: %s", response.Status)
		}
	}

	// GET https://www.sbisec.co.jp/ETGate/?_ControlID=WPLETsmR001Control&_PageID=WPLETsmR001Sdtl23&_DataStoreID=DSWPLETsmR001Control&_ActionID=NoActionID&getFlg=on&OutSide=on&path=fund%2Ftop
	{
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.sbisec.co.jp/ETGate/?_ControlID=WPLETsmR001Control&_PageID=WPLETsmR001Sdtl23&_DataStoreID=DSWPLETsmR001Control&_ActionID=NoActionID&getFlg=on&OutSide=on&path=fund%2Ftop", nil)
		if err != nil {
			return err
		}

		request.Header.Set("Sec-Ch-Ua", `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", "document")
		request.Header.Set("Sec-Fetch-Mode", "navigate")
		request.Header.Set("Sec-Fetch-Site", "none")
		request.Header.Set("Sec-Fetch-User", "?1")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response: %s", response.Status)
		}

		document, err := goquery.NewDocumentFromReader(response.Body)
		if err != nil {
			return err
		}

		token, ok := document.Find(`[name="_csrf"]`).Attr("content")
		if !ok {
			return errors.New("csrf token not found")
		}
		c.csrfToken = token
	}

	// GET https://member.c.sbisec.co.jp/system/api/account/info
	{
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://member.c.sbisec.co.jp/system/api/account/info", nil)
		if err != nil {
			return err
		}

		request.Header.Set("Accept", "application/json; charset=utf-8")
		request.Header.Set("Referer", "https://member.c.sbisec.co.jp/fund/top")
		request.Header.Set("Sec-Ch-Ua", `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`)
		request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
		request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
		request.Header.Set("Sec-Fetch-Dest", "empty")
		request.Header.Set("Sec-Fetch-Mode", "cors")
		request.Header.Set("Sec-Fetch-Site", "same-origin")
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected response: %s", response.Status)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		var info AccountInfo
		if err = json.Unmarshal(body, &info); err != nil {
			return err
		}

		if info.Status != "SUCCESS" {
			return fmt.Errorf("unexpected status: %s", info.Status)
		}
	}

	return nil
}

type FundAssets struct {
	Data struct {
		AmountHoldings                *FundHoldings `json:"amountHoldings"`                // 金額指定保有
		CostSummary                   string        `json:"costSummary"`                   // 合計取得金額 (円)
		EstimateAmountSummary         string        `json:"estimateAmountSummary"`         // 合計評価額 (円)
		EstimateProfitLossRateSummary float64       `json:"estimateProfitLossRateSummary"` // 合計評価損益 (%)
		EstimateProfitLossSummary     string        `json:"estimateProfitLossSummary"`     // 合計評価損益 (円)
		JrNISAOpened                  bool          `json:"jrnisaOpened"`                  // ?
		JrNISASpecificOpened          bool          `json:"jrnisaSpecificOpened"`          // ?
		PreviousChangeSummary         string        `json:"previousChangeSummary"`         // 合計評価額の前日比 (円)
		PreviousRatioSummary          float64       `json:"previousRatioSummary"`          // 合計評価額の前日比 (%)
		SpecificOpened                bool          `json:"specificOpened"`                // 特定口座を持っているか
		TotalCount                    int           `json:"totalCount"`                    // 保有ファンドの数
		UnitHoldings                  *FundHoldings `json:"unitHoldings"`                  // 口数指定保有
	} `json:"data"`
	Status string `json:"status"`
}

func (a *FundAssets) Holdings() []*FundHoldings {
	return []*FundHoldings{a.Data.AmountHoldings, a.Data.UnitHoldings}
}

type FundHoldings struct {
	JuniorNisaContinuous      any          `json:"juniorNisaContinuous"`      // ?
	JuniornisaNisaDeposit     any          `json:"juniornisaNisaDeposit"`     // ?
	JuniornisaNormalDeposit   any          `json:"juniornisaNormalDeposit"`   // ?
	JuniornisaSpecificDeposit any          `json:"juniornisaSpecificDeposit"` // ?
	NisaDeposit               any          `json:"nisaDeposit"`               // ?
	NisaGrowth                *FundDeposit `json:"nisaGrowth"`                // NISA (成長)預り
	NisaReserve               *FundDeposit `json:"nisaReserve"`               // NISA (つみたて)預り
	NormalDeposit             *FundDeposit `json:"normalDeposit"`             // 普通預り
	SpecificDeposit           *FundDeposit `json:"specificDeposit"`           // 特定預り
	TnisaDeposit              *FundDeposit `json:"tnisaDeposit"`              // 旧つみたてNISA預り

	HoldingType HoldingType `json:"-"`
}

func (h *FundHoldings) Deposits() []*FundDeposit {
	return []*FundDeposit{
		h.NisaGrowth,
		h.NisaReserve,
		h.NormalDeposit,
		h.SpecificDeposit,
		h.TnisaDeposit,
	}
}

type HoldingType string

const (
	HoldingTypeAmount HoldingType = "amount" // 金額指定保有
	HoldingTypeUnit   HoldingType = "unit"   // 口数指定保有
)

func (t HoldingType) Label() string {
	switch t {
	case HoldingTypeAmount:
		return "金額指定保有"
	case HoldingTypeUnit:
		return "口数指定保有"
	default:
		panic(fmt.Sprintf("unexpected holding type: %s", t))
	}
}

type FundDeposit struct {
	CostTotal             string      `json:"costTotal"`           // 合計取得金額 (円)
	EstimateAmountTotal   string      `json:"estimateAmountTotal"` // 合計評価額 (円)
	FundInfos             []*FundInfo `json:"fundInfos"`
	HitCount              int         `json:"hitCount"`              // 保有ファンドの数
	PreviousChange        string      `json:"previousChange"`        // 合計評価額の前日比 (円)
	PreviousRatio         float64     `json:"previousRatio"`         // 合計評価額の前日比 (%)
	ProfitLossAmountTotal string      `json:"profitLossAmountTotal"` // 合計評価損益 (円)
	ProfitLossRateTotal   float64     `json:"profitLossRateTotal"`   // 合計評価損益 (%)

	DepositType DepositType `json:"-"`
}

type DepositType string

const (
	DepositTypeNisaGrowth  DepositType = "nisa_growth"  // NISA (成長)預り
	DepositTypeNisaReserve DepositType = "nisa_reserve" // NISA (つみたて)預り
	DepositTypeNormal      DepositType = "normal"       // 普通預り
	DepositTypeSpecific    DepositType = "specific"     // 特定預り
	DepositTypeTnisa       DepositType = "tnisa"        // 旧つみたてNISA預り
)

func (t DepositType) Label() string {
	switch t {
	case DepositTypeNisaGrowth:
		return "NISA (成長投資枠)"
	case DepositTypeNisaReserve:
		return "NISA (つみたて投資枠)"
	case DepositTypeNormal:
		return "普通預り"
	case DepositTypeSpecific:
		return "特定預り"
	case DepositTypeTnisa:
		return "旧つみたてNISA預り"
	default:
		panic(fmt.Sprintf("unexpected deposit type: %s", t))
	}
}

type FundInfo struct {
	AmountBuyable                    string  `json:"amountBuyable"`                    // ?
	AssociationCode                  string  `json:"associationCode"`                  // https://member.c.sbisec.co.jp/fund/detail/${AssociationCode}
	CancellationOrderDuringUnit      string  `json:"cancellationOrderDuringUnit"`      // ?
	CancellationOrderDuringValuation string  `json:"cancellationOrderDuringValuation"` // ?
	Cost                             int     `json:"cost"`                             // 取得金額 (円)
	DividendChangeType               string  `json:"dividendChangeType"`               // ?
	EstimateAmount                   string  `json:"estimateAmount"`                   // 評価額 (円)
	EstimateAmountPrevious           string  `json:"estimateAmountPrivious"`           // 前日の評価額 (円)
	EstimateChange                   string  `json:"estimateChange"`                   // 評価額の前日比 (円)
	EstimateChangeRate               float64 `json:"estimateChangeRate"`               // 評価額の前日比 (%)
	EstimateProfitLoss               string  `json:"estimateProfitLoss"`               // 評価損益 (円)
	EstimateProfitLossPrevious       string  `json:"estimateProfitLossPrivious"`       // 前日の評価損益 (円)
	EstimateProfitLossRate           float64 `json:"estimateProfitLossRate"`           // 評価損益 / 取得金額 (%)
	EstimateProfitLossRatePrevious   float64 `json:"estimateProfitLossRatePrivious"`   // 前日の評価損益 / 取得金額 (%)
	FavoriteFlag                     bool    `json:"favoriteFlag"`                     // ?
	FundName                         string  `json:"fundName"`                         // ファンド名
	FundType                         string  `json:"fundType"`                         // ?
	KaisuGou                         string  `json:"kaisuGou"`                         // ?
	MsCategoryCode                   string  `json:"msCategoryCode"`                   // ?
	ParticularPrincipal              float64 `json:"particularPrincipal"`              // ?
	Position                         string  `json:"position"`                         // 保有口数 (口)
	Price                            float64 `json:"price"`                            // 取得単価 (円)
	Reinvest                         string  `json:"reinvest"`                         // ?
	ReserveBuyable                   string  `json:"reserveBuyable"`                   // ?
	ReserveSellable                  string  `json:"reserveSellable"`                  // ?
	StandardPrice                    int     `json:"standardPrice"`                    // 基準価額 (円)
	StandardPriceDate                string  `json:"standardPriceDate"`                // yyyymmdd
	StandardPricePrevious            int     `json:"standardPricePrivious"`            // 前日の基準価額 (円)
	StandardPriceUnit                int     `json:"standardPriceUnit"`                // ?
	UnitBuyable                      string  `json:"unitBuyable"`                      // ?
	WebHandlingType                  string  `json:"webHandlingType"`                  // ?
}

func (c *SBISecuritiesClient) GetFundAssets(ctx context.Context) (*FundAssets, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://member.c.sbisec.co.jp/fund/api/account/assets?accountGetType=2", nil)
	if err != nil {
		return nil, err
	}

	request.Header.Set("Accept", "application/json; charset=utf-8")
	request.Header.Set("Accept-Language", "ja,en;q=0.9")
	request.Header.Set("Dnt", "1")
	request.Header.Set("Referer", "https://member.c.sbisec.co.jp/fund/account/assets")
	request.Header.Set("Sec-Ch-Ua", `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`)
	request.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	request.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	request.Header.Set("Sec-Fetch-Dest", "empty")
	request.Header.Set("Sec-Fetch-Mode", "cors")
	request.Header.Set("Sec-Fetch-Site", "same-origin")
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0")
	request.Header.Set("x-csrf-token", c.csrfToken)

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		if response.StatusCode == http.StatusUnauthorized {
			return nil, ErrUnauthorized
		}
		return nil, fmt.Errorf("unexpected response: %s", response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var assets FundAssets
	if err = json.Unmarshal(body, &assets); err != nil {
		return nil, err
	}

	if assets.Status != "SUCCESS" {
		return nil, fmt.Errorf("unexpected status: %s", assets.Status)
	}

	// 走査するときに不便なので HoldingType と DepositType を埋め込む
	{
		assets.Data.AmountHoldings.HoldingType = HoldingTypeAmount
		assets.Data.UnitHoldings.HoldingType = HoldingTypeUnit
	}
	{
		setDepositType := func(deposit *FundDeposit, depositType DepositType) {
			if deposit != nil {
				deposit.DepositType = depositType
			}
		}

		setDepositType(assets.Data.AmountHoldings.NisaGrowth, DepositTypeNisaGrowth)
		setDepositType(assets.Data.AmountHoldings.NisaReserve, DepositTypeNisaReserve)
		setDepositType(assets.Data.AmountHoldings.NormalDeposit, DepositTypeNormal)
		setDepositType(assets.Data.AmountHoldings.SpecificDeposit, DepositTypeSpecific)
		setDepositType(assets.Data.AmountHoldings.TnisaDeposit, DepositTypeTnisa)

		setDepositType(assets.Data.UnitHoldings.NisaGrowth, DepositTypeNisaGrowth)
		setDepositType(assets.Data.UnitHoldings.NisaReserve, DepositTypeNisaReserve)
		setDepositType(assets.Data.UnitHoldings.NormalDeposit, DepositTypeNormal)
		setDepositType(assets.Data.UnitHoldings.SpecificDeposit, DepositTypeSpecific)
		setDepositType(assets.Data.UnitHoldings.TnisaDeposit, DepositTypeTnisa)
	}

	return &assets, nil
}
