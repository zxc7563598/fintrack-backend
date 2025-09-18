package controller

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/dto"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/response"
)

// 获取交易列表请求体
type GetBillListRequest struct {
	StartFormattedDate *string `json:"start_formatted_date"` // 开始日期
	EndFormattedDate   *string `json:"end_formatted_date"`   // 结束日期
	Search             *string `json:"search"`               // 搜索内容
	IncomeType         *int    `json:"income_type"`          // 收支类型
	Page               *int    `json:"page"`                 // 页码
	ItemsPerPage       *int    `json:"items_per_page"`       // 每页条数
	SortKey            *string `json:"sort_key"`             // 排序字段
	SortOrder          *string `json:"sort_order"`           // 排序顺序
}

// 获取交易列表接口
func GetBillListHandler(c *gin.Context) {
	layout := "2006-01-02"
	// 获取用户ID
	userIDAny, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, 300001)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		response.Fail(c, 300002)
		return
	}
	// 获取请求参数
	req, ok := c.MustGet("payload").(GetBillListRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取账单数据
	var records []dto.BillListItem
	var total int64
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp := t.Unix() // 当天 00:00:00 的秒级时间戳
			db = db.Where("trade_time >= ?", startTimestamp)
		}
	}
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			// 加上 23:59:59
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp := endOfDay.Unix()
			db = db.Where("trade_time <= ?", endTimestamp)
		}
	}
	if req.Search != nil && *req.Search != "" {
		like := fmt.Sprintf("%%%s%%", *req.Search)
		db = db.Where(
			config.DB.Where("product_name LIKE ?", like).
				Or("counterparty LIKE ?", like).
				Or("remark LIKE ?", like).
				Or("trade_no LIKE ?", like).
				Or("merchant_order_no LIKE ?", like),
		)
	}
	if req.IncomeType != nil {
		db = db.Where("income_type = ?", *req.IncomeType)
	}
	// 计算总数
	if err := db.Count(&total).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 排序
	sortKey := "trade_time"
	if req.SortKey != nil && *req.SortKey != "" {
		sortKey = *req.SortKey
	}
	sortOrder := "desc"
	if req.SortOrder != nil && (*req.SortOrder == "asc" || *req.SortOrder == "desc") {
		sortOrder = *req.SortOrder
	}
	db = db.Order(fmt.Sprintf("%s %s", sortKey, sortOrder))
	// 分页
	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	itemsPerPage := 20
	if req.ItemsPerPage != nil && *req.ItemsPerPage > 0 {
		itemsPerPage = *req.ItemsPerPage
	}
	offset := (page - 1) * itemsPerPage
	db = db.Offset(offset).Limit(itemsPerPage)
	// 查询
	if err := db.Find(&records).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"total": total,
		"data":  records,
	})
}

// 获取交易信息请求体
type GetBillInfoRequest struct {
	ID uint `json:"id" binding:"required"` // ID，修改透传，添加为0
}

// 获取交易信息接口
func GetBillInfoHandler(c *gin.Context) {
	// 获取用户ID
	userIDAny, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, 300001)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		response.Fail(c, 300002)
		return
	}
	// 获取请求参数
	req, ok := c.MustGet("payload").(GetBillInfoRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	var tradeTypes []string
	if err := config.DB.Model(&model.BillRecord{}).
		Distinct("trade_type").
		Where("user_id = ?", userID).
		Pluck("trade_type", &tradeTypes).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	var counterpartys []string
	if err := config.DB.Model(&model.BillRecord{}).
		Distinct("counterparty").
		Where("user_id = ?", userID).
		Pluck("counterparty", &counterpartys).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	var paymentMethod []string
	if err := config.DB.Model(&model.BillRecord{}).
		Distinct("payment_method").
		Where("user_id = ?", userID).
		Pluck("payment_method", &paymentMethod).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 获取数据
	var bill dto.BillInfoItem
	if req.ID > 0 {
		result := config.DB.Model(&model.BillRecord{}).Where("id = ? and user_id = ?", req.ID, userID).First(&bill)
		if result.Error != nil {
			response.Fail(c, 100001)
			return
		}
	}
	// 返回数据
	response.Ok(c, gin.H{
		"trade_types":    tradeTypes,
		"counterpartys":  counterpartys,
		"payment_method": paymentMethod,
		"data":           bill,
	})
}

// 存储交易信息请求体
type StoreBillRecordRequest struct {
	ID            uint    `json:"id" binding:"required"`             // ID，修改透传，添加为0
	Platform      uint8   `json:"platform" binding:"required"`       // 交易平台
	IncomeType    uint8   `json:"income_type" binding:"required"`    // 收支类型
	TradeType     string  `json:"trade_type" binding:"required"`     // 交易类型
	ProductName   string  `json:"product_name" binding:"required"`   // 交易名称
	Counterparty  string  `json:"counterparty" binding:"required"`   // 商户名称
	PaymentMethod string  `json:"payment_method" binding:"required"` // 支付方式
	Amount        float64 `json:"amount" binding:"required"`         // 金额
	TradeTime     string  `json:"trade_time" binding:"required"`     // 交易时间
	Remark        string  `json:"remark"`                            // 备注
}

// 存储交易信息接口
func StoreBillRecordHandler(c *gin.Context) {
	layout := "2006-01-02 15:04:05"
	// 获取用户ID
	userIDAny, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, 300001)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		response.Fail(c, 300002)
		return
	}
	// 获取请求参数
	req, ok := c.MustGet("payload").(StoreBillRecordRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	t, err := time.ParseInLocation(layout, req.TradeTime, time.Local)
	if err != nil {
		fmt.Println("解析时间出错:", err)
		response.Fail(c, 100012)
		return
	}
	bill := model.BillRecord{
		UserID:        userID,
		Platform:      req.Platform,
		IncomeType:    req.IncomeType,
		TradeType:     req.TradeType,
		ProductName:   req.ProductName,
		Counterparty:  req.Counterparty,
		PaymentMethod: req.PaymentMethod,
		Amount:        req.Amount,
		TradeTime:     t.Unix(),
		Remark:        req.Remark,
	}
	if req.ID > 0 {
		// 修改
		err := config.DB.Model(&model.BillRecord{}).
			Where("id = ? AND user_id = ?", req.ID, userID).
			Updates(bill).Error
		if err != nil {
			response.Fail(c, 100013)
			return
		}
	} else {
		// 新增
		bill.TradeNo = uuid.NewString()
		bill.MerchantOrderNo = uuid.NewString()
		bill.TradeStatus = "交易成功"
		if err := config.DB.Create(&bill).Error; err != nil {
			response.Fail(c, 100013)
			return
		}
	}
	// 返回成功
	response.Ok(c, gin.H{})
}

// 删除交易信息请求体
type DeleteBillRecordRequest struct {
	ID uint `json:"id" binding:"required"`
}

// 删除交易信息接口
func DeleteBillRecordHandler(c *gin.Context) {
	// 获取用户ID
	userIDAny, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, 300001)
		return
	}
	userID, ok := userIDAny.(uint)
	if !ok {
		response.Fail(c, 300002)
		return
	}
	// 获取请求参数
	req, ok := c.MustGet("payload").(DeleteBillRecordRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	if err := config.DB.Where("id = ? and user_id = ?", req.ID, userID).Delete(&model.BillRecord{}).Error; err != nil {
		response.Fail(c, 100014)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{})
}
