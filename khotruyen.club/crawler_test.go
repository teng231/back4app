package khotruyenclub

import (
	"log"
	"testing"

	"github.com/golang-module/carbon/v2"
)

func TestCrawler(t *testing.T) {
	Crawl()
}

func TestTime(t *testing.T) {
	created := carbon.ParseByFormat("22/02/2024", "d/m/Y", carbon.Saigon).EndOfDay().Timestamp()
	log.Print(created)
}
