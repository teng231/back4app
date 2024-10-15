package db

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/teng231/back4app/errhandler"
	"github.com/teng231/back4app/ledger"
	"gorm.io/gorm"
)

const (
	// tblClient    = "client"
	tblPortfolio = "portfolio"
	tblHolding   = "holding"
	tblTx        = "tx"
)

func (d *TiDB) InsertPortfolio(config *ledger.Portfolio) error {
	result := d.engine.Table(tblPortfolio).Create(config)
	if result.Error != nil {
		return result.Error
	}
	if int(result.RowsAffected) == 0 {
		return errors.New(errhandler.E_can_not_insert)
	}
	return nil
}

// UpdatePortfolio ...
func (d *TiDB) UpdatePortfolio(updator, selector *ledger.Portfolio) error {
	result := d.engine.Table(tblPortfolio).Where(selector).Updates(updator)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New(errhandler.E_can_not_update)
	}
	return nil
}

// FindPortfolio ...
func (d *TiDB) FindPortfolio(in *ledger.Portfolio) (*ledger.Portfolio, error) {
	out := &ledger.Portfolio{}
	err := d.engine.Table(tblPortfolio).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New(errhandler.E_not_found)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ###################### HOLDING ########################

func (d *TiDB) InsertHolding(config *ledger.Holding) error {
	result := d.engine.Table(tblHolding).Create(config)
	if result.Error != nil {
		return result.Error
	}
	if int(result.RowsAffected) == 0 {
		return errors.New(errhandler.E_can_not_insert)
	}
	return nil
}

// UpdateHolding ...
func (d *TiDB) UpdateHolding(updator, selector *ledger.Holding) error {
	result := d.engine.Table(tblHolding).Where(selector).Updates(updator)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New(errhandler.E_can_not_update)
	}
	return nil
}

