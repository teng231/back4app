package blockchaincom

import (
	"log"

	"github.com/teng231/back4app/wsclient"
)

type CoinTx struct {
	Coin        string `json:"coin"`
	Entity      string `json:"entity"`
	Transaction struct {
		Hash             string `json:"hash"`
		BlockHash        string `json:"blockHash"`
		BlockNumber      string `json:"blockNumber"`
		From             string `json:"from"`
		To               string `json:"to"`
		ContractAddress  string `json:"contractAddress"`
		Value            string `json:"value"`
		Nonce            int    `json:"nonce"`
		GasPrice         string `json:"gasPrice"`
		GasLimit         int    `json:"gasLimit"`
		GasUsed          int    `json:"gasUsed"`
		Data             string `json:"data"`
		TransactionIndex int    `json:"transactionIndex"`
		Success          bool   `json:"success"`
		Error            string `json:"error"`
		FirstSeen        int    `json:"firstSeen"`
		Timestamp        int    `json:"timestamp"`
		State            string `json:"state"`
	} `json:"transaction"`
}

func ETH(wsurl string) error {
	reqHeader := map[string]string{}
	// connect ws
	conn, err := wsclient.NewConn(wsurl, reqHeader)
	if err != nil {
		log.Print("connect fail: ", err)
		return err
	}
	log.Print("connected ws ETH")

	defer conn.Close()

	err = conn.Write(map[string]any{
		// "op":      "unconfirmed_sub",
		"coin":    "eth",
		"command": "subscribe",
		"entity":  "pending_transaction",
	})

	log.Print("write done ws ETH ", err)

	if err != nil {
		log.Print("write msg fail: ", err)
		return err
	}
	eventBufs := make(chan []byte, 1000)
	go func() {
		for {
			events := <-eventBufs
			log.Print(string(events))
		}
	}()
	conn.Read(eventBufs)
	return nil
}
