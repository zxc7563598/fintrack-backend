package controller

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/dto"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/service"
	"github.com/zxc7563598/fintrack-backend/service/ai"
	"github.com/zxc7563598/fintrack-backend/utils/response"
)

// 获取交易列表请求体
type GetBillListRequest struct {
	StartFormattedDate *string   `json:"start_formatted_date"` // 开始日期
	EndFormattedDate   *string   `json:"end_formatted_date"`   // 结束日期
	Search             *string   `json:"search"`               // 搜索内容
	IncomeType         *int      `json:"income_type"`          // 收支类型
	Page               *int      `json:"page"`                 // 页码
	ItemsPerPage       *int      `json:"items_per_page"`       // 每页条数
	SortKey            *string   `json:"sort_key"`             // 排序字段
	SortOrder          *string   `json:"sort_order"`           // 排序顺序
	PaymentMethod      *[]string `json:"payment_method"`       // 账户
	Counterpartys      *[]string `json:"counterpartys"`        // 交易平台
	TradeTypes         *[]string `json:"trade_types"`          // 交易分类
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
	if req.Counterpartys != nil && len(*req.Counterpartys) > 0 {
		db = db.Where("counterparty IN ?", *req.Counterpartys)
	}
	if req.PaymentMethod != nil && len(*req.PaymentMethod) > 0 {
		db = db.Where("payment_method IN ?", *req.PaymentMethod)
	}
	if req.TradeTypes != nil && len(*req.TradeTypes) > 0 {
		db = db.Where("trade_type IN ?", *req.TradeTypes)
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

// 获取账单日历请求体
type GetBillCalendarRequest struct {
	StartAt string `json:"start_at" binding:"required"`
	EndAt   string `json:"end_at" binding:"required"`
}

// 获取账单日历接口
func GetBillCalendarHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetBillCalendarRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 校验参数
	start, err := time.Parse("2006-01-02", req.StartAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
		return
	}
	end, err := time.Parse("2006-01-02", req.EndAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
		return
	}
	// 统一到当天 00:00:00 和 23:59:59
	startUnix := start.Unix()
	endUnix := end.AddDate(0, 0, 1).Add(-time.Second).Unix()

	// 查询数据库
	var records []model.BillRecord
	if err := config.DB.Where("user_id = ? AND trade_time BETWEEN ? AND ?", userID, startUnix, endUnix).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// 初始化结果 map[日期]DailySummary
	resultMap := make(map[string]*dto.BillDailySummary)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		resultMap[dayStr] = &dto.BillDailySummary{
			Date:    dayStr,
			Income:  0,
			Expense: 0,
		}
	}
	// 累加收入支出
	for _, r := range records {
		day := time.Unix(r.TradeTime, 0).Format("2006-01-02")
		if summary, ok := resultMap[day]; ok {
			switch r.IncomeType {
			case 1: // 收入
				summary.Income += r.Amount
			case 2: // 支出
				summary.Expense += r.Amount
			}
		}
	}
	// 转换为数组，保持日期顺序
	var result []dto.BillDailySummary
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dayStr := d.Format("2006-01-02")
		result = append(result, *resultMap[dayStr])
	}
	// 返回成功
	response.Ok(c, gin.H{
		"data": result,
	})
}

// 账单导出请求体
type ExportBillRequest struct {
	StartFormattedDate *string   `json:"start_formatted_date"` // 开始日期
	EndFormattedDate   *string   `json:"end_formatted_date"`   // 结束日期
	IncomeType         *int      `json:"income_type"`          // 收支类型
	PaymentMethod      *[]string `json:"payment_method"`       // 账户
	Counterpartys      *[]string `json:"counterpartys"`        // 交易平台
	TradeTypes         *[]string `json:"trade_types"`          // 交易分类
}

