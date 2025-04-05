package rakutensecurities

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var ErrUnauthorized = errors.New("unauthorized")

var locationRegexp = regexp.MustCompile(`location\.href\s*=\s*"([^"]+)"`)

type RakutenSecuritiesClient struct {
	httpClient *http.Client
	sessionID  string
}

func NewRakutenSecuritiesClient() (*RakutenSecuritiesClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &RakutenSecuritiesClient{
		httpClient: &http.Client{
			Jar: jar,
		},
	}, nil
}

func (c *RakutenSecuritiesClient) Login(ctx context.Context, username, password string) error {
	// GET https://www.rakuten-sec.co.jp/
	{
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.rakuten-sec.co.jp/", nil)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ja",
			"Priority":                  "u=0, i",
			"Referer":                   "https://www.rakuten-sec.co.jp/",
			"Sec-Ch-Ua":                 `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "none",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}
	}

	// POST https://member.rakuten-sec.co.jp/app/MhLogin.do
	var location string
	{
		requestForm := url.Values{}
		requestForm.Set("loginid", username)
		requestForm.Set("passwd", password)
		requestForm.Set("homeid", "HOME")
		requestBody := strings.NewReader(requestForm.Encode())

		request, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://member.rakuten-sec.co.jp/app/MhLogin.do", requestBody)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ja",
			"Cache-Control":             "max-age=0",
			"Content-Type":              "application/x-www-form-urlencoded",
			"Origin":                    "https://www.rakuten-sec.co.jp",
			"Priority":                  "u=0, i",
			"Referer":                   "https://www.rakuten-sec.co.jp/",
			"Sec-Ch-Ua":                 `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}

		defer response.Body.Close()
		responseBody, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		match := locationRegexp.FindSubmatch(responseBody)
		if len(match) != 2 {
			return fmt.Errorf("unexpected response: %s", responseBody)
		}

		location = string(match[1])
		u, err := url.Parse(location)
		if err != nil {
			return err
		}

		c.sessionID = u.Query().Get("BV_SessionID")
		if c.sessionID == "" {
			return errors.New("session id not found")
		}
	}

	// GET https://member.rakuten-sec.co.jp${location}
	{
		url := fmt.Sprintf("https://member.rakuten-sec.co.jp%s", location)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ja",
			"Priority":                  "u=0, i",
			"Referer":                   "https://member.rakuten-sec.co.jp/app/MhLogin.do",
			"Sec-Ch-Ua":                 `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}
	}

	// GET https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do?eventType=directInit
	{
		url := fmt.Sprintf("https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do;BV_SessionID=%s?eventType=directInit&l-id=mem_pc_top_all-possess-lst&gmn=H&smn=01&lmn=&fmn=", c.sessionID)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ja",
			"Priority":                  "u=0, i",
			"Referer":                   fmt.Sprintf("https://member.rakuten-sec.co.jp/app/home.do;BV_SessionID=%s?eventType=init&BV_SessionID=%s", c.sessionID, c.sessionID),
			"Sec-Ch-Ua":                 `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}
	}

	// POST https://member.rakuten-sec.co.jp/app/async_all_possess_lst_balance_lst.do
	{
		url := fmt.Sprintf("https://member.rakuten-sec.co.jp/app/async_all_possess_lst_balance_lst.do;BV_SessionID=%s?assetCloseFlg=1", c.sessionID)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":             "*/*",
			"Accept-Language":    "ja",
			"Origin":             "https://member.rakuten-sec.co.jp",
			"Priority":           "u=1, i",
			"Referer":            fmt.Sprintf("https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do;BV_SessionID=%s?eventType=directInit&l-id=mem_pc_top_all-possess-lst&gmn=H&smn=01&lmn=&fmn=", c.sessionID),
			"Sec-Ch-Ua":          `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":   "?0",
			"Sec-Ch-Ua-Platform": `"Windows"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
			"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
			"X-Requested-With":   "XMLHttpRequest",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}
	}

	// POST https://member.rakuten-sec.co.jp/app/async_all_possess_lst_pos_lst.do
	{
		url := fmt.Sprintf("https://member.rakuten-sec.co.jp/app/async_all_possess_lst_pos_lst.do;BV_SessionID=%s?assetCloseFlg=1", c.sessionID)
		request, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
		if err != nil {
			return err
		}

		headers := map[string]string{
			"Accept":             "*/*",
			"Accept-Language":    "ja",
			"Origin":             "https://member.rakuten-sec.co.jp",
			"Priority":           "u=1, i",
			"Referer":            fmt.Sprintf("https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do;BV_SessionID=%s?eventType=directInit&l-id=mem_pc_top_all-possess-lst&gmn=H&smn=01&lmn=&fmn=", c.sessionID),
			"Sec-Ch-Ua":          `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":   "?0",
			"Sec-Ch-Ua-Platform": `"Windows"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
			"User-Agent":         "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
			"X-Requested-With":   "XMLHttpRequest",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return err
		}

		if response.StatusCode != http.StatusOK {
			return fmt.Errorf("unexpected status: %s", response.Status)
		}
	}

	return nil
}

