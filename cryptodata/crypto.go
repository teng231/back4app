package cryptodata

import (
	"encoding/json"
	"log"

	"github.com/teng231/gotools/v2/httpclient"
)

type TokenDetail struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Symbol    string  `json:"symbol"`
	Dominance float64 `json:"dominance"`
	Image     string  `json:"image"`
	Rank      int     `json:"rank"`
	Stable    bool    `json:"stable"`
	Price     float64 `json:"price"`
	Marketcap int64   `json:"marketcap"`
	Volume    int64   `json:"volume"`
	CgID      string  `json:"cg_id"`
	// Symbols   struct {
	// 	Gateio   string `json:"gateio"`
	// 	Coinbase string `json:"coinbase"`
	// 	Kucoin   string `json:"kucoin"`
	// 	Bingx    string `json:"bingx"`
	// 	Binance  string `json:"binance"`
	// 	Bybit    string `json:"bybit"`
	// 	Bitget   string `json:"bitget"`
	// 	Mexc     string `json:"mexc"`
	// 	Bitmart  string `json:"bitmart"`
	// 	Okx      string `json:"okx"`
	// } `json:"symbols"`
	// Performance struct {
	// 	Month float64 `json:"month"`
	// 	Day   float64 `json:"day"`
	// 	Min5  float64 `json:"min5"`
	// 	Min1  float64 `json:"min1"`
	// 	Min15 float64 `json:"min15"`
	// 	Hour  float64 `json:"hour"`
	// 	Year  float64 `json:"year"`
	// 	Week  float64 `json:"week"`
	// } `json:"performance"`
	// RankDiffs struct {
	// 	Month int `json:"month"`
	// 	Hour  int `json:"hour"`
	// 	Week  int `json:"week"`
	// 	Day   int `json:"day"`
	// 	Year  int `json:"year"`
	// } `json:"rankDiffs"`
	// ExchangePrices struct {
	// 	Binance  float64 `json:"binance"`
	// 	Bybit    float64 `json:"bybit"`
	// 	Mexc     float64 `json:"mexc"`
	// 	Okx      float64 `json:"okx"`
	// 	Kucoin   float64 `json:"kucoin"`
	// 	Bingx    float64 `json:"bingx"`
	// 	Bitmart  float64 `json:"bitmart"`
	// 	Coinbase float64 `json:"coinbase"`
	// 	Bitget   float64 `json:"bitget"`
	// 	Gateio   float64 `json:"gateio"`
	// } `json:"exchangePrices"`
}

var (
	cryptoMap map[string]*TokenDetail
)

func ListCrytos() map[string]*TokenDetail {
	if len(cryptoMap) != 0 {
		return cryptoMap
	}

	resp, err := httpclient.Exec("https://cryptobubbles.net/backend/data/bubbles1000.usd.json",
		httpclient.WithHeader(map[string]string{
			"accept":             `*/*`,
			"accept-language":    `vi,en-US;q=0.9,en;q=0.8,la;q=0.7,ko;q=0.6,it;q=0.5,ja;q=0.4,und;q=0.3`,
			"priority":           `u=1, i`,
			"referer":            `https://cryptobubbles.net/`,
			"sec-ch-ua":          `"Google Chrome";v="129", "Not=A?Brand";v="8", "Chromium";v="129"`,
			"sec-ch-ua-mobile":   `?0`,
			"sec-ch-ua-platform": `"macOS"`,
			"sec-fetch-dest":     `empty`,
			"sec-fetch-mode":     `cors`,
			"sec-fetch-site":     `same-origin`,
			"user-agent":         `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36`,
		}),
	)
	if err != nil {
		log.Print(err)
		return map[string]*TokenDetail{}
	}
	if resp.HttpCode != 200 {
		log.Print(err)
		return map[string]*TokenDetail{}
	}
	cryptodatas := make([]*TokenDetail, 0)
	json.Unmarshal(resp.Body, &cryptodatas)
	cryptoMap = make(map[string]*TokenDetail)
	for _, crytoVal := range cryptodatas {
		cryptoMap[crytoVal.Symbol] = crytoVal
	}
	return cryptoMap
}
