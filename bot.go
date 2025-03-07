package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/teng231/back4app/chart"
	"github.com/teng231/back4app/cryptodata"
	"github.com/teng231/back4app/db"
	"github.com/teng231/back4app/errhandler"
	"github.com/teng231/back4app/ledger"
	"github.com/teng231/back4app/telebot"
	"github.com/teng231/back4app/utils"
	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/react"
)

type Bot struct {
	*telebot.Bot
	db                   db.ITiDB
	cryptoData           map[string]*cryptodata.TokenDetail
	timeUpdateCryptoData int64
}

func trimFloat(number float64) string {
	if number < 0.01 {
		return fmt.Sprintf("%.4f", number)
	}
	if number < 1 {
		return fmt.Sprintf("%.3f", number)
	}
	if number < 10 {
		return fmt.Sprintf("%.2f", number)
	}
	if number > 100 {
		return fmt.Sprintf("%.0f", number)
	}
	return fmt.Sprintf("%.0f", number)
}

func shortHolding(holdings []*ledger.Holding, cryptoDatas map[string]*cryptodata.TokenDetail) []map[string]any {
	out := make([]map[string]any, 0)
	tvlAll := 0.0
	for _, holding := range holdings {
		if holding.TVL < 0 {
			continue
		}
		tvlAll += holding.TVL
	}
	for _, val := range holdings {
		amount := fmt.Sprintf("%.1f", val.Amount)
		if val.Amount < 1 {
			amount = fmt.Sprintf("%.3f", val.Amount)
		}
		// avg := val.TVL / val.Amount
		avgStr := fmt.Sprintf("%.1f", val.AVG)
		if val.AVG < 1 {
			avgStr = fmt.Sprintf("%.3f", val.AVG)
		}
		pc := 0.0
		coin := cryptoDatas[val.Symbol]
		if coin != nil {
			pc = val.Amount * (coin.Price - float64(val.AVG))
		}
		out = append(out, map[string]any{
			"sym": val.Symbol,
			"amt": amount,
			"tvl": fmt.Sprintf("%.1f", val.TVL),
			"avg": avgStr,
			"+/-": int64(pc),
			// "%":   perc,
		})
	}
	return out
}

func shortHoldingDetail(holdings []*ledger.Holding, cryptoDatas map[string]*cryptodata.TokenDetail) []map[string]any {
	out := make([]map[string]any, 0)
	tvlAll := 0.0
	for _, holding := range holdings {
		if holding.TVL < 0 {
			continue
		}
		tvlAll += holding.TVL
	}
	for _, val := range holdings {
		pc := 0.0
		coin := cryptoDatas[val.Symbol]
		if coin != nil {
			pc = val.Amount * (coin.Price - float64(val.AVG))
		}

		price := ""

		if coin.Price < 1 {
			price = fmt.Sprintf("%.3f", coin.Price)
		} else if coin.Price < 100 {
			price = fmt.Sprintf("%.1f", coin.Price)
		} else {
			price = fmt.Sprintf("%.0f", coin.Price)
		}

		avgStr := fmt.Sprintf("%.1f", val.AVG)
		if val.AVG < 1 {
			avgStr = fmt.Sprintf("%.3f", val.AVG)
		}
		out = append(out, map[string]any{
			"sym":    val.Symbol,
			"+%-":    fmt.Sprintf("%.1f", pc*100/val.TVL),
			"price":  price,
			"pr_avg": avgStr,
		})
	}
	return out
}
func shortTx(txs []*ledger.Tx) []map[string]any {
	out := make([]map[string]any, 0)
	for _, tx := range txs {
		avg := tx.Income / tx.Amount
		avgStr := ""
		if avg > 1 && avg <= 100 {
			avgStr = fmt.Sprintf("%.1f", avg)
		} else if avg <= 1 {
			avgStr = fmt.Sprintf("%.2f", avg)
		} else {
			avgStr = fmt.Sprintf("%.0f", avg)
		}
		out = append(out, map[string]any{
			"sym":  tx.Symbol,
			"amt":  fmt.Sprintf("%.3f", tx.Amount),
			"inco": fmt.Sprintf("%.1f", tx.Income),
			"avg":  avgStr,
			"act":  tx.Action,
		})
	}

	return out
}