// 账单导出接口
func ExportBillHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(ExportBillRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取账单数据
	var records []dto.BillExportItem
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
	if req.IncomeType != nil {
		db = db.Where("income_type = ?", *req.IncomeType)
	}
	if req.Counterpartys != nil && len(*req.Counterpartys) > 0 {
		db = db.Where("counterparty IN ?", *req.Counterpartys)
	}
	if req.PaymentMethod != nil && len(*req.PaymentMethod) > 0 {
		db = db.Where("payment_method IN ?", *req.PaymentMethod)
	}
	if req.TradeTypes != nil && len(*req.TradeTypes) > 0 {
		db = db.Where("trade_type IN ?", *req.TradeTypes)
	}
	// 排序
	sortKey := "trade_time"
	sortOrder := "desc"
	db = db.Order(fmt.Sprintf("%s %s", sortKey, sortOrder))
	// 查询
	if err := db.Find(&records).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 新建 CSV writer
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(&buf)
	// 写数据
	writer.Write([]string{"交易号", "商户订单号", "平台", "收支类型", "交易类型", "商品名称", "对方", "支付方式", "金额", "交易状态", "交易时间", "备注"})
	for _, r := range records {
		writer.Write([]string{
			r.TradeNo,
			r.MerchantOrderNo,
			map[uint8]string{1: "微信", 2: "支付宝"}[r.Platform],
			map[uint8]string{1: "收入", 2: "支出", 3: "不计收支", 4: "未知"}[r.IncomeType],
			r.TradeType,
			r.ProductName,
			r.Counterparty,
			r.PaymentMethod,
			strconv.FormatFloat(r.Amount, 'f', 2, 64),
			r.TradeStatus,
			time.Unix(r.TradeTime, 0).Format("2006-01-02 15:04:05"),
			r.Remark,
		})
	}
	writer.Flush()
	// 返回字节流
	c.Data(http.StatusOK, "text/csv", buf.Bytes())
}

// AI账单分析请求体
type AnalysisBillRequest struct {
	StartFormattedDate *string   `json:"start_formatted_date"` // 开始日期
	EndFormattedDate   *string   `json:"end_formatted_date"`   // 结束日期
	IncomeType         *int      `json:"income_type"`          // 收支类型
	PaymentMethod      *[]string `json:"payment_method"`       // 账户
	Counterpartys      *[]string `json:"counterpartys"`        // 交易平台
	TradeTypes         *[]string `json:"trade_types"`          // 交易分类
}

// AI账单分析接口
func AnalysisBillHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(AnalysisBillRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取账单数据
	var records []dto.BillExportItem
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
	if req.IncomeType != nil {
		db = db.Where("income_type = ?", *req.IncomeType)
	}
	if req.Counterpartys != nil && len(*req.Counterpartys) > 0 {
		db = db.Where("counterparty IN ?", *req.Counterpartys)
	}
	if req.PaymentMethod != nil && len(*req.PaymentMethod) > 0 {
		db = db.Where("payment_method IN ?", *req.PaymentMethod)
	}
	if req.TradeTypes != nil && len(*req.TradeTypes) > 0 {
		db = db.Where("trade_type IN ?", *req.TradeTypes)
	}
	// 排序
	sortKey := "trade_time"
	sortOrder := "desc"
	db = db.Order(fmt.Sprintf("%s %s", sortKey, sortOrder))
	// 查询
	if err := db.Find(&records).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 新建 CSV writer
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	writer := csv.NewWriter(&buf)
	writer.Write([]string{"平台", "收支类型", "交易类型", "商品名称", "对方", "支付方式", "金额", "交易时间", "备注"})
	for _, r := range records {
		writer.Write([]string{
			map[uint8]string{1: "微信", 2: "支付宝"}[r.Platform],
			map[uint8]string{1: "收入", 2: "支出", 3: "不计收支", 4: "未知"}[r.IncomeType],
			r.TradeType,
			r.ProductName,
			r.Counterparty,
			r.PaymentMethod,
			strconv.FormatFloat(r.Amount, 'f', 2, 64),
			time.Unix(r.TradeTime, 0).Format("2006-01-02 15:04:05"),
			r.Remark,
		})
	}
	writer.Flush()
	// 获取用户DeepseekAPIKey
	var apiKey string
	err := config.DB.Model(&model.User{}).Select("deepseek_api_key").Where("id = ?", userID).Scan(&apiKey).Error
	if err != nil || apiKey == "" {
		response.Fail(c, 100018)
		return
	}
	// 初始化 client
	client := deepseek.NewClient(apiKey)
	client.Timeout = 600 * time.Second
	client.HTTPClient = &http.Client{
		Timeout: client.Timeout,
	}
	aiClient := &ai.AIClient{
		Client: client,
	}
	// 初始化 service（也可以复用同一个 service 结构体，只是传入不同 client）
	classifier := service.NewBillAnalysis(aiClient)
	ctx := context.Background()
	report, err := classifier.AnalysisFromBytes(ctx, buf.Bytes())
	if err != nil {
		fmt.Println("生成报告失败:", err)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"report": report,
	})
}
