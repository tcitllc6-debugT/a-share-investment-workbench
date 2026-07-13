package main

import (
	"testing"

	"go-stock/backend/data"
)

func TestInstrumentQuoteRequestKeepsSearchProviderMetadata(t *testing.T) {
	request := instrumentQuoteRequest(data.AssetInstrument{
		AssetType:    data.AssetTypePreciousMetal,
		Symbol:       "XAU",
		QuoteSymbol:  "XAU",
		Currency:     "USD",
		MetadataJSON: `{"sina_symbol":"hf_XAU","quote_provider":"sina"}`,
	})
	if request.Provider != "sina" || request.Meta("sina_symbol") != "hf_XAU" {
		t.Fatalf("search quote metadata was lost: %#v", request)
	}
}

func TestMajorFXCurrenciesAreStableAndIncludeCorePairs(t *testing.T) {
	seen := map[string]bool{}
	for _, currency := range majorFXCurrencies {
		seen[currency] = true
	}
	for _, expected := range []string{"USD", "HKD", "EUR", "GBP", "JPY", "AUD", "CAD", "SGD"} {
		if !seen[expected] {
			t.Fatalf("missing major FX currency %s: %#v", expected, majorFXCurrencies)
		}
	}
}
