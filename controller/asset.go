package controller

import (
	"github.com/zxc7563598/fintrack-backend/utils/response"

	"github.com/gin-gonic/gin"
)

func AssetOverviewHandler(c *gin.Context) {
	user_id, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, 300001)
		return
	}
	// 返回数据
	response.Ok(c, gin.H{
		"user_id": user_id,
	})
}
