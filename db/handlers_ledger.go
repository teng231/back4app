package db

import (
	"errors"
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

	if err := tx.Table(tblTx).Create(&ledger.Tx{
		PortfolioID: selector.PortfolioID,
		Symbol:      selector.Symbol,
		Created:     time.Now().Unix(),
		Action:      "RESET",
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
	return ss
}

// ################## TX #######################
// TxHoldByTransation
func (d *TiDB) TxHoldByTransation(req *ledger.Tx) error {
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

	if req.Action == "SELL" && holding.Amount < req.Amount {
		tx.Commit()
		return errors.New("insufficient")
	}

	amount := float64(0)
	if req.Action == "SELL" {
		amount = holding.Amount - req.Amount
	}
	if req.Action == "BUY" {
		amount = holding.Amount + req.Amount
	}

	err = tx.Table(tblHolding).
		Where("id", holding.ID).
		Updates(map[string]any{
			"updated": now,
			"amount":  amount,
			"tvl":     gorm.Expr("tvl + ?", req.Income),
		}).Error
	if err != nil {
		tx.Commit()
		return err
	}

	req.AmountHoldingBefore = holding.Amount
	req.AmountHoldingAfter = amount

	if err = tx.Table(tblTx).Create(req).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}
