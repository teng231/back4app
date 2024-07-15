package utils

import (
	"fmt"
	"testing"
)

func TestPrint2(t *testing.T) {
	// Ví dụ với slice của map[string]interface{}
	data := []map[string]interface{}{
		{"amount": 33436465, "tvl": 234.95108400000004, "expired_time": "ENA"},
	}

	fmt.Println(PrintTable(data))

	// Ví dụ với slice của struct
	type Product struct {
		ID          int `json:"id,omitempty"`
		Codex       string
		ProductID   float32
		ExpiredTime int
	}

	products := []*Product{
		{ID: 33436465, Codex: "uSuJnZuhMzLqKIK9dCaUg_LAudm2y6dNon9d7NZRDqs", ProductID: 5429, ExpiredTime: 1719853199},
		{ID: 33436466, Codex: "sWnJnZu_IZlLrQW9uNaUg_LAudm2y6dNon9d7NDRZqs", ProductID: 5430, ExpiredTime: 1719853299},
	}

	// Chuyển struct slice thành map slice
	mapSlice := StructSliceToMapSlice(products)
	fmt.Println(PrintTable(mapSlice))
}
