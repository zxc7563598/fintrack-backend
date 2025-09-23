package controller

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"github.com/zxc7563598/fintrack-backend/utils/response"
)

// 获取交易列表请求体
type GetStatisticsRequest struct {
	StartFormattedDate *string   `json:"start_formatted_date"` // 开始日期
	EndFormattedDate   *string   `json:"end_formatted_date"`   // 结束日期
	IncomeType         *int      `json:"income_type"`          // 收支类型
	PaymentMethod      *[]string `json:"payment_method"`       // 账户
	Counterpartys      *[]string `json:"counterpartys"`        // 交易平台
	TradeTypes         *[]string `json:"trade_types"`          // 交易分类
}

type AmountSummary struct {
	IncomeTotal  helpers.Money `json:"income_total"`
	ExpenseTotal helpers.Money `json:"expense_total"`
}

// 账户收支（分类图）
func AccountBalanceCategoryHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
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
	// 查询数据
	var summary AmountSummary
	err := db.Select(`
		COALESCE(SUM(CASE WHEN income_type = 1 THEN amount END),0) AS income_total,
		COALESCE(SUM(CASE WHEN income_type = 2 THEN amount END),0) AS expense_total
	`).Scan(&summary).Error
	if err != nil {
		response.Fail(c, 100001)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"income_total":  helpers.Money(summary.IncomeTotal),
		"expense_total": helpers.Money(summary.ExpenseTotal),
	})
}

type TradeTypeIncome struct {
	TradeType   string        `json:"trade_type"`
	IncomeTotal helpers.Money `json:"income_total"`
}

// 收入分类（分类图）
func IncomeCategoryHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
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
	// 查询数据
	var results []TradeTypeIncome
	err := db.
		Select("trade_type, COALESCE(SUM(amount),0) AS income_total").
		Where("income_type = ?", 1).
		Group("trade_type").
		Having("SUM(amount) > 0").
		Scan(&results).Error
	if err != nil {
		response.Fail(c, 100001)
		return
	}
	response.Ok(c, gin.H{
		"list": results,
	})
}

type TradeTypeExpense struct {
	TradeType    string        `json:"trade_type"`
	ExpenseTotal helpers.Money `json:"expense_total"`
}

// 支出分类（分类图）
func ExpenseCategoryHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
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
	// 查询数据
	var results []TradeTypeExpense
	err := db.
		Select("trade_type, COALESCE(SUM(amount),0) AS expense_total").
		Where("income_type = ?", 2).
		Group("trade_type").
		Having("SUM(amount) > 0").
		Scan(&results).Error
	if err != nil {
		response.Fail(c, 100001)
		return
	}
	response.Ok(c, gin.H{
		"list": results,
	})
}

type PaymentMethodIncome struct {
	PaymentMethod string        `json:"payment_method"`
	Amount        helpers.Money `json:"amount"`
}

// 收入账户（分类图）
func IncomeAccountCategoryHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
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
	// 查询数据
	var results []PaymentMethodIncome
	err := db.
		Select("payment_method, COALESCE(SUM(amount),0) AS amount").
		Where("income_type = ?", 1).
		Group("payment_method").
		Having("SUM(amount) > 0").
		Scan(&results).Error
	if err != nil {
		response.Fail(c, 100001)
		return
	}
	response.Ok(c, gin.H{
		"list": results,
	})
}

type PaymentMethodExpense struct {
	PaymentMethod string        `json:"payment_method"`
	Amount        helpers.Money `json:"amount"`
}

// 支出账户（分类图）
func ExpenseAccountCategoryHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
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
	// 查询数据
	var results []PaymentMethodExpense
	err := db.
		Select("payment_method, COALESCE(SUM(amount),0) AS amount").
		Where("income_type = ?", 2).
		Group("payment_method").
		Having("SUM(amount) > 0").
		Scan(&results).Error
	if err != nil {
		response.Fail(c, 100001)
		return
	}
	response.Ok(c, gin.H{
		"list": results,
	})
}

