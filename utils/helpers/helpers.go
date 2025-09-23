package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xuri/excelize/v2"
	"github.com/yeka/zip"
	"golang.org/x/crypto/argon2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func StartOfDay(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return StartOfDay(t).AddDate(0, 0, -weekday+1)
}

func StartOfMonth(t time.Time) time.Time {
	y, m, _ := t.Date()
	return time.Date(y, m, 1, 0, 0, 0, 0, t.Location())
}

func StartOfYear(t time.Time) time.Time {
	y := t.Year()
	return time.Date(y, 1, 1, 0, 0, 0, 0, t.Location())
}

// 生成随机扰乱码
func GenerateSalt(n int) ([]byte, error) {
	salt := make([]byte, n)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return salt, nil
}

// 密码加密
func HashPassword(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.RawStdEncoding.EncodeToString(hash)
}

// 读取 CSV 文件并返回二维数组
func ReadCSV(filePath string) ([][]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("无法打开文件: %w", err)
	}
	defer f.Close()
	// 创建 CSV reader
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	// 读取所有数据
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("读取 CSV 出错: %w", err)
	}
	return records, nil
}

// 读取 XLSX 文件并返回二维数组
func ReadXLSX(filePath string) ([][]string, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	rows, err := f.GetRows("Sheet1")
	if err != nil {
		log.Fatal(err)
	}
	return rows, nil
}

// 微信XLSX基本信息结构体
type WeChatBillSummary struct {
	Nickname      string
	StartTime     string
	EndTime       string
	ExportType    string
	ExportTime    string
	TotalRecords  int
	IncomeCount   int
	IncomeAmount  string
	ExpenseCount  int
	ExpenseAmount string
	NeutralCount  int
	NeutralAmount string
}

// 解析微信XLSX基本信息
func ParseWeChatXLSX(rows [][]string) (*WeChatBillSummary, error) {
	summary := &WeChatBillSummary{}
	for i, row := range rows {
		if len(row) == 0 {
			continue
		}
		line := row[0]
		switch {
		case strings.HasPrefix(line, "微信昵称："):
			summary.Nickname = strings.Trim(line[len("微信昵称："):], "[]")
		case strings.HasPrefix(line, "起始时间："):
			parts := strings.Split(line, "终止时间：")
			if len(parts) == 2 {
				start := strings.Trim(parts[0][len("起始时间："):], "[] ")
				end := strings.Trim(parts[1], "[] ")
				summary.StartTime = start
				summary.EndTime = end
			}
		case strings.HasPrefix(line, "导出类型："):
			summary.ExportType = strings.Trim(line[len("导出类型："):], "[]")
		case strings.HasPrefix(line, "导出时间："):
			summary.ExportTime = strings.Trim(line[len("导出时间："):], "[]")
		case strings.HasPrefix(line, "共"):
			numStr := strings.TrimPrefix(line, "共")
			numStr = strings.TrimSuffix(numStr, "笔记录")
			if n, err := strconv.Atoi(strings.TrimSpace(numStr)); err == nil {
				summary.TotalRecords = n
			}
		case strings.HasPrefix(line, "收入："):
			// 收入：5笔 548.00元
			fmt.Sscanf(line, "收入：%d笔 %s", &summary.IncomeCount, &summary.IncomeAmount)
		case strings.HasPrefix(line, "支出："):
			fmt.Sscanf(line, "支出：%d笔 %s", &summary.ExpenseCount, &summary.ExpenseAmount)
		case strings.HasPrefix(line, "中性交易："):
			fmt.Sscanf(line, "中性交易：%d笔 %s", &summary.NeutralCount, &summary.NeutralAmount)
		}
		if i > 10 {
			break
		}
	}
	return summary, nil
}

// 去掉人民币符号 ¥ 并转成 float64
func ParseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "¥")
	return strconv.ParseFloat(s, 64)
}

// 阿里云CSV基本信息结构体
type ExportInfo struct {
	Name          string
	AlipayAccount string
	StartTime     string
	EndTime       string
	TradeType     string
	ExportTime    string
	TotalCount    int
	IncomeCount   int
	IncomeAmount  float64
	ExpenseCount  int
	ExpenseAmount float64
	NoneCount     int
	NoneAmount    float64
}