type Metrics struct {
	Summaries     []*AssetSummary
	Assets        []*Asset
	ExchangeRates []*ExchangeRate
}

type AssetSummary struct {
	Title                         string
	EstimateAmount                int     // 評価額 (円)
	EstimateAmountChange          int     // 評価額の前日比 (円)
	EstimateAmountChangeRate      float64 // 評価額の前日比 (%)
	EstimateAmountMonthChange     int     // 評価額の前月比 (円)
	EstimateAmountMonthChangeRate float64 // 評価額の前月比 (円)
	EstimateProfitLoss            int     // 評価損益 (円)
	EstimateProfitLossRate        float64 // 評価損益率 (%)
	RealizedProfitLoss            int     // 実現損益 (円)
	Dividend                      int     // 配当・分配金 (円)
}

type Asset struct {
	Kind                   string  // 種別
	Code                   string  // 銘柄コード・ティッカー
	Name                   string  // 銘柄
	Account                string  // 口座
	Position               int     // 保有数量 (口)
	AverageCost            float64 // 平均取得価額 (円)
	Price                  int     // 基準価額 (円)
	PriceChange            int     // 基準価額の前日比 (円)
	EstimateAmount         int     // 評価額 (円)
	EstimateProfitLoss     int     // 評価損益 (円)
	EstimateProfitLossRate float64 // 評価損益率 (%)
}
type ExchangeRate struct {
	CurrencyName string
	Rate         float64
	Unit         string
}

