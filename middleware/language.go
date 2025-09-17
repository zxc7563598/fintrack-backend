package middleware

import (
	"github.com/gin-gonic/gin"
)

// 语言中间件
func LanguageMiddleware(defaultLang string) gin.HandlerFunc {
	return func(c *gin.Context) {
		lang := c.GetHeader("Accept-Language")
		if lang == "" {
			lang = defaultLang
		}
		c.Set("lang", lang)
		c.Next()
	}
}
