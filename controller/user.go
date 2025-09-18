package controller

import (
	"encoding/base64"
	"errors"

	"github.com/zxc7563598/fintrack-backend/config"
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
