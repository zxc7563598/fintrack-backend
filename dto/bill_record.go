package dto

type BillListItem struct {
	ID            uint    `json:"id"`
	TradeTime     int64   `json:"trade_time"`
	TradeType     string  `json:"trade_type"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	ProductName   string  `json:"product_name"`
	IncomeType    uint8   `json:"income_type"`
	Remark        string  `json:"remark"`
}

type BillInfoItem struct {
	ID              uint    `json:"id"`
	UserID          uint    `json:"user_id"`
	TradeNo         string  `json:"trade_no"`
	MerchantOrderNo string  `json:"merchant_order_no"`
	Platform        uint8   `json:"platform"`
	IncomeType      uint8   `json:"income_type"`
	TradeType       string  `json:"trade_type"`
	ProductName     string  `json:"product_name"`
	Counterparty    string  `json:"counterparty"`
	PaymentMethod   string  `json:"payment_method"`
	Amount          float64 `json:"amount"`
	TradeStatus     string  `json:"trade_status"`
	TradeTime       int64   `json:"trade_time"`
	Remark          string  `json:"remark"`
}

type BillDailySummary struct {
	Date    string  `json:"date"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
}
