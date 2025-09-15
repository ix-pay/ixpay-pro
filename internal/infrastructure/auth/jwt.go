package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/golang-jwt/jwt/v5"
)

// JWTAuth 提供JWT认证功能
type JWTAuth struct {
	secretKey          []byte
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
	log                logger.Logger
}

// Claims 自定义JWT声明

type Claims struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	LoginType string `json:"login_type"` // "password" or "wechat"
	jwt.RegisteredClaims
}

// NewJWTAuth 创建新的JWT认证实例
func NewJWTAuth(cfg *config.Config, log logger.Logger) (*JWTAuth, error) {
	// 解析过期时间
	accessTokenExpire, err := time.ParseDuration(cfg.JWT.AccessTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("invalid access token expire duration: %w", err)
	}

	refreshTokenExpire, err := time.ParseDuration(cfg.JWT.RefreshTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token expire duration: %w", err)
	}

	return &JWTAuth{
			secretKey:          []byte(cfg.JWT.SecretKey),
			accessTokenExpire:  accessTokenExpire,
			refreshTokenExpire: refreshTokenExpire,
			log:                log,
		},
		nil
}

// GenerateToken 生成访问令牌和刷新令牌
func (j *JWTAuth) GenerateToken(userID uint, username string, role string, loginType string) (string, string, error) {
	// 生成访问令牌
	accessToken, err := j.generateAccessToken(userID, username, role, loginType)
	if err != nil {
		return "", "", err
	}

	// 生成刷新令牌
	refreshToken, err := j.generateRefreshToken(userID, username, role, loginType)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken 生成访问令牌
func (j *JWTAuth) generateAccessToken(userID uint, username string, role string, loginType string) (string, error) {
	expirationTime := time.Now().Add(j.accessTokenExpire)

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		LoginType: loginType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		j.log.Error("Failed to generate access token", "error", err)
		return "", err
	}

	return tokenString, nil
}

// generateRefreshToken 生成刷新令牌
func (j *JWTAuth) generateRefreshToken(userID uint, username string, role string, loginType string) (string, error) {
	expirationTime := time.Now().Add(j.refreshTokenExpire)

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Role:      role,
		LoginType: loginType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secretKey)
	if err != nil {
		j.log.Error("Failed to generate refresh token", "error", err)
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析令牌
func (j *JWTAuth) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// RefreshToken 刷新访问令牌
func (j *JWTAuth) RefreshToken(refreshToken string) (string, string, error) {
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		j.log.Error("Failed to parse refresh token", "error", err)
		return "", "", err
	}

	// 检查令牌是否接近过期
	if time.Until(claims.ExpiresAt.Time) > j.accessTokenExpire {
		// 令牌还有很长时间才会过期，只生成新的访问令牌
		accessToken, err := j.generateAccessToken(claims.UserID, claims.Username, claims.Role, claims.LoginType)
		if err != nil {
			return "", "", err
		}
		return accessToken, refreshToken, nil
	}

	// 令牌即将过期，生成新的访问令牌和刷新令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.Role, claims.LoginType)
}

// GetContextWithUserID 将用户ID添加到上下文
func GetContextWithUserID(ctx context.Context, userID uint) context.Context {
	return context.WithValue(ctx, "userID", userID)
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value("userID").(uint)
	return userID, ok
}

// GetContextWithClaims 将JWT声明添加到上下文
func GetContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, "claims", claims)
}

// GetClaimsFromContext 从上下文获取JWT声明
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value("claims").(*Claims)
	return claims, ok
}