// 解析阿里云CSV基本信息
func ParseAlipayCSV(records [][]string) (*ExportInfo, error) {
	info := &ExportInfo{}
	// 正则预编译
	reName := regexp.MustCompile(`姓名\s*[:：]\s*(.+)`)
	reAccount := regexp.MustCompile(`支付宝账户\s*[:：]\s*(.+)`)
	reTime := regexp.MustCompile(`起始时间\s*[:：]\s*\[(.+?)]\s*终止时间\s*[:：]\s*\[(.+?)]`)
	reTradeType := regexp.MustCompile(`导出交易类型\s*[:：]\s*\[(.+?)]`)
	reExportTime := regexp.MustCompile(`导出时间\s*[:：]\s*\[(.+?)]`)
	reTotal := regexp.MustCompile(`共\s*(\d+)\s*笔记录`)
	reIncome := regexp.MustCompile(`收入\s*[:：]\s*(\d+)\s*笔\s*([\d.]+)元`)
	reExpense := regexp.MustCompile(`支出\s*[:：]\s*(\d+)\s*笔\s*([\d.]+)元`)
	reNone := regexp.MustCompile(`不计收支\s*[:：]\s*(\d+)\s*笔\s*([\d.]+)元`)
	for _, row := range records {
		if len(row) == 0 {
			continue
		}
		// 清理文本
		line := strings.TrimSpace(row[0])
		line = strings.TrimPrefix(line, "\uFEFF")
		line = strings.ReplaceAll(line, "\t", " ")
		// 匹配各个字段
		if m := reName.FindStringSubmatch(line); len(m) > 1 {
			info.Name = m[1]
		} else if m := reAccount.FindStringSubmatch(line); len(m) > 1 {
			info.AlipayAccount = m[1]
		} else if m := reTime.FindStringSubmatch(line); len(m) > 2 {
			info.StartTime = m[1]
			info.EndTime = m[2]
		} else if m := reTradeType.FindStringSubmatch(line); len(m) > 1 {
			info.TradeType = m[1]
		} else if m := reExportTime.FindStringSubmatch(line); len(m) > 1 {
			info.ExportTime = m[1]
		} else if m := reTotal.FindStringSubmatch(line); len(m) > 1 {
			info.TotalCount, _ = strconv.Atoi(m[1])
		} else if m := reIncome.FindStringSubmatch(line); len(m) > 2 {
			info.IncomeCount, _ = strconv.Atoi(m[1])
			info.IncomeAmount, _ = strconv.ParseFloat(m[2], 64)
		} else if m := reExpense.FindStringSubmatch(line); len(m) > 2 {
			info.ExpenseCount, _ = strconv.Atoi(m[1])
			info.ExpenseAmount, _ = strconv.ParseFloat(m[2], 64)
		} else if m := reNone.FindStringSubmatch(line); len(m) > 2 {
			info.NoneCount, _ = strconv.Atoi(m[1])
			info.NoneAmount, _ = strconv.ParseFloat(m[2], 64)
		}
	}
	return info, nil
}

// 压缩包爆破
func CrackZipPassword(filePath string) (string, error) {
	start, end := 0, 999999
	numCPU := runtime.NumCPU()
	if numCPU <= 0 {
		numCPU = 4
	}
	runtime.GOMAXPROCS(numCPU)
	step := (end - start + 1) / numCPU
	var wg sync.WaitGroup
	passwordChan := make(chan string, 1)
	var done int32
	for id := 0; id < numCPU; id++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			localStart := start + id*step
			localEnd := localStart + step
			if id == numCPU-1 {
				localEnd = end + 1
			}
			for p := localStart; p < localEnd; p++ {
				if atomic.LoadInt32(&done) == 1 {
					return
				}
				pass := fmt.Sprintf("%0*d", 6, p)
				if tryPassword(filePath, pass) {
					if atomic.CompareAndSwapInt32(&done, 0, 1) {
						passwordChan <- pass
					}
					return
				}
			}
		}(id)
	}
	go func() {
		wg.Wait()
		close(passwordChan)
	}()
	pass, ok := <-passwordChan
	if !ok {
		return "", fmt.Errorf("密码未找到")
	}
	return pass, nil
}

