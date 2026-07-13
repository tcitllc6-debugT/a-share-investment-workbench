package assetsearch

import "net/http"

const (
	AssetStock         = "stock"
	AssetFund          = "fund"
	AssetBond          = "bond"
	AssetPreciousMetal = "precious_metal"
	AssetForex         = "forex"
	AssetFutures       = "futures"
	AssetCrypto        = "crypto"
	AssetBankWealth    = "bank_wealth"
)

type Instrument struct {
	Symbol      string            `json:"symbol"`
	Name        string            `json:"name"`
	Exchange    string            `json:"exchange"`
	QuoteSymbol string            `json:"quoteSymbol"`
	Currency    string            `json:"currency"`
	Unit        string            `json:"unit"`
	Source      string            `json:"source"`
	Category    string            `json:"category"`
	Metadata    map[string]string `json:"metadata"`
}

type Config struct {
	Client       *http.Client
	EastmoneyURL string
	BinanceURL   string
}
