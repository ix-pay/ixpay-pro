package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
)

// ContextKey 定义上下文键类型
type ContextKey string

// 上下文键常量
const (
	UserIDKey      ContextKey = "userID"
	ClaimsKey      ContextKey = "claims"
	RoleKey        ContextKey = "role"
	UsernameKey    ContextKey = "userName"
	LoginTypeKey   ContextKey = "loginType"
	UserButtonsKey ContextKey = "userButtons"
)

// JWTAuth 提供JWT认证功能
type JWTAuth struct {
	secretKey          []byte
	accessTokenExpire  time.Duration
	refreshTokenExpire time.Duration
	log                logger.Logger
}

// Claims 自定义 JWT 声明
type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"userName"`
	Nickname  string `json:"nickname"`
	Role      string `json:"role"`
	LoginType string `json:"login_type"`
	jwt.RegisteredClaims
}

// NewJWTAuth 创建新的 JWT 认证实例
func SetupJWTAuth(cfg *config.Config, log logger.Logger) (*JWTAuth, error) {
	// 解析过期时间
	accessTokenExpire, err := time.ParseDuration(cfg.JWT.AccessTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("访问令牌过期时间格式无效：%w", err)
	}

	refreshTokenExpire, err := time.ParseDuration(cfg.JWT.RefreshTokenExpire)
	if err != nil {
		return nil, fmt.Errorf("刷新令牌过期时间格式无效：%w", err)
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
func (j *JWTAuth) GenerateToken(userID string, userName string, nickname string, role string, loginType string) (string, string, time.Time, time.Time, error) {
	// 生成访问令牌
	accessToken, accessExpire, err := j.generateAccessToken(userID, userName, nickname, role, loginType)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	// 生成刷新令牌
	refreshToken, refreshExpire, err := j.generateRefreshToken(userID, userName, nickname, role, loginType)
	if err != nil {
		return "", "", time.Time{}, time.Time{}, err
	}

	return accessToken, refreshToken, accessExpire, refreshExpire, nil
}

// generateAccessToken 生成访问令牌
func (j *JWTAuth) generateAccessToken(userID string, userName string, nickname string, role string, loginType string) (string, time.Time, error) {
	expirationTime := time.Now().Add(j.accessTokenExpire)

	claims := &Claims{
		UserID:    userID,
		Username:  userName,
		Nickname:  nickname,
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
		j.log.Error("生成访问令牌失败", "error", err)
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// generateRefreshToken 生成刷新令牌
func (j *JWTAuth) generateRefreshToken(userID string, userName string, nickname string, role string, loginType string) (string, time.Time, error) {
	expirationTime := time.Now().Add(j.refreshTokenExpire)

	claims := &Claims{
		UserID:    userID,
		Username:  userName,
		Nickname:  nickname,
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
		j.log.Error("生成刷新令牌失败", "error", err)
		return "", time.Time{}, err
	}

	return tokenString, expirationTime, nil
}

// ParseToken 解析令牌
func (j *JWTAuth) ParseToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法：%v", token.Header["alg"])
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("令牌无效")
	}

	return claims, nil
}

// RefreshToken 刷新访问令牌
func (j *JWTAuth) RefreshToken(refreshToken string) (string, string, time.Time, time.Time, error) {
	claims, err := j.ParseToken(refreshToken)
	if err != nil {
		j.log.Error("解析刷新令牌失败", "error", err)
		return "", "", time.Time{}, time.Time{}, err
	}

	// 检查令牌是否接近过期
	if time.Until(claims.ExpiresAt.Time) > j.accessTokenExpire {
		// 令牌还有很长时间才会过期，只生成新的访问令牌
		accessToken, accessExpire, err := j.generateAccessToken(claims.UserID, claims.Username, claims.Nickname, claims.Role, claims.LoginType)
		if err != nil {
			return "", "", time.Time{}, time.Time{}, err
		}
		return accessToken, refreshToken, accessExpire, claims.ExpiresAt.Time, nil
	}

	// 令牌即将过期，生成新的访问令牌和刷新令牌
	return j.GenerateToken(claims.UserID, claims.Username, claims.Nickname, claims.Role, claims.LoginType)
}

// GetContextWithUserID 将用户ID添加到上下文
func GetContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// GetUserIDFromContext 从上下文获取用户ID
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetContextWithClaims 将JWT声明添加到上下文
func GetContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// GetClaimsFromContext 从上下文获取JWT声明
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*Claims)
	return claims, ok
}
