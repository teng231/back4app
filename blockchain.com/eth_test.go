package blockchaincom

import (
	"log"
	"testing"
)

func TestETH(t *testing.T) {

	ETH("wss://ws.blockchain.info/coins")
}

func TestETHValue(t *testing.T) {
	log.Print(0x2e90edd000 / 1e18)
}