// 账户收支（趋势图）
func AccountBalanceTrendHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	var startTimestamp, endTimestamp int64
	now := time.Now()
	// 结束时间：如果没有传，默认本月月末
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp = endOfDay.Unix()
		}
	} else {
		// 默认本月最后一日
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfNextMonth := time.Date(currentYear, currentMonth+1, 1, 0, 0, 0, 0, location)
		endTimestamp = firstOfNextMonth.Add(-time.Second).Unix()
	}
	// 开始时间：如果没有传，默认12个月前的1日
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp = t.Unix()
		}
	} else {
		// 12个月前的1日
		startYear, startMonth, _ := now.AddDate(0, -11, 0).Date()
		location := now.Location()
		startTimestamp = time.Date(startYear, startMonth, 1, 0, 0, 0, 0, location).Unix()
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
	// 查询数据
	type IncomeTypeMonth struct {
		IncomeType string        `json:"income_type"`
		Month      string        `json:"month"`
		Amount     helpers.Money `json:"amount"`
	}
	var results []IncomeTypeMonth
	err := db.
		Select(`
		income_type, 
		strftime('%Y-%m', datetime(trade_time, 'unixepoch')) AS month, 
		COALESCE(SUM(amount),0) AS amount
	`).
		Where("trade_time >= ? AND trade_time <= ?", startTimestamp, endTimestamp).
		Group("income_type, month").
		Order("income_type, month").
		Scan(&results).Error
	if err != nil {
		log.Println(err)
		response.Fail(c, 100001)
		return
	}
	// 确定月份列表
	layoutMonth := "2006-01"
	startMonth := time.Unix(startTimestamp, 0)
	startMonth = time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, startMonth.Location())
	endMonth := time.Unix(endTimestamp, 0)
	endMonth = time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, endMonth.Location())
	months := []string{}
	for t := startMonth; !t.After(endMonth); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format(layoutMonth))
	}
	// 收集所有分类
	incomeTypeMap := map[string]struct{}{}
	for _, r := range results {
		incomeTypeMap[r.IncomeType] = struct{}{}
	}
	incomeType := []string{}
	for k := range incomeTypeMap {
		incomeType = append(incomeType, k)
	}
	// 构建 map[tradeType][month]float64，并填充0
	dataMap := map[string]map[string]helpers.Money{}
	for _, account := range incomeType {
		dataMap[account] = map[string]helpers.Money{}
		for _, m := range months {
			dataMap[account][m] = 0
		}
	}
	for _, r := range results {
		dataMap[r.IncomeType][r.Month] = helpers.Money(r.Amount)
	}
	// 构建前端折线图格式
	type LineData struct {
		IncomeType string          `json:"income_type"`
		Data       []helpers.Money `json:"data"`
	}
	lineData := []LineData{}
	for _, account := range incomeType {
		row := LineData{IncomeType: account, Data: []helpers.Money{}}
		for _, m := range months {
			row.Data = append(row.Data, helpers.Money(dataMap[account][m]))
		}
		lineData = append(lineData, row)
	}
	// 返回
	response.Ok(c, gin.H{
		"months": months,
		"list":   lineData,
	})

}

