package model

import (
	"time"

	"gorm.io/gorm"
)

// BillRecord 账单记录表
type BillRecord struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint           `gorm:"index:user_id_no_deleted_at;not null;comment:用户ID" json:"user_id"`
	TradeNo         string         `gorm:"size:255;comment:交易单号" json:"trade_no"`
	MerchantOrderNo string         `gorm:"size:255;comment:商户单号" json:"merchant_order_no"`
	Platform        uint8          `gorm:"comment:平台（支付宝、微信）" json:"platform"`
	IncomeType      uint8          `gorm:"comment:收支类型（收入、支出、不记收支）" json:"income_type"`
	TradeType       string         `gorm:"size:255;comment:交易类型（分类）" json:"trade_type"`
	ProductName     string         `gorm:"size:255;comment:商品（交易名称）" json:"product_name"`
	Counterparty    string         `gorm:"size:255;comment:交易对方（商户名称）" json:"counterparty"`
	PaymentMethod   string         `gorm:"size:255;comment:交易方式（余额、银行卡）" json:"payment_method"`
	Amount          float64        `gorm:"type:decimal(10,2);comment:金额" json:"amount"`
	TradeStatus     string         `gorm:"size:255;comment:交易状态（成功、失败、关闭、退款等）" json:"trade_status"`
	TradeTime       int64          `gorm:"not null;comment:交易时间" json:"trade_time"`
	Remark          string         `gorm:"size:255;comment:备注" json:"remark"`
	CreatedAt       time.Time      `json:"created_at"`              // 创建时间，自动填充
	UpdatedAt       time.Time      `json:"updated_at"`              // 更新时间，自动更新
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at"` // 逻辑删除
}

// Platform 平台枚举
type Platform uint8

const (
	PlatformWechat Platform = 1 // 微信
	PlatformAlipay Platform = 2 // 支付宝
)

// IncomeType 收支类型枚举
type IncomeType uint8

const (
	IncomeTypeIncome  IncomeType = 1 // 收入
	IncomeTypeExpense IncomeType = 2 // 支出
	IncomeTypeNone    IncomeType = 3 // 不记收支
	IncomeTypeUnknown IncomeType = 4 // 未知
)

// 将字符串收入类型转换为 uint8 枚举
func IncomeTypeFromString(s string) uint8 {
	switch s {
	case "收入":
		return uint8(IncomeTypeIncome)
	case "支出":
		return uint8(IncomeTypeExpense)
	case "不计收支":
		return uint8(IncomeTypeNone)
	default:
		return uint8(IncomeTypeUnknown)
	}
}
