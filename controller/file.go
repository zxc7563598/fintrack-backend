package controller

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/dto"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"github.com/zxc7563598/fintrack-backend/utils/response"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 支付宝账单CSV文件上传接口
func UploadAlipayCSVHandler(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 100008)
		return
	}
	// 打开上传的文件
	f, err := file.Open()
	if err != nil {
		response.Fail(c, 100007)
		return
	}
	defer f.Close()
	// 确保保存目录存在
	saveDir := "./data/uploads"
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		response.Fail(c, 100009)
		return
	}
	// 生成随机文件名
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".csv"
	}
	newFileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	dst := filepath.Join(saveDir, newFileName)
	// 创建目标文件
	outFile, err := os.Create(dst)
	if err != nil {
		response.Fail(c, 100007)
		return
	}
	defer outFile.Close()
	// 使用 GBK -> UTF8 转换
	decoder := transform.NewReader(f, simplifiedchinese.GBK.NewDecoder())
	// 复制文件内容到目标文件（已经转成 UTF-8）
	if _, err := io.Copy(outFile, decoder); err != nil {
		response.Fail(c, 100007)
		return
	}
	// 验证数据
	records, err := helpers.ReadCSV(dst)
	if err != nil {
		log.Fatal("读取失败:", err)
		response.Fail(c, 100008)
		return
	}
	csv, err := helpers.ParseAlipayCSV(records)
	if err != nil {
		log.Fatal("读取失败:", err)
		response.Fail(c, 100008)
		return
	}
	if csv.Name == "" ||
		csv.AlipayAccount == "" ||
		csv.StartTime == "" ||
		csv.EndTime == "" ||
		csv.TradeType == "" ||
		csv.ExportTime == "" ||
		csv.TotalCount == 0 {
		response.Fail(c, 100010)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"path": dst,
	})
}

// 支付宝账单ZIP文件上传接口
func UploadAlipayZIPHandler(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, 100008)
		return
	}
	// 打开上传的文件
	f, err := file.Open()
	if err != nil {
		response.Fail(c, 100007)
		return
	}
	defer f.Close()
	// 确保保存目录存在
	saveDir := "./data/uploads"
	if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
		response.Fail(c, 100009)
		return
	}
	// 生成随机文件名
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".zip"
	}
	newFileName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	dst := filepath.Join(saveDir, newFileName)
	// 创建目标文件
	outFile, err := os.Create(dst)
	if err != nil {
		response.Fail(c, 100007)
		return
	}
	defer outFile.Close()
	// 写入文件内容
	if _, err := io.Copy(outFile, f); err != nil {
		response.Fail(c, 100007)
		return
	}
	zipPassword := c.PostForm("zip_salt")
	if zipPassword == "" {
		password, err := helpers.CrackZipPassword(dst)
		if err != nil {
			response.Fail(c, 100015)
			return
		}
		zipPassword = password
	}
	// 解压文件
	extractedFilePath, unzip := helpers.UnzipWithPassword(dst, zipPassword)
	if unzip != nil {
		response.Fail(c, 100016)
		return
	}
	path, err := helpers.ConvertCSVGBKToUTF8(extractedFilePath)
	if err != nil {
		response.Fail(c, 100007)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"path": path,
	})
}

// 获取支付宝CSV概览信息请求体
type GetAlipayCSVOverviewRequest struct {
	Path string `json:"path" binding:"required"` // CSV路径
}

// 获取支付宝CSV概览信息接口
func GetAlipayCSVOverviewHandler(c *gin.Context) {
	// 获取参数
	req := c.MustGet("payload").(GetAlipayCSVOverviewRequest)
	records, err := helpers.ReadCSV(req.Path)
	if err != nil {
		log.Fatal("读取失败:", err)
		response.Fail(c, 100008)
		return
	}
	csv, err := helpers.ParseAlipayCSV(records)
	if err != nil {
		log.Fatal("读取失败:", err)
		response.Fail(c, 100008)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"name":           csv.Name,
		"alipay_account": csv.AlipayAccount,
		"start_time":     csv.StartTime,
		"end_time":       csv.EndTime,
		"trade_type":     csv.TradeType,
		"export_time":    csv.ExportTime,
		"total_count":    csv.TotalCount,
		"income_count":   csv.IncomeCount,
		"income_amount":  csv.IncomeAmount,
		"expense_count":  csv.ExpenseCount,
		"expense_amount": csv.ExpenseAmount,
		"none_count":     csv.NoneCount,
		"none_amount":    csv.NoneAmount,
	})
}

// 存储支付宝CSV账单数据请求体
type StoreAlipayCSVInfoRequest struct {
	Path string `json:"path" binding:"required"` // CSV路径
}

// 存储支付宝CSV账单数据接口
func StoreAlipayCSVInfoHandler(c *gin.Context) {
	layout := "2006-1-2 15:04:05"
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
	req, ok := c.MustGet("payload").(StoreAlipayCSVInfoRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 读取 CSV
	records, err := helpers.ReadCSV(req.Path)
	if err != nil {
		response.Fail(c, 100008)
		return
	}
	if len(records) <= 23 {
		response.Fail(c, 100010)
		return
	}
	// 开启事务
	tx := config.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.Fail(c, 100011)
		}
	}()
	for i := 23; i < len(records); i++ {
		row := records[i]
		if len(row) < 12 {
			continue
		}
		// 去除多余的空格
		for i := range row {
			row[i] = strings.TrimSpace(row[i])
		}
		// 判断是否已存在
		var exist model.BillRecord
		if err := tx.Where("trade_no = ?", row[9]).First(&exist).Error; err == nil {
			continue
		}
		// 解析时间
		t, err := time.ParseInLocation(layout, row[0], time.Local)
		if err != nil {
			continue
		}
		// 解析金额
		amount, err := strconv.ParseFloat(row[6], 64)
		if err != nil {
			continue
		}
		// 构建记录
		var paymentMethod string
		if row[7] == "" {
			paymentMethod = "未知"
		} else {
			paymentMethod = row[7]
		}
		bill := model.BillRecord{
			UserID:          userID,
			TradeNo:         row[9],
			MerchantOrderNo: row[10],
			Platform:        uint8(model.PlatformAlipay),
			IncomeType:      model.IncomeTypeFromString(row[5]),
			TradeType:       row[1],
			ProductName:     row[4],
			Counterparty:    row[2],
			PaymentMethod:   paymentMethod,
			Amount:          amount,
			TradeStatus:     row[8],
			TradeTime:       t.Unix(),
			Remark:          row[11],
		}
		if err := tx.Create(&bill).Error; err != nil {
			tx.Rollback()
			response.Fail(c, 100006)
			return
		}
	}
	if err := tx.Commit().Error; err != nil {
		response.Fail(c, 100007)
		return
	}
	// 返回数据
	response.Ok(c, gin.H{})
}

// 获取支付宝账单邮件请求体
type GetAlipayBillMailRequest struct {
	ID uint `json:"id" binding:"required"` // 邮箱ID
}

// 获取支付宝账单邮件接口
func GetAlipayBillMailHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(GetAlipayBillMailRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 获取绑定邮件
	var userMailbox dto.UserMailboxListItem
	model := config.DB.Model(&model.UserMailbox{}).Where("id = ? and user_id = ?", req.ID, userID).First(&userMailbox)
	if model.Error != nil {
		response.Fail(c, 100001)
		return
	}
	// 获取邮件信息
	// 待定···
	// 未来实现
}