func (c *RakutenSecuritiesClient) GetMetrics(ctx context.Context) (*Metrics, error) {
	if c.sessionID == "" {
		return nil, ErrUnauthorized
	}

	// GET https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do?eventType=csv
	{
		url := fmt.Sprintf("https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do;BV_SessionID=%s?eventType=csv", c.sessionID)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}

		headers := map[string]string{
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
			"Accept-Language":           "ja",
			"Priority":                  "u=0, i",
			"Referer":                   fmt.Sprintf("https://member.rakuten-sec.co.jp/app/ass_all_possess_lst.do;BV_SessionID=%s?eventType=directInit&l-id=mem_pc_top_all-possess-lst&gmn=H&smn=01&lmn=&fmn=", c.sessionID),
			"Sec-Ch-Ua":                 `"Chromium";v="134", "Not:A-Brand";v="24", "Microsoft Edge";v="134"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
			"Sec-Fetch-Dest":            "document",
			"Sec-Fetch-Mode":            "navigate",
			"Sec-Fetch-Site":            "same-origin",
			"Sec-Fetch-User":            "?1",
			"Upgrade-Insecure-Requests": "1",
			"User-Agent":                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36 Edg/134.0.0.0",
		}
		for k, v := range headers {
			request.Header.Set(k, v)
		}

		response, err := c.httpClient.Do(request)
		if err != nil {
			return nil, err
		}

		mediaType, _, err := mime.ParseMediaType(response.Header.Get("Content-Type"))
		if err != nil {
			return nil, err
		}
		if mediaType != "text/comma-separated-values" {
			return nil, ErrUnauthorized
		}

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("unexpected status: %s", response.Status)
		}

		defer response.Body.Close()
		r := transform.NewReader(response.Body, japanese.ShiftJIS.NewDecoder()) // Shift-JIS -> UTF-8
		return c.parseCSV(r)
	}
}

func (c *RakutenSecuritiesClient) parseCSV(r io.Reader) (*Metrics, error) {
	reader := csv.NewReader(r)

	var metrics Metrics
	var mode parserMode
	for {
		line, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}

		if len(line) == 0 {
			continue
		}

		if len(line) == 1 {
			switch line[0] {
			case "■資産合計欄":
				mode = 1
			case "■ 保有商品詳細 (すべて）":
				mode = 2
			case "■参考為替レート":
				mode = 3
			}
			continue
		}

		switch mode {
		case parserModeTotalAssets:
			// "","時価評価額[円]","前日比[円]","前日比[％]","前月比[円]","前月比[％]","評価損益[円]","評価損益[％]","","実現損益[円]","配当・分配金[円貨]","配当・分配金[外貨]"
			if line[0] == "" {
				continue
			}

			// 楽天銀行普通預金残高 未取得
			if line[0] == "楽天銀行普通預金残高" && line[1] == "未取得" {
				continue
			}

			if len(line) != 12 {
				slog.Warn("illegal line while parsing total assets", slog.Any("line", line))
				continue
			}

			metrics.Summaries = append(metrics.Summaries, &AssetSummary{
				Title:                         line[0],
				EstimateAmount:                parseInt(line[1]),
				EstimateAmountChange:          parseInt(line[2]),
				EstimateAmountChangeRate:      parseFloat(line[3]),
				EstimateAmountMonthChange:     parseInt(line[4]),
				EstimateAmountMonthChangeRate: parseFloat(line[5]),
				EstimateProfitLoss:            parseInt(line[6]),
				EstimateProfitLossRate:        parseFloat(line[7]),
				RealizedProfitLoss:            parseInt(line[9]),
				Dividend:                      parseInt(line[10]),
			})

		case parserModePossession:
			// "種別","銘柄コード・ティッカー","銘柄","口座","保有数量","［単位］","平均取得価額","［単位］","現在値","［単位］","現在値(更新日)","(参考為替)","前日比","［単位］","時価評価額[円]","時価評価額[外貨]","評価損益[円]","評価損益[％]"
			if line[0] == "種別" {
				continue
			}

			if len(line) != 18 {
				slog.Warn("illegal line while parsing possession", slog.Any("line", line))
				continue
			}

			metrics.Assets = append(metrics.Assets, &Asset{
				Kind:                   line[0],
				Code:                   line[1],
				Name:                   line[2],
				Account:                line[3],
				Position:               parseInt(line[4]),
				AverageCost:            parseFloat(line[6]),
				Price:                  parseInt(line[8]),
				PriceChange:            parseInt(line[12]),
				EstimateAmount:         parseInt(line[14]),
				EstimateProfitLoss:     parseInt(line[16]),
				EstimateProfitLossRate: parseFloat(line[17]),
			})

		case parserModeExchangeRate:
			// "米ドル","146.26","円/USD","(04/04  01:20)"
			if len(line) != 4 {
				slog.Warn("illegal line while parsing exchange rate", slog.Any("line", line))
				continue
			}

			metrics.ExchangeRates = append(metrics.ExchangeRates, &ExchangeRate{
				CurrencyName: line[0],
				Rate:         parseFloat(line[1]),
				Unit:         line[2],
			})
		}
	}

	return &metrics, nil
}

var replacer = strings.NewReplacer(
	",", "",
	"+", "",
)

func parseInt(s string) int {
	if s == "" || s == "-" {
		return 0
	}

	i, _ := strconv.Atoi(replacer.Replace(s))
	return i
}

func parseFloat(s string) float64 {
	if s == "" || s == "-" {
		return 0
	}

	f, _ := strconv.ParseFloat(replacer.Replace(s), 64)
	return f
}

type parserMode int

const (
	parserModeTotalAssets  parserMode = 1
	parserModePossession   parserMode = 2
	parserModeExchangeRate parserMode = 3
)