// 收入分类（趋势图）
func IncomeCategoryTrendHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	var startTimestamp, endTimestamp int64
	now := time.Now()
	// 结束时间：如果没有传，默认本月月末
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp = endOfDay.Unix()
		}
	} else {
		// 默认本月最后一日
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfNextMonth := time.Date(currentYear, currentMonth+1, 1, 0, 0, 0, 0, location)
		endTimestamp = firstOfNextMonth.Add(-time.Second).Unix()
	}
	// 开始时间：如果没有传，默认12个月前的1日
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp = t.Unix()
		}
	} else {
		// 12个月前的1日
		startYear, startMonth, _ := now.AddDate(0, -11, 0).Date()
		location := now.Location()
		startTimestamp = time.Date(startYear, startMonth, 1, 0, 0, 0, 0, location).Unix()
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
	// 查询数据
	type TradeTypeMonthIncome struct {
		TradeType string        `json:"trade_type"`
		Month     string        `json:"month"`
		Income    helpers.Money `json:"income"`
	}
	var results []TradeTypeMonthIncome
	err := db.
		Select(`
		trade_type, 
		strftime('%Y-%m', datetime(trade_time, 'unixepoch')) AS month, 
		COALESCE(SUM(amount),0) AS income
	`).
		Where("income_type = ?", 1).
		Where("trade_time >= ? AND trade_time <= ?", startTimestamp, endTimestamp).
		Group("trade_type, month").
		Order("trade_type, month").
		Scan(&results).Error
	if err != nil {
		log.Println(err)
		response.Fail(c, 100001)
		return
	}
	// 确定月份列表
	layoutMonth := "2006-01"
	startMonth := time.Unix(startTimestamp, 0)
	startMonth = time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, startMonth.Location())
	endMonth := time.Unix(endTimestamp, 0)
	endMonth = time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, endMonth.Location())
	months := []string{}
	for t := startMonth; !t.After(endMonth); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format(layoutMonth))
	}
	// 收集所有分类
	tradeTypesMap := map[string]struct{}{}
	for _, r := range results {
		tradeTypesMap[r.TradeType] = struct{}{}
	}
	tradeTypes := []string{}
	for k := range tradeTypesMap {
		tradeTypes = append(tradeTypes, k)
	}
	// 构建 map[tradeType][month]float64，并填充0
	dataMap := map[string]map[string]helpers.Money{}
	for _, trade := range tradeTypes {
		dataMap[trade] = map[string]helpers.Money{}
		for _, m := range months {
			dataMap[trade][m] = 0
		}
	}
	for _, r := range results {
		dataMap[r.TradeType][r.Month] = helpers.Money(r.Income)
	}
	// 构建前端折线图格式
	type LineData struct {
		TradeType string          `json:"trade_type"`
		Data      []helpers.Money `json:"data"`
	}
	lineData := []LineData{}
	for _, trade := range tradeTypes {
		row := LineData{TradeType: trade, Data: []helpers.Money{}}
		for _, m := range months {
			row.Data = append(row.Data, helpers.Money(dataMap[trade][m]))
		}
		lineData = append(lineData, row)
	}
	// 返回
	response.Ok(c, gin.H{
		"months": months,
		"list":   lineData,
	})

}

// 支出分类（趋势图）
func ExpenseCategoryTrendHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	var startTimestamp, endTimestamp int64
	now := time.Now()
	// 结束时间：如果没有传，默认本月月末
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp = endOfDay.Unix()
		}
	} else {
		// 默认本月最后一日
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfNextMonth := time.Date(currentYear, currentMonth+1, 1, 0, 0, 0, 0, location)
		endTimestamp = firstOfNextMonth.Add(-time.Second).Unix()
	}
	// 开始时间：如果没有传，默认12个月前的1日
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp = t.Unix()
		}
	} else {
		// 12个月前的1日
		startYear, startMonth, _ := now.AddDate(0, -11, 0).Date()
		location := now.Location()
		startTimestamp = time.Date(startYear, startMonth, 1, 0, 0, 0, 0, location).Unix()
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
	// 查询数据
	type TradeTypeMonthExpense struct {
		TradeType string        `json:"trade_type"`
		Month     string        `json:"month"`
		Expense   helpers.Money `json:"expense"`
	}
	var results []TradeTypeMonthExpense
	err := db.
		Select(`
		trade_type, 
		strftime('%Y-%m', datetime(trade_time, 'unixepoch')) AS month, 
		COALESCE(SUM(amount),0) AS expense
	`).
		Where("income_type = ?", 2).
		Where("trade_time >= ? AND trade_time <= ?", startTimestamp, endTimestamp).
		Group("trade_type, month").
		Order("trade_type, month").
		Scan(&results).Error
	if err != nil {
		log.Println(err)
		response.Fail(c, 100001)
		return
	}
	// 确定月份列表
	layoutMonth := "2006-01"
	startMonth := time.Unix(startTimestamp, 0)
	startMonth = time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, startMonth.Location())
	endMonth := time.Unix(endTimestamp, 0)
	endMonth = time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, endMonth.Location())
	months := []string{}
	for t := startMonth; !t.After(endMonth); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format(layoutMonth))
	}
	// 收集所有分类
	tradeTypesMap := map[string]struct{}{}
	for _, r := range results {
		tradeTypesMap[r.TradeType] = struct{}{}
	}
	tradeTypes := []string{}
	for k := range tradeTypesMap {
		tradeTypes = append(tradeTypes, k)
	}
	// 构建 map[tradeType][month]float64，并填充0
	dataMap := map[string]map[string]helpers.Money{}
	for _, trade := range tradeTypes {
		dataMap[trade] = map[string]helpers.Money{}
		for _, m := range months {
			dataMap[trade][m] = 0
		}
	}
	for _, r := range results {
		dataMap[r.TradeType][r.Month] = helpers.Money(r.Expense)
	}
	// 构建前端折线图格式
	type LineData struct {
		TradeType string          `json:"trade_type"`
		Data      []helpers.Money `json:"data"`
	}
	lineData := []LineData{}
	for _, trade := range tradeTypes {
		row := LineData{TradeType: trade, Data: []helpers.Money{}}
		for _, m := range months {
			row.Data = append(row.Data, helpers.Money(dataMap[trade][m]))
		}
		lineData = append(lineData, row)
	}
	// 返回
	response.Ok(c, gin.H{
		"months": months,
		"list":   lineData,
	})

}