func sendError(ctx tele.Context, txt1, txt2 string) error {
	return ctx.Send(
		fmt.Sprintf(`%s %s:
		%s%s%s`, react.ManShrugging.Emoji, txt1, "`", txt2, "`"), &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
}
func newBot(botToken string, db *db.TiDB) *Bot {
	b := telebot.Start(botToken)
	b.PrivateHandlers()
	bot := &Bot{Bot: b, db: db, cryptoData: cryptodata.ListCrytos(), timeUpdateCryptoData: time.Now().Unix()}
	return bot
}

func (b *Bot) registerHandlers() *Bot {
	b.Handle("/ping", func(ctx tele.Context) error {
		commands := []tele.Command{
			{Text: "sell", Description: "Bán <symbol> <amount> <price>"},
			{Text: "buy", Description: "Mua <symbol> <amount> <price>"},
			{Text: "del", Description: "Xoá <symbol>"},
			{Text: "tx", Description: "List Tx <?symbol>"},
			{Text: "portfolio", Description: "List đang holdings"},
			{Text: "suggestion", Description: "Hệ thống cảnh báo và gợi ý"},
		}
		// Set commands
		if err := b.SetCommands(commands); err != nil {
			log.Fatal(err)
		}
		return ctx.Send(react.Brain.Emoji + "I'm here!")
	})
	b.Handle("/start", func(ctx tele.Context) error {
		clientId := strconv.Itoa(int(ctx.Sender().ID))
		_, err := b.db.FindPortfolio(&ledger.Portfolio{ClientID: clientId})
		if err != nil && err.Error() == errhandler.E_not_found {
			// insert
			port := &ledger.Portfolio{
				ClientID: clientId,
				Channel:  1,
				StartAt:  time.Now().Unix(),
			}
			err := b.db.InsertPortfolio(port)
			if err != nil {
				ctx.Send("1. -,- " + ctx.Sender().Username + " Fail portfolio !")
			}
			if err == nil {
				ctx.Send("1. ^.^ " + ctx.Sender().Username + " Insert portfolio done !")
			}
			err = b.db.InsertHolding(&ledger.Holding{
				PortfolioID: port.ID,
				Symbol:      "USDT",
				Amount:      0,
				Status:      2,
				Created:     time.Now().Unix(),
			})
			if err != nil {
				ctx.Send("2. -,- " + ctx.Sender().Username + " Fail stable coin !")
			}
			if err == nil {
				ctx.Send("2. ^.^ " + ctx.Sender().Username + " Insert holding stable coin done !")
			}

		}
		return ctx.Send(react.Brain.Emoji + "hi @" + ctx.Sender().Username)
	})
	b.Handle("/sell", func(ctx tele.Context) error {
		// example: /sell <symbol> <amount> <price>
		args := ctx.Args()
		if len(args) < 3 {
			return sendError(ctx, "Sai cú pháp:  ", "/sell <symbol> <amount> <price>")
		}

		sym := strings.Trim(args[0], " ")
		if sym == "" {
			return ctx.Send(react.ManShrugging.Emoji + "Sai symbol:  " + "/sell <symbol> <amount> <price>")
		}
		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Sai amount:  " + "/sell <symbol> <amount> <price>")
		}
		sellPrice, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Sai sellprice:  " + "/sell <symbol> <amount> <price>")
		}

		por, err := b.db.FindPortfolio(&ledger.Portfolio{ClientID: strconv.Itoa(int(ctx.Sender().ID))})
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Không tìm thấy thông tin")
		}
		led := &ledger.Tx{
			PortfolioID: por.ID,
			Symbol:      strings.ToUpper(sym),
			Action:      "SELL",
			Amount:      amount,
			Income:      sellPrice * amount,
			Created:     time.Now().Unix(),
		}
		err = b.db.TxHoldByTransation(led)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + err.Error())
		}

		ctx.Send("Done")
		return ctx.Send("```\n"+utils.PrintTable(shortTx([]*ledger.Tx{led}))+"```", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	})

	b.Handle("/buy", func(ctx tele.Context) error {
		// example: /buy <symbol> <amount> <price>
		args := ctx.Args()
		if len(args) < 3 {
			return ctx.Send(react.ManShrugging.Emoji + "Sai cú pháp:  " + "/buy <symbol> <amount> <price>")
		}

		sym := strings.Trim(args[0], " ")
		if sym == "" {
			return ctx.Send(react.ManShrugging.Emoji + "Sai symbol:  " + "/buy <symbol> <amount> <price>")
		}
		amount, err := strconv.ParseFloat(args[1], 64)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Sai amount:  " + "/buy <symbol> <amount> <price>")
		}
		buyPrice, err := strconv.ParseFloat(args[2], 64)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Sai buyprice:  " + "/buy <symbol> <amount> <price>")
		}
		por, err := b.db.FindPortfolio(&ledger.Portfolio{ClientID: strconv.Itoa(int(ctx.Sender().ID))})
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Không tìm thấy thông tin")
		}
		led := &ledger.Tx{
			PortfolioID: por.ID,
			Symbol:      strings.ToUpper(sym),
			Action:      "BUY",
			Amount:      amount,
			Income:      buyPrice * amount,
			Created:     time.Now().Unix(),
		}
		err = b.db.TxHoldByTransation(led)
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + err.Error())
		}

		ctx.Send("Done")
		return ctx.Send("```\n"+utils.PrintTable(shortTx([]*ledger.Tx{led}))+"```", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})
	})
	b.Handle("/del", func(ctx tele.Context) error {
		// example: /del <symbol>
		args := ctx.Args()
		if len(args) < 1 {
			return ctx.Send(react.ManShrugging.Emoji + "Sai cú pháp:  " + "/del <symbol>")
		}

		sym := strings.Trim(args[0], " ")
		if sym == "" {
			return ctx.Send(react.ManShrugging.Emoji + "Sai symbol:  " + "/del <symbol>")
		}
		// update status hoding and remove amount
		por, err := b.db.FindPortfolio(&ledger.Portfolio{ClientID: strconv.Itoa(int(ctx.Sender().ID))})
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Không tìm thấy thông tin")
		}

		if err := b.db.ResetHolding(&ledger.Holding{Symbol: strings.ToUpper(sym), PortfolioID: por.ID}); err != nil {
			return ctx.Send(react.ManShrugging.Emoji + err.Error())
		}
		return ctx.Send("Done!")
	})

	b.Handle("/tx", func(ctx tele.Context) error {
		// example: /tx <?symbol>
		args := ctx.Args()
		sym := ""
		if len(args) > 0 {
			sym = strings.Trim(args[0], " ")
		}
		// update status hoding and remove amount
		por, err := b.db.FindPortfolio(&ledger.Portfolio{ClientID: strconv.Itoa(int(ctx.Sender().ID))})
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Không tìm thấy thông tin")
		}

		txs, err := b.db.ListTxs(&ledger.TxRequest{PortfolioID: por.ID, Symbol: strings.ToUpper(sym), Status: 2})
		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + err.Error())
		}
		ctx.Send("Danh sách giao dịch")
		return ctx.Send("```\n"+utils.PrintTable(shortTx(txs))+"```", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})

	})

	b.Handle("/portfolio", func(ctx tele.Context) error {

		// check value outdate
		//
		if b.timeUpdateCryptoData+60 < time.Now().Unix() {
			b.cryptoData = cryptodata.ListCrytos()
			b.timeUpdateCryptoData = time.Now().Unix()
		}
		por, _ := b.db.FindPortfolio(&ledger.Portfolio{ClientID: strconv.Itoa(int(ctx.Sender().ID))})
		holdings, err := b.db.ListHoldings(&ledger.HoldingRequest{
			PortfolioID: por.ID,
			Limit:       25,
			Order:       "TVL DESC",
			Cols:        []string{"symbol", "amount", "tvl", "avg"},
			Status:      2,
		})

		if err != nil {
			return ctx.Send(react.ManShrugging.Emoji + "Không tìm thấy thông tin")
		}
		now := time.Now().Unix()
		filename := fmt.Sprintf("./%s_%d.png", ctx.Sender().Username, now)
		chart.MakePieDataChartToImg(filename, holdings)
		defer os.Remove(filename)

		file, err := os.Open(filename)
		if err == nil {
			// Tạo đối tượng Photo
			photo := &tele.Photo{
				File: tele.FromReader(file),
			}
			// Gửi ảnh
			if err := ctx.Send(photo); err != nil {
				log.Print(err)
			}
			file.Close()
		} else {
			log.Print(err)
		}
		ctx.Send("```\n"+utils.PrintTable(shortHolding(holdings, b.cryptoData))+"```", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})

		ctx.Send("```\n"+utils.PrintTable(shortHoldingDetail(holdings, b.cryptoData))+"```", &tele.SendOptions{ParseMode: tele.ModeMarkdownV2})

		return nil
	})
	return b
}
