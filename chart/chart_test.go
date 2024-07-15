package chart

import (
	"testing"

	"github.com/teng231/back4app/ledger"
)

func TestMakeChart(t *testing.T) {
	MakePieDataChartToHTML([]*ledger.Holding{
		{Symbol: "ENA", Amount: 300, TVL: 100},
		{Symbol: "ETH", Amount: 20, TVL: 400},
		{Symbol: "BNB", Amount: 20, TVL: 120},
		{Symbol: "UNI", Amount: 20, TVL: 80},
	})
}

func TestMakeChartV2(t *testing.T) {
	MakePieDataChartToImg("./pie_holding_data.png", []*ledger.Holding{
		{Symbol: "ENA", Amount: 300, TVL: 100},
		{Symbol: "ETH", Amount: 20, TVL: 400},
		{Symbol: "BNB", Amount: 20, TVL: 120},
		{Symbol: "LINK", Amount: 20, TVL: 30},
		{Symbol: "UNI", Amount: 20, TVL: 80},
	})
}
