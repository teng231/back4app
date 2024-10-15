package ledger

// type Client struct {
// 	ID      int64  `json:"id,omitempty"`
// 	UserId  string `json:"user_id,omitempty"`
// 	Created int64  `json:"created,omitempty"`
// 	Channel int32  `json:"channel,omitempty"` // 1: telegram
// }

type Portfolio struct {
	ID           int64   `json:"id,omitempty" gorm:"primaryKey;not null;autoIncrement"`
	ClientID     string  `json:"client_id,omitempty"  gorm:"index"`
	Channel      int32   `json:"channel,omitempty"` // 1: telegram
	StartAt      int64   `json:"start_at,omitempty"`
	Exchanges    string  `json:"exchanges,omitempty"` // binance | bybit
	TVL          float64 `json:"tvl,omitempty"`       // tổng tiền đầu tư
	ClientIdHash string  `json:"hash,omitempty"`
}

type HoldingRequest struct {
	ID          int64    `json:"id,omitempty"`
	PortfolioID int64    `json:"portfolio_id,omitempty"`
	Symbol      string   `json:"symbol,omitempty"`  // ETH, BTC...
	Symbols     []string `json:"symbols,omitempty"` // ETH, BTC...
	Status      int      `json:"status,omitempty"`  // 2 active | 3 inactive or del
	Limit       int      `json:"limit,omitempty"`
	Page        int      `json:"page,omitempty"`
	Cols        []string `json:"cols,omitempty"`
	Order       string   `json:"order,omitempty"`
}

type Holding struct {
	ID          int64   `json:"id,omitempty" gorm:"primaryKey;not null;autoIncrement"`
	PortfolioID int64   `json:"portfolio_id,omitempty"  gorm:"index"`
	Symbol      string  `json:"symbol,omitempty" gorm:"index"` // ETH, BTC...
	Amount      float64 `json:"amount,omitempty"`
	Updated     int64   `json:"updated,omitempty"`
	Created     int64   `json:"created,omitempty"`
	Status      int     `json:"status,omitempty"` // 2 active | -1 inactive or del
	TVL         float64 `json:"tvl,omitempty"`    // tổng vào tiền
	AVG         float32 `json:"avg,omitempty"`
}

type TxRequest struct {
	ID          int64    `json:"id,omitempty"`
	PortfolioID int64    `json:"portfolio_id,omitempty"`
	Symbol      string   `json:"symbol,omitempty"`  // ETH, BTC...
	Symbols     []string `json:"symbols,omitempty"` // ETH, BTC...
	Action      string   `json:"action,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	Page        int      `json:"page,omitempty"`
	Income      float64  `json:"income,omitempty"`
	Status      int      `json:"status,omitempty"` // 2 active 1 del

}

type Tx struct {
	ID                  int64   `json:"id,omitempty" gorm:"primaryKey;not null;autoIncrement"`
	PortfolioID         int64   `json:"portfolio_id,omitempty"  gorm:"index"`
	Symbol              string  `json:"symbol,omitempty"  gorm:"index"`
	Amount              float64 `json:"amount,omitempty"`
	Action              string  `json:"action,omitempty"` // 1: buy, 2: sell, 3: swap
	Created             int64   `json:"created,omitempty"`
	AmountHoldingBefore float64 `json:"amount_holding_before,omitempty"`
	AmountHoldingAfter  float64 `json:"amount_holding_after,omitempty"`
	Income              float64 `json:"income,omitempty"`
	Description         string  `json:"description,omitempty"`
	Status              int     `json:"status,omitempty"` // 2 active 1 del
}
