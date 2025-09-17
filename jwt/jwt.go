package jwt

import (
	"fmt"
	"time"

	"github.com/zxc7563598/fintrack-backend/config"
	"github.com/zxc7563598/fintrack-backend/model"

	"github.com/golang-jwt/jwt/v5"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Claims struct {
	UserID uint `json:"user_id"`
	Role   Role `json:"role"`
	jwt.RegisteredClaims
}

// 登录生成 Access + Refresh Token
func GenerateTokens(userID uint, role Role) (accessToken string, refreshToken string, err error) {
	now := time.Now()
	// Access Token
	accessClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.Cfg.JWT.AccessTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = atoken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return
	}
	// Refresh Token
	refreshClaims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.Cfg.JWT.RefreshTokenExp)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = rtoken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return
	}
	// 保存 Refresh Token 到 Redis
	config.DB.Create(&model.Token{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    now.Add(config.Cfg.JWT.RefreshTokenExp),
	})
	return
}

// 解析 Access Token
func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWT.Secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}

// 刷新 Access Token，并顺带延长 Refresh Token 的有效期
func RefreshAccessToken(refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	// 验证 Refresh Token
	token, err := jwt.ParseWithClaims(refreshToken, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Cfg.JWT.Secret), nil
	})
	if err != nil {
		return "", "", err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", "", fmt.Errorf("invalid refresh token")
	}
	// 检查 Redis 是否存在
	var tokenModel model.Token
	result := config.DB.Where("refresh_token = ? AND expires_at > ?", refreshToken, time.Now()).First(&tokenModel)
	if result.Error != nil {
		return "", "", fmt.Errorf("refresh token expired")
	}
	// --- 生成新的 Access Token ---
	now := time.Now()
	accessClaims := Claims{
		UserID: claims.UserID,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.Cfg.JWT.AccessTokenExp)), // 短期
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	newAccessToken, err = atoken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}
	// --- 刷新 Refresh Token（延长时间/重新生成） ---
	refreshClaims := Claims{
		UserID: claims.UserID,
		Role:   claims.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.Cfg.JWT.RefreshTokenExp)), // 延长
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	rtoken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	newRefreshToken, err = rtoken.SignedString([]byte(config.Cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}
	// 删除旧的 refresh_token
	config.DB.Delete(&tokenModel)
	// 存储新的 refresh_token
	config.DB.Create(&model.Token{
		UserID:       claims.UserID,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    now.Add(config.Cfg.JWT.RefreshTokenExp),
	})
	return newAccessToken, newRefreshToken, nil
}

// 注销 Refresh Token
func RevokeRefreshToken(refreshToken string) error {
	return config.DB.Where("refresh_token = ?", refreshToken).Delete(&model.Token{}).Error
}
