package utils

import (
	"errors"
	"time"

	"line-management/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTClaims JWT声明结构
type JWTClaims struct {
	UserID         uint   `json:"user_id"`
	Username       string `json:"username"`
	Role           string `json:"role"`
	ActivationCode string `json:"activation_code,omitempty"` // 子账号登录时使用
	GroupID        uint   `json:"group_id,omitempty"`        // 子账号登录时使用
	jwt.RegisteredClaims
}

// GenerateToken 生成JWT Token
func GenerateToken(userID uint, username, role string) (string, error) {
	cfg := config.GlobalConfig.JWT
	expireTime := time.Now().Add(time.Duration(cfg.ExpireHour) * time.Hour)

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "line-management",
			Subject:   username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GenerateSubAccountToken 生成子账号JWT Token
func GenerateSubAccountToken(groupID uint, activationCode string) (string, error) {
	cfg := config.GlobalConfig.JWT
	expireTime := time.Now().Add(time.Duration(cfg.ExpireHour) * time.Hour)

	claims := JWTClaims{
		GroupID:        groupID,
		ActivationCode: activationCode,
		Role:           "subaccount",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "line-management",
			Subject:   activationCode,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析JWT Token
func ParseToken(tokenString string) (*JWTClaims, error) {
	cfg := config.GlobalConfig.JWT

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return []byte(cfg.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// RefreshToken 刷新Token
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果Token已过期超过1小时，不允许刷新
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) < -time.Hour {
		return "", errors.New("Token已过期，无法刷新")
	}

	// 根据角色生成新Token
	if claims.Role == "subaccount" {
		return GenerateSubAccountToken(claims.GroupID, claims.ActivationCode)
	}

	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}

