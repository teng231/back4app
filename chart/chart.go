package chart

import (
	"io"
	"os"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/teng231/back4app/ledger"
)

func pieRoseRadius(holdings []*ledger.Holding) *charts.Pie {
	pie := charts.NewPie()
	pie.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{
			Title: "Danh sách phân bổ",
		}),
	)

	pie.AddSeries("By amount", makeItemsFromHoldingsByUsdt(holdings)).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show:      opts.Bool(true),
				Formatter: "{b}: {c}",
			}),
			charts.WithPieChartOpts(opts.PieChart{
				Radius: []string{"35%", "75%"},
			}),
		)

	return pie
}

// func makeItemsFromHoldingsByAmount(holdings []*ledger.Holding) []opts.PieData {
// 	items := make([]opts.PieData, 0)

// 	for _, holding := range holdings {
// 		items = append(items, opts.PieData{Name: holding.Symbol, Value: holding.Amount})
// 	}
// 	return items
// }

func makeItemsFromHoldingsByUsdt(holdings []*ledger.Holding) []opts.PieData {
	items := make([]opts.PieData, 0)

	for _, holding := range holdings {
		items = append(items, opts.PieData{Name: holding.Symbol, Value: holding.TVL})
	}
	return items
}
func MakePieDataChartToHTML(holdings []*ledger.Holding) {
	page := components.NewPage()

	page.AddCharts(
		pieRoseRadius(holdings),
	)
	f, err := os.Create("./pie_holding_data.html")
	if err != nil {
		panic(err)
	}
	page.Render(io.MultiWriter(f))
}

