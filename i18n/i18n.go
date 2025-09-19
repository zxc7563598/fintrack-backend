package i18n

import (
	"embed"
	"encoding/json"
	"log"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Bundle *goi18n.Bundle
var I18nFS embed.FS

func InitI18n() {
	Bundle = goi18n.NewBundle(language.Chinese)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	// 尝试从嵌入的文件系统读取
	if _, err := I18nFS.Open("i18n/en.json"); err == nil {
		enData, err := I18nFS.ReadFile("i18n/en.json")
		if err != nil {
			log.Fatal("未能读取嵌入内容 en.json:", err)
		}
		if _, err := Bundle.ParseMessageFileBytes(enData, "i18n/en.json"); err != nil {
			log.Fatal("无法解析嵌入内容 en.json:", err)
		}
		zhData, err := I18nFS.ReadFile("i18n/zh.json")
		if err != nil {
			log.Fatal("未能读取嵌入内容 zh.json:", err)
		}
		if _, err := Bundle.ParseMessageFileBytes(zhData, "i18n/zh.json"); err != nil {
			log.Fatal("无法解析嵌入内容 zh.json:", err)
		}
	} else {
		// 回退到文件系统读取（开发环境）
		if _, err := Bundle.LoadMessageFile("i18n/en.json"); err != nil {
			log.Fatal(err)
		}
		if _, err := Bundle.LoadMessageFile("i18n/zh.json"); err != nil {
			log.Fatal(err)
		}
	}
	log.Println("✅ i18n文件已成功加载")
}

func SetI18nFS(fs embed.FS) {
	I18nFS = fs
}

func NewLocalizer(lang string) *goi18n.Localizer {
	return goi18n.NewLocalizer(Bundle, lang)
}
