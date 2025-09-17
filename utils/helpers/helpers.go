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

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"github.com/yeka/zip"
	"golang.org/x/crypto/argon2"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

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

// ReadCSV 读取 CSV 文件并返回二维数组
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

// 邮件结构体
type Email struct {
	Subject     string
	From        string
	Date        string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	FileName string
	Data     []byte
}

// 获取特定发件人的未读邮件（包含正文和附件）
func GetUnreadEmailsFromSender(addr, password, imapServer, sender string) ([]Email, error) {
	// 连接 IMAP
	c, err := client.DialTLS(imapServer, nil)
	if err != nil {
		return nil, err
	}
	defer c.Logout()
	// 登录
	if err := c.Login(addr, password); err != nil {
		return nil, err
	}
	// 选择收件箱
	mbox, err := c.Select("INBOX", false)
	if err != nil {
		return nil, err
	}
	if mbox.Messages == 0 {
		return []Email{}, nil
	}
	// 搜索未读 + 特定发件人
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{"\\Seen"}
	criteria.Header.Add("FROM", sender)
	ids, err := c.Search(criteria)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []Email{}, nil
	}
	// 获取消息
	seqset := new(imap.SeqSet)
	seqset.AddNum(ids...)
	section := &imap.BodySectionName{}
	messages := make(chan *imap.Message, 10)
	go func() {
		if err := c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages); err != nil {
			log.Fatal(err)
		}
	}()
	var emails []Email
	for msg := range messages {
		if msg == nil {
			continue
		}
		r := msg.GetBody(section)
		if r == nil {
			continue
		}
		// 解析邮件
		mr, err := mail.CreateReader(r)
		if err != nil {
			continue
		}
		email := Email{}
		header := mr.Header
		// 标题
		if subject, err := header.Subject(); err == nil {
			email.Subject = subject
		}
		// 发件人
		if from, err := header.AddressList("From"); err == nil && len(from) > 0 {
			email.From = from[0].String()
		}
		// 时间
		if date, err := header.Date(); err == nil {
			email.Date = date.String()
		}
		// 遍历内容
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				ctype, _, _ := h.ContentType()
				body, _ := io.ReadAll(p.Body)
				switch ctype {
				case "text/plain":
					email.TextBody = string(body)
				case "text/html":
					email.HTMLBody = string(body)
				}
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()
				data, _ := io.ReadAll(p.Body)
				email.Attachments = append(email.Attachments, Attachment{
					FileName: filename,
					Data:     data,
				})
			}
		}
		emails = append(emails, email)
	}
	return emails, nil
}