// ResetHolding ...
func (d *TiDB) ResetHolding(selector *ledger.Holding) error {
	tx := d.engine.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	result := tx.Table(tblHolding).Where(selector).Updates(map[string]any{
		"updated": time.Now().Unix(),
		"Amount":  0.0,
		"tvl":     0.0,
	})
	if result.Error != nil {
		tx.Commit()
		return result.Error
	}
	if result.RowsAffected == 0 {
		tx.Commit()
		return errors.New(errhandler.E_can_not_update)
	}

	if err := tx.Table(tblTx).Where(&ledger.Tx{
		PortfolioID: selector.PortfolioID,
		Symbol:      selector.Symbol,
	}).Updates(&ledger.Tx{
		Status: 1,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := tx.Table(tblTx).Create(&ledger.Tx{
		PortfolioID: selector.PortfolioID,
		Symbol:      selector.Symbol,
		Created:     time.Now().Unix(),
		Action:      "RESET",
		Status:      2,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// FindHolding ...
func (d *TiDB) FindHolding(in *ledger.Holding) (*ledger.Holding, error) {
	out := &ledger.Holding{}
	err := d.engine.Table(tblHolding).Where(in).Take(out).Error
	if err == gorm.ErrRecordNotFound {
		return nil, errors.New(errhandler.E_not_found)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}

// listHoldingQuery list config
func (d *TiDB) listHoldingQuery(p *ledger.HoldingRequest) *gorm.DB {
	ss := d.engine.Table(tblHolding)
	if p.PortfolioID != 0 {
		ss.Where("portfolio_id = ?", p.PortfolioID)
	}
	if len(p.Symbols) > 0 {
		ss.Where("symbol in ?", p.Symbols)
	}
	if p.Symbol != "" {
		ss.Where("symbol = ?", p.Symbol)
	}
	if p.Status != 0 {
		ss.Where("status = ?", p.Status)
	}
	ss.Where("tvl > 0")
	return ss
}

func (d *TiDB) ListHoldings(rq *ledger.HoldingRequest) ([]*ledger.Holding, error) {
	var holdings []*ledger.Holding
	ss := d.listHoldingQuery(rq)
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	if rq.Order != "" {
		ss.Order(rq.Order)
	}
	if len(rq.Cols) > 0 {
		ss.Select(rq.Cols)
	}
	err := ss.Find(&holdings).Error
	if err != nil {
		return nil, err
	}
	return holdings, nil
}

func (d *TiDB) ListTxs(rq *ledger.TxRequest) ([]*ledger.Tx, error) {
	var holdings []*ledger.Tx
	ss := d.listTxQuery(rq)
	if rq.Limit != 0 {
		ss.Limit(rq.Limit)
	}
	if rq.Page > 1 {
		ss.Offset(rq.Limit * rq.Page)
	}
	err := ss.Order("id desc").Find(&holdings).Error
	if err != nil {
		return nil, err
	}
	return holdings, nil
}

func (d *TiDB) GetAvg(rq *ledger.TxRequest) (float32, error) {
	var avg float32
	err := d.engine.Table(tblTx).Select("AVG(income/amount)").
		Where("portfolio_id = ?", rq.PortfolioID).
		Where("symbol = ?", rq.Symbol).
		Where("status = ?", rq.Status).
		Where("action = ?", rq.Action).Take(&avg).Error
	return avg, err
}

// listTxQuery list config
func (d *TiDB) listTxQuery(p *ledger.TxRequest) *gorm.DB {
	ss := d.engine.Table(tblTx)
	if p.PortfolioID != 0 {
		ss.Where("portfolio_id = ?", p.PortfolioID)
	}
	if len(p.Symbols) > 0 {
		ss.Where("symbol in ?", p.Symbols)
	}
	if p.Symbol != "" {
		ss.Where("symbol = ?", p.Symbol)
	}
	if p.Status != 0 {
		ss.Where("status = ?", p.Status)
	}
	return ss
}
func (d *TiDB) TxHoldStableCoin(req *ledger.Tx) error {
	now := time.Now().Unix()
	tx := d.engine.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	var usd *ledger.Holding
	err := tx.Table(tblHolding).
		Where(&ledger.Holding{PortfolioID: req.PortfolioID, Symbol: "USDT"}).
		Take(&usd).Error
	if err != nil && err.Error() == gorm.ErrRecordNotFound.Error() {
		usd = &ledger.Holding{
			PortfolioID: req.PortfolioID,
			Symbol:      "USDT",
			Amount:      0,
			Status:      2,
			Created:     time.Now().Unix(),
		}

		if err = tx.Table(tblHolding).Create(usd).Error; err != nil {
			tx.Commit()
			return err
		}
	}
	if req.Action == "SELL" && usd.Amount < req.Amount {
		tx.Commit()
		return errors.New("insufficient")
	}

	amount := float64(0)
	if req.Action == "SELL" {
		amount = usd.Amount - req.Amount
	}
	if req.Action == "BUY" {
		amount = usd.Amount + req.Amount
	}

	err = tx.Table(tblHolding).
		Where("id", usd.ID).
		Updates(map[string]any{
			"updated": now,
			"amount":  amount,
			"tvl":     amount,
		}).Error
	if err != nil {
		tx.Commit()
		return err
	}

	req.AmountHoldingBefore = usd.Amount
	req.AmountHoldingAfter = amount
	req.Status = 2

	if err = tx.Table(tblTx).Create(req).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

// ################## TX #######################
// TxHoldByTransation
func (d *TiDB) TxHoldByTransation(req *ledger.Tx) error {

	if req.Symbol == "USDT" {
		return d.TxHoldStableCoin(req)
	}

	tx := d.engine.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	if err := tx.Error; err != nil {
		return err
	}
	now := time.Now().Unix()
	// check hold
	// insert tx
	var holding *ledger.Holding
	err := tx.Table(tblHolding).
		Where(&ledger.Holding{PortfolioID: req.PortfolioID, Symbol: req.Symbol}).
		Take(&holding).Error

	if err != nil && err.Error() == gorm.ErrRecordNotFound.Error() {
		holding = &ledger.Holding{
			PortfolioID: req.PortfolioID,
			Symbol:      req.Symbol,
			Status:      2,
			Created:     now,
		}
		if err := tx.Table(tblHolding).Create(holding).Error; err != nil {
			tx.Commit()
			return err
		}
	}
	if holding.Status != 2 {
		tx.Commit()
		return errors.New("holding not active")
	}

	var usd *ledger.Holding
	err = tx.Table(tblHolding).
		Where(&ledger.Holding{PortfolioID: req.PortfolioID, Symbol: "USDT"}).
		Take(&usd).Error
	if err != nil && err.Error() == gorm.ErrRecordNotFound.Error() {
		usd = &ledger.Holding{
			PortfolioID: holding.ID,
			Symbol:      "USDT",
			Amount:      0,
			Status:      2,
			Created:     time.Now().Unix(),
		}
		if err = tx.Table(tblHolding).Create(usd).Error; err != nil {
			tx.Commit()
			return err
		}
	}

	if req.Action == "SELL" && holding.Amount < req.Amount {
		tx.Commit()
		return errors.New("insufficient " + holding.Symbol)
	}
	if req.Action == "BUY" && usd.Amount < req.Income {
		tx.Commit()

		return errors.New("insufficient " + usd.Symbol)
	}

	amount := float64(0)
	updateData := map[string]any{
		"updated": now,
		// "amount":  amount,
		// "tvl":     tvlIncome,
	}
	if req.Action == "SELL" {
		usd.Amount += req.Income
		updateData["amount"] = holding.Amount - req.Amount
		if holding.TVL-req.Income < 0 {
			updateData["tvl"] = 0
		} else {
			updateData["tvl"] = gorm.Expr("tvl - ?", req.Income)
		}
	}
	if req.Action == "BUY" {
		usd.Amount -= req.Income
		updateData["amount"] = holding.Amount + req.Amount
		updateData["tvl"] = gorm.Expr("tvl + ?", req.Income)
	}

	req.AmountHoldingBefore = holding.Amount
	req.AmountHoldingAfter = amount
	req.Status = 2
	if err = tx.Table(tblTx).Create(req).Error; err != nil {
		tx.Rollback()
		return err
	}

	if req.Action == "BUY" {
		var avg float32
		err := tx.Table(tblTx).Select("AVG(income/amount)").
			Where("portfolio_id = ?", holding.PortfolioID).
			Where("symbol = ?", strings.ToUpper(holding.Symbol)).
			Where("status = ?", 2).
			Where("action = ?", "BUY").Take(&avg).Error
		if err != nil {
			log.Print(err)
		}
		if avg != 0 {
			updateData["avg"] = avg
		}
	}
	err = tx.Table(tblHolding).
		Where("id", holding.ID).
		Updates(updateData).Error
	if err != nil {
		tx.Commit()
		return err
	}

	err = tx.Table(tblHolding).
		Where("id", usd.ID).
		Updates(map[string]any{
			"updated": now,
			"amount":  usd.Amount,
			"tvl":     usd.Amount,
		}).Error
	if err != nil {
		tx.Commit()
		return err
	}

	tx.Commit()
	return nil
}
