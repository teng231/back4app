package chart

import (
	"fmt"
	"log"
	"os"

	"github.com/wcharczuk/go-chart/v2"

	"github.com/teng231/back4app/ledger"
)

func MakePieDataChartToImg(filePath string, holdings []*ledger.Holding) {
	pie := chart.PieChart{
		Width:  650,
		Height: 750,
	}
	pie.DPI = 100
	valItems := make([]chart.Value, 0)

	tvlAll := 0.0
	for _, holding := range holdings {
		tvlAll += holding.TVL
	}
	pie.Title = "@danh mục " + fmt.Sprintf("%.1f%s", tvlAll, "$")
	for _, holding := range holdings {
		valItems = append(valItems, chart.Value{
			Value: holding.TVL,
			Label: fmt.Sprintf("%s(%.1f$) - %.2f%s", holding.Symbol, holding.TVL, holding.TVL*100/tvlAll, "%"),
			Style: chart.Style{FontSize: 7.5},
		})
	}

	pie.Values = valItems
	f, err := os.Create(filePath)
	if err != nil {
		log.Print(err)
	}
	defer f.Close()
	pie.Render(chart.PNG, f)
}