// 通过密码解压压缩包
func tryPassword(filePath, password string) bool {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return false
	}
	defer r.Close()
	buf := make([]byte, 16*1024)
	for _, f := range r.File {
		if f.IsEncrypted() {
			f.SetPassword(password)
		}
		rc, err := f.Open()
		if err != nil {
			return false
		}
		totalRead := 0
		for {
			n, err := rc.Read(buf)
			totalRead += n
			if err == io.EOF {
				break
			}
			if err != nil {
				rc.Close()
				return false
			}
		}
		rc.Close()
		// 验证文件完整读取
		if uint64(totalRead) != f.UncompressedSize64 {
			return false
		}
	}
	return true
}

// 解压带密码的 ZIP 文件到 ZIP 文件所在目录
func UnzipWithPassword(zipPath, password string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()
	// 压缩包所在目录
	baseDir := filepath.Dir(zipPath)
	// 去掉扩展名作为文件夹名
	zipName := strings.TrimSuffix(filepath.Base(zipPath), filepath.Ext(zipPath))
	destDir := filepath.Join(baseDir, zipName)
	// 创建解压目标文件夹
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return "", err
	}
	var extractedFilePath string
	for _, f := range r.File {
		if f.IsEncrypted() {
			f.SetPassword(password)
		}
		destPath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return "", err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
			return "", err
		}
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		outFile, err := os.Create(destPath)
		if err != nil {
			rc.Close()
			return "", err
		}
		if _, err := io.Copy(outFile, rc); err != nil {
			rc.Close()
			outFile.Close()
			return "", err
		}
		rc.Close()
		outFile.Close()
		extractedFilePath = destPath
	}
	if extractedFilePath == "" {
		return "", fmt.Errorf("没有解压出文件")
	}
	return extractedFilePath, nil
}

// 将 CSV 文件从 GBK 转成 UTF-8，保存到同名文件（覆盖或新文件）
func ConvertCSVGBKToUTF8(csvPath string) (string, error) {
	// 打开原 CSV 文件
	inFile, err := os.Open(csvPath)
	if err != nil {
		return "", err
	}
	defer inFile.Close()
	outPath := csvPath + ".utf8.csv"
	outFile, err := os.Create(outPath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()
	// GBK -> UTF8 转码
	decoder := transform.NewReader(inFile, simplifiedchinese.GBK.NewDecoder())
	if _, err := io.Copy(outFile, decoder); err != nil {
		return "", err
	}
	return outPath, err
}

// 拼接绝对路径，兼容桌面应用和服务器模式
func GetDataPath(elem ...string) string {
	var baseDir string
	exePath, err := os.Executable()
	if err != nil {
		// 获取可执行文件失败，fallback 到当前工作目录
		cwd, _ := os.Getwd()
		baseDir = cwd
	} else {
		exeDir := filepath.Dir(exePath)
		switch runtime.GOOS {
		case "darwin":
			// macOS 桌面应用：如果 exe 路径包含 .app/Contents/MacOS，就认为是打包应用
			if filepath.Base(filepath.Dir(filepath.Dir(exeDir))) == ".app" || filepath.Ext(filepath.Dir(filepath.Dir(exeDir))) == ".app" {
				homeDir, _ := os.UserHomeDir()
				baseDir = filepath.Join(homeDir, "Library", "Application Support", "FinancialManagement")
			} else {
				// 开发模式或命令行运行
				baseDir = ""
			}
		case "windows":
			// Windows 打包应用
			appData := os.Getenv("APPDATA")
			if appData != "" {
				baseDir = filepath.Join(appData, "FinancialManagement")
			} else {
				baseDir = ""
			}
		default:
			// 其他平台（Linux/服务器）
			baseDir = ""
		}
	}
	// 拼接最终路径
	finalPath := filepath.Join(append([]string{baseDir}, elem...)...)
	// 确保目录存在
	if err := os.MkdirAll(finalPath, os.ModePerm); err != nil {
		panic(err)
	}
	return finalPath
}
