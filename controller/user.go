package controller

import (
	"encoding/base64"
	"errors"

	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/dto"
	"github.com/zxc7563598/fintrack-backend/jwt"
	"github.com/zxc7563598/fintrack-backend/model"
	"github.com/zxc7563598/fintrack-backend/utils/helpers"
	"github.com/zxc7563598/fintrack-backend/utils/response"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

// 用户注册请求体
type LoginRegisterRequest struct {
	Name     string `json:"name" binding:"required"`     // 用户昵称
	Email    string `json:"email" binding:"required"`    // 用户邮箱
	Password string `json:"password" binding:"required"` // 用户密码
}

// 用户注册接口
func LoginRegisterHandler(c *gin.Context) {
	// 获取参数
	req := c.MustGet("payload").(LoginRegisterRequest)
	// 获取用户信息
	var user model.User
	result := config.DB.Where("email = ?", req.Email).First(&user)
	if result.Error == nil {
		response.Fail(c, 100004)
		return
	}
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		response.Fail(c, 100001)
		return
	}
	// 创建用户
	salt, err := helpers.GenerateSalt(16)
	if err != nil {
		panic(err)
	}
	hash := helpers.HashPassword(req.Password, salt)
	created := config.DB.Create(&model.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
		Salt:     base64.RawStdEncoding.EncodeToString(salt),
	})
	if created.Error != nil {
		response.Fail(c, 100006)
		return
	}
	// 返回信息
	response.Ok(c, gin.H{})
}

// 用户登录请求体
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`    // 用户邮箱
	Password string `json:"password" binding:"required"` // 用户密码
}

// 用户登录接口
func LoginHandler(c *gin.Context) {
	// 获取参数
	req := c.MustGet("payload").(LoginRequest)
	// 获取用户信息
	var user model.User
	result := config.DB.Where("email = ?", req.Email).First(&user)
	if result.Error != nil {
		response.Fail(c, 100002)
		return
	}
	is_password := user.VerifyPassword(req.Password)
	if !is_password {
		response.Fail(c, 100003)
		return
	}
	accessToken, refreshToken, err := jwt.GenerateTokens(user.ID, jwt.RoleUser)
	if err != nil {
		response.Fail(c, 200002)
		return
	}
	// 返回信息
	response.Ok(c, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// 刷新token请求体
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"` // refresh token
}

// 刷新token接口
func RefreshTokenHandler(c *gin.Context) {
	// 获取参数
	req := c.MustGet("payload").(RefreshTokenRequest)
	// 调用刷新逻辑
	newAccessToken, newRefreshToken, err := jwt.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		response.Fail(c, 100005)
		return
	}
	// 返回信息
	response.Ok(c, gin.H{
		"access_token":  newAccessToken,
		"refresh_token": newRefreshToken,
	})
}

// 获取用户绑定邮箱接口
func GetUserEmailsHandler(c *gin.Context) {
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
	// 获取数据
	var list []dto.UserMailboxListItem
	db := config.DB.Model(&model.UserMailbox{}).Where("user_id = ?", userID)
	if err := db.Find(&list).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 返回信息
	response.Ok(c, gin.H{
		"list": list,
	})
}

// 存储用户绑定邮箱请求体
type StoreUserEmailRequest struct {
	ID       uint   `json:"id"`                           // ID，修改透传，添加为0
	Email    string `json:"email" binding:"required"`     // 邮箱
	AuthCode string `json:"auth_code" binding:"required"` // 授权码
	IMAP     string `json:"imap" binding:"required"`      // IMAP地址
	Remark   string `json:"remark"`                       // 备注
}

// 存储用户绑定邮箱接口
func StoreUserEmailHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(StoreUserEmailRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 存储数据
	userMailbox := model.UserMailbox{
		UserID:   userID,
		Email:    req.Email,
		AuthCode: req.AuthCode,
		IMAP:     req.IMAP,
		Remark:   req.Remark,
	}
	if req.ID > 0 {
		// 修改
		err := config.DB.Model(&model.UserMailbox{}).
			Where("id = ? AND user_id = ?", req.ID, userID).
			Updates(userMailbox).Error
		if err != nil {
			response.Fail(c, 100013)
			return
		}
	} else {
		// 新增
		if err := config.DB.Create(&userMailbox).Error; err != nil {
			response.Fail(c, 100013)
			return
		}
	}
	// 返回成功
	response.Ok(c, gin.H{})
}

// 删除用户绑定邮箱请求体
type DeleteUserEmailRequest struct {
	ID uint `json:"id" binding:"required"` // ID
}

// 删除用户绑定邮箱接口
func DeleteUserEmailHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(DeleteUserEmailRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 删除数据
	if err := config.DB.Where("id = ? and user_id = ?", req.ID, userID).Delete(&model.UserMailbox{}).Error; err != nil {
		response.Fail(c, 100014)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{})
}

// 获取用户DeepseekApiKey接口
func GetDeepseekApiKeyHandler(c *gin.Context) {
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
	// 获取数据
	var key string
	if err := config.DB.Model(&model.User{}).
		Distinct("deepseek_api_key").
		Where("id = ?", userID).
		Pluck("deepseek_api_key", &key).Error; err != nil {
		response.Fail(c, 100001)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{
		"key": key,
	})
}

// 存储用户DeepseekApiKey请求体
type StoreDeepseekApiKeyRequest struct {
	Key string `json:"key" binding:"required"`
}

// 存储用户DeepseekApiKey接口
func StoreDeepseekApiKeyHandler(c *gin.Context) {
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
	req, ok := c.MustGet("payload").(StoreDeepseekApiKeyRequest)
	if !ok {
		response.Fail(c, 100010)
		return
	}
	// 存储数据
	user := model.User{
		DeepseekApiKey: req.Key,
	}
	err := config.DB.Model(&model.User{}).
		Where("id = ?", userID).
		Updates(user).Error
	if err != nil {
		response.Fail(c, 100013)
		return
	}
	// 返回成功
	response.Ok(c, gin.H{})
}
