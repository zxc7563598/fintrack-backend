package controller

import (
	"time"

	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"github.com/zxc7563598/fintrack-backend/utils/response"

	"github.com/gin-gonic/gin"
)

type MonthlyStat struct {
	Year    int           `json:"year"`
	Month   int           `json:"month"`
	Income  helpers.Money `json:"income"`
	Expense helpers.Money `json:"expense"`
}

type BillSummary struct {
	TotalCount   int64         `json:"total_count"`
	LastRecord   int64         `json:"last_record"`
	TotalIncome  helpers.Money `json:"total_income"`
	TotalExpense helpers.Money `json:"total_expense"`
	TodayIncome  helpers.Money `json:"today_income"`
	TodayExpense helpers.Money `json:"today_expense"`
	WeekIncome   helpers.Money `json:"week_income"`
	WeekExpense  helpers.Money `json:"week_expense"`
	MonthIncome  helpers.Money `json:"month_income"`
	MonthExpense helpers.Money `json:"month_expense"`
	YearIncome   helpers.Money `json:"year_income"`
	YearExpense  helpers.Money `json:"year_expense"`
	Last12Months []MonthlyStat `json:"last12_months"`
}

func AssetOverviewHandler(c *gin.Context) {
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
	var summary BillSummary
	now := time.Now()
	// 总计收入、总计支出、总笔数（一次聚合查询）
	type aggResult struct {
		Income, Expense float64
		Count           int64
	}
	var agg aggResult
	config.DB.Model(&model.BillRecord{}).
		Where("user_id = ?", userID).
		Select("SUM(CASE WHEN income_type = 1 THEN amount ELSE 0 END) as income, SUM(CASE WHEN income_type = 2 THEN amount ELSE 0 END) as expense, COUNT(*) as count").
		Scan(&agg)
	summary.TotalIncome = helpers.Money(agg.Income)
	summary.TotalExpense = helpers.Money(agg.Expense)
	summary.TotalCount = agg.Count
	// 最新一条记录
	var times []int64
	err := config.DB.Model(&model.BillRecord{}).Where("user_id = ?", userID).Order("trade_time DESC").Limit(1).Pluck("trade_time", &times).Error
	if err != nil || len(times) == 0 {
		summary.LastRecord = 0
	} else {
		summary.LastRecord = times[0]
	}
	// 本年及本月数据，用于计算今日/本周/本月/本年收入支出 + 最近12个月
	yearStart := helpers.StartOfYear(now)
	monthStart := helpers.StartOfMonth(now)
	twelveMonthsAgo := monthStart.AddDate(0, -11, 0) // 最近12个月
	var records []model.BillRecord
	config.DB.Where("user_id = ? AND trade_time >= ?", userID, twelveMonthsAgo.Unix()).Find(&records)
	// 初始化 Last12Months
	summary.Last12Months = make([]MonthlyStat, 12)
	for i := 0; i < 12; i++ {
		t := monthStart.AddDate(0, -11+i, 0)
		summary.Last12Months[i] = MonthlyStat{
			Year:  t.Year(),
			Month: int(t.Month()),
		}
	}
	todayStart := helpers.StartOfDay(now)
	weekStart := helpers.StartOfWeek(now)
	// 遍历计算
	for _, r := range records {
		amt := r.Amount
		t := time.Unix(r.TradeTime, 0)
		// 本年收入/支出
		if t.After(yearStart) || t.Equal(yearStart) {
			if r.IncomeType == 1 {
				summary.YearIncome += helpers.Money(amt)
			}
			if r.IncomeType == 2 {
				summary.YearExpense += helpers.Money(amt)
			}
		}
		// 本月收入/支出
		if t.After(monthStart) || t.Equal(monthStart) {
			if r.IncomeType == 1 {
				summary.MonthIncome += helpers.Money(amt)
			}
			if r.IncomeType == 2 {
				summary.MonthExpense += helpers.Money(amt)
			}
		}
		// 本周收入/支出
		if t.After(weekStart) || t.Equal(weekStart) {
			if r.IncomeType == 1 {
				summary.WeekIncome += helpers.Money(amt)
			}
			if r.IncomeType == 2 {
				summary.WeekExpense += helpers.Money(amt)
			}
		}
		// 今日收入/支出
		if t.After(todayStart) || t.Equal(todayStart) {
			if r.IncomeType == 1 {
				summary.TodayIncome += helpers.Money(amt)
			}
			if r.IncomeType == 2 {
				summary.TodayExpense += helpers.Money(amt)
			}
		}
		// 最近12个月收入/支出
		for i := range summary.Last12Months {
			m := summary.Last12Months[i]
			if t.Year() == m.Year && int(t.Month()) == m.Month {
				if r.IncomeType == 1 {
					summary.Last12Months[i].Income += helpers.Money(amt)
				}
				if r.IncomeType == 2 {
					summary.Last12Months[i].Expense += helpers.Money(amt)
				}
				break
			}
		}
	}
	// 返回数据
	response.Ok(c, gin.H{
		"summary": summary,
	})
}
