package response

import (
	"strconv"

	"github.com/zxc7563598/fintrack-backend/i18n"

	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
)

// 10 - 业务/业务逻辑：应用层的业务验证错误
// 20 - 数据库/持久化：数据库操作相关的错误
// 30 - 中间件/框架：请求解析、认证、限流等错误
// 40 - 系统/通用：系统内部、第三方服务等错误

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

func JSON(c *gin.Context, code int, data interface{}) {
	langAny, _ := c.Get("lang")
	lang := "zh"
	if l, ok := langAny.(string); ok {
		lang = l
	}
	localizer := goi18n.NewLocalizer(i18n.Bundle, lang)
	msg, _ := localizer.Localize(&goi18n.LocalizeConfig{
		MessageID: strconv.Itoa(code),
	})
	c.JSON(200, Resp{
		Code: code,
		Msg:  msg,
		Data: data,
	})
}

func Ok(c *gin.Context, data interface{}) {
	JSON(c, 0, data)
}

func Fail(c *gin.Context, code int, data ...interface{}) {
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}
	JSON(c, code, d)
}