// 收入账户（趋势图）
func IncomeAccountTrendHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	var startTimestamp, endTimestamp int64
	now := time.Now()
	// 结束时间：如果没有传，默认本月月末
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp = endOfDay.Unix()
		}
	} else {
		// 默认本月最后一日
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfNextMonth := time.Date(currentYear, currentMonth+1, 1, 0, 0, 0, 0, location)
		endTimestamp = firstOfNextMonth.Add(-time.Second).Unix()
	}
	// 开始时间：如果没有传，默认12个月前的1日
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp = t.Unix()
		}
	} else {
		// 12个月前的1日
		startYear, startMonth, _ := now.AddDate(0, -11, 0).Date()
		location := now.Location()
		startTimestamp = time.Date(startYear, startMonth, 1, 0, 0, 0, 0, location).Unix()
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
	// 查询数据
	type TradeTypeMonthIncome struct {
		PaymentMethod string        `json:"payment_method"`
		Month         string        `json:"month"`
		Income        helpers.Money `json:"income"`
	}
	var results []TradeTypeMonthIncome
	err := db.
		Select(`
		payment_method, 
		strftime('%Y-%m', datetime(trade_time, 'unixepoch')) AS month, 
		COALESCE(SUM(amount),0) AS income
	`).
		Where("income_type = ?", 1).
		Where("trade_time >= ? AND trade_time <= ?", startTimestamp, endTimestamp).
		Group("payment_method, month").
		Order("payment_method, month").
		Scan(&results).Error
	if err != nil {
		log.Println(err)
		response.Fail(c, 100001)
		return
	}
	// 确定月份列表
	layoutMonth := "2006-01"
	startMonth := time.Unix(startTimestamp, 0)
	startMonth = time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, startMonth.Location())
	endMonth := time.Unix(endTimestamp, 0)
	endMonth = time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, endMonth.Location())
	months := []string{}
	for t := startMonth; !t.After(endMonth); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format(layoutMonth))
	}
	// 收集所有分类
	paymentMethodMap := map[string]struct{}{}
	for _, r := range results {
		paymentMethodMap[r.PaymentMethod] = struct{}{}
	}
	paymentMethod := []string{}
	for k := range paymentMethodMap {
		paymentMethod = append(paymentMethod, k)
	}
	// 构建 map[paymentMethod][month]float64，并填充0
	dataMap := map[string]map[string]helpers.Money{}
	for _, method := range paymentMethod {
		dataMap[method] = map[string]helpers.Money{}
		for _, m := range months {
			dataMap[method][m] = 0
		}
	}
	for _, r := range results {
		dataMap[r.PaymentMethod][r.Month] = helpers.Money(r.Income)
	}
	// 构建前端折线图格式
	type LineData struct {
		PaymentMethod string          `json:"payment_method"`
		Data          []helpers.Money `json:"data"`
	}
	lineData := []LineData{}
	for _, method := range paymentMethod {
		row := LineData{PaymentMethod: method, Data: []helpers.Money{}}
		for _, m := range months {
			row.Data = append(row.Data, helpers.Money(dataMap[method][m]))
		}
		lineData = append(lineData, row)
	}
	// 返回
	response.Ok(c, gin.H{
		"months": months,
		"list":   lineData,
	})

}

