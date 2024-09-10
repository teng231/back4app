package db

import (
	"context"
	"log"
	"time"

	"github.com/teng231/back4app/ledger"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TiDB struct {
	engine *gorm.DB
}

type ITiDB interface {
	StatusCheck() error
	Migrate() error
	InsertPortfolio(config *ledger.Portfolio) error
	UpdatePortfolio(updator, selector *ledger.Portfolio) error
	FindPortfolio(in *ledger.Portfolio) (*ledger.Portfolio, error)
	InsertHolding(config *ledger.Holding) error
	UpdateHolding(updator, selector *ledger.Holding) error
	ResetHolding(selector *ledger.Holding) error
	FindHolding(in *ledger.Holding) (*ledger.Holding, error)
	ListHoldings(rq *ledger.HoldingRequest) ([]*ledger.Holding, error)
	ListTxs(rq *ledger.TxRequest) ([]*ledger.Tx, error)

	TxHoldByTransation(req *ledger.Tx) error
	TxHoldStableCoin(req *ledger.Tx) error
}

func (d *TiDB) StatusCheck() error {
	conn, err := d.engine.DB()

	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.PingContext(ctx)
}

func (d *TiDB) Migrate() error {
	if err := d.engine.Table(tblPortfolio).AutoMigrate(ledger.Portfolio{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblHolding).AutoMigrate(ledger.Holding{}); err != nil {
		return err
	}
	if err := d.engine.Table(tblTx).AutoMigrate(ledger.Tx{}); err != nil {
		return err
	}
	return nil
}

func New(dsn string) (*TiDB, error) {
	db, err := gorm.Open(mysql.New(
		mysql.Config{DSN: dsn}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		/*i
		GORM perform write (create/update/delete) operations run inside a transaction to ensure data consistency, you can disable it during initialization if it is not required, you will gain about 30%+ performance improvement after that
		*/
		SkipDefaultTransaction: true,
		// PrepareStmt:            true,
	})
	if err != nil {
		log.Print("connect db fail ", err)
		return nil, err
	}

	return &TiDB{engine: db}, nil
}
