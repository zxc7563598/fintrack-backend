package i18n

import (
	"encoding/json"
	"log"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var Bundle *goi18n.Bundle

func InitI18n() {
	Bundle = goi18n.NewBundle(language.Chinese)
	Bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	if _, err := Bundle.LoadMessageFile("i18n/en.json"); err != nil {
		log.Fatal(err)
	}
	if _, err := Bundle.LoadMessageFile("i18n/zh.json"); err != nil {
		log.Fatal(err)
	}
}

func NewLocalizer(lang string) *goi18n.Localizer {
	return goi18n.NewLocalizer(Bundle, lang)
}