// 支出账户（趋势图）
func ExpenseAccountTrendHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetStatisticsRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取数据
	db := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID)
	// 搜索条件
	var startTimestamp, endTimestamp int64
	now := time.Now()
	// 结束时间：如果没有传，默认本月月末
	if req.EndFormattedDate != nil && *req.EndFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.EndFormattedDate, time.Local)
		if err == nil {
			endOfDay := t.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			endTimestamp = endOfDay.Unix()
		}
	} else {
		// 默认本月最后一日
		currentYear, currentMonth, _ := now.Date()
		location := now.Location()
		firstOfNextMonth := time.Date(currentYear, currentMonth+1, 1, 0, 0, 0, 0, location)
		endTimestamp = firstOfNextMonth.Add(-time.Second).Unix()
	}
	// 开始时间：如果没有传，默认12个月前的1日
	if req.StartFormattedDate != nil && *req.StartFormattedDate != "" {
		t, err := time.ParseInLocation(layout, *req.StartFormattedDate, time.Local)
		if err == nil {
			startTimestamp = t.Unix()
		}
	} else {
		// 12个月前的1日
		startYear, startMonth, _ := now.AddDate(0, -11, 0).Date()
		location := now.Location()
		startTimestamp = time.Date(startYear, startMonth, 1, 0, 0, 0, 0, location).Unix()
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
	// 查询数据
	type TradeTypeMonthExpense struct {
		PaymentMethod string        `json:"payment_method"`
		Month         string        `json:"month"`
		Expense       helpers.Money `json:"expense"`
	}
	var results []TradeTypeMonthExpense
	err := db.
		Select(`
		payment_method, 
		strftime('%Y-%m', datetime(trade_time, 'unixepoch')) AS month, 
		COALESCE(SUM(amount),0) AS expense
	`).
		Where("income_type = ?", 2).
		Where("trade_time >= ? AND trade_time <= ?", startTimestamp, endTimestamp).
		Group("payment_method, month").
		Order("payment_method, month").
		Scan(&results).Error
	if err != nil {
		log.Println(err)
		response.Fail(c, 100001)
		return
	}
	// 确定月份列表
	layoutMonth := "2006-01"
	startMonth := time.Unix(startTimestamp, 0)
	startMonth = time.Date(startMonth.Year(), startMonth.Month(), 1, 0, 0, 0, 0, startMonth.Location())
	endMonth := time.Unix(endTimestamp, 0)
	endMonth = time.Date(endMonth.Year(), endMonth.Month(), 1, 0, 0, 0, 0, endMonth.Location())
	months := []string{}
	for t := startMonth; !t.After(endMonth); t = t.AddDate(0, 1, 0) {
		months = append(months, t.Format(layoutMonth))
	}
	// 收集所有分类
	paymentMethodMap := map[string]struct{}{}
	for _, r := range results {
		paymentMethodMap[r.PaymentMethod] = struct{}{}
	}
	paymentMethod := []string{}
	for k := range paymentMethodMap {
		paymentMethod = append(paymentMethod, k)
	}
	// 构建 map[paymentMethod][month]float64，并填充0
	dataMap := map[string]map[string]helpers.Money{}
	for _, method := range paymentMethod {
		dataMap[method] = map[string]helpers.Money{}
		for _, m := range months {
			dataMap[method][m] = 0
		}
	}
	for _, r := range results {
		dataMap[r.PaymentMethod][r.Month] = helpers.Money(r.Expense)
	}
	// 构建前端折线图格式
	type LineData struct {
		PaymentMethod string          `json:"payment_method"`
		Data          []helpers.Money `json:"data"`
	}
	lineData := []LineData{}
	for _, method := range paymentMethod {
		row := LineData{PaymentMethod: method, Data: []helpers.Money{}}
		for _, m := range months {
			row.Data = append(row.Data, helpers.Money(dataMap[method][m]))
		}
		lineData = append(lineData, row)
	}
	// 返回
	response.Ok(c, gin.H{
		"months": months,
		"list":   lineData,
	})

}
