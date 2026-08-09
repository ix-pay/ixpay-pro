package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ix-pay/ixpay-pro/internal/config"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/observability/logger"
	a "github.com/ix-pay/ixpay-pro/internal/infrastructure/security/auth"
	"github.com/stretchr/testify/assert"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, fields ...interface{})       {}
func (m *mockLogger) Info(msg string, fields ...interface{})        {}
func (m *mockLogger) Warn(msg string, fields ...interface{})        {}
func (m *mockLogger) Error(msg string, fields ...interface{})       {}
func (m *mockLogger) Fatal(msg string, fields ...interface{})       {}
func (m *mockLogger) With(fields ...interface{}) logger.Logger      { return m }
func (m *mockLogger) WithContext(ctx context.Context) logger.Logger { return m }
func (m *mockLogger) Sync() error                                   { return nil }

func newTestJWTAuth(t *testing.T) *a.JWTAuth {
	log := &mockLogger{}

	jwtAuth, err := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "test_secret_key_for_jwt",
			AccessTokenExpire:  "15m",
			RefreshTokenExpire: "168h",
		},
	}, log)
	assert.NoError(t, err)
	assert.NotNil(t, jwtAuth)

	return jwtAuth
}

func TestSetupJWTAuth(t *testing.T) {
	log := &mockLogger{}

	jwtAuth, err := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "test_secret",
			AccessTokenExpire:  "15m",
			RefreshTokenExpire: "168h",
		},
	}, log)

	assert.NoError(t, err)
	assert.NotNil(t, jwtAuth)
}

func TestSetupJWTAuth_InvalidAccessTokenExpire(t *testing.T) {
	log := &mockLogger{}

	_, err := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "test_secret",
			AccessTokenExpire:  "invalid",
			RefreshTokenExpire: "7d",
		},
	}, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "访问令牌过期时间格式无效")
}

func TestSetupJWTAuth_InvalidRefreshTokenExpire(t *testing.T) {
	log := &mockLogger{}

	_, err := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "test_secret",
			AccessTokenExpire:  "15m",
			RefreshTokenExpire: "invalid",
		},
	}, log)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "刷新令牌过期时间格式无效")
}

func TestGenerateToken(t *testing.T) {
	jwtAuth := newTestJWTAuth(t)

	accessToken, refreshToken, accessExpire, refreshExpire, err := jwtAuth.GenerateToken(
		"123", "testuser", "测试用户", "user", "password", "",
	)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.False(t, accessExpire.IsZero())
	assert.False(t, refreshExpire.IsZero())
	assert.True(t, accessExpire.Before(refreshExpire))
}

func TestParseToken_Valid(t *testing.T) {
	jwtAuth := newTestJWTAuth(t)

	accessToken, _, _, _, err := jwtAuth.GenerateToken(
		"123", "testuser", "测试用户", "admin", "password", "",
	)
	assert.NoError(t, err)

	claims, err := jwtAuth.ParseToken(accessToken)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "123", claims.UserID)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "测试用户", claims.Nickname)
	assert.Equal(t, "admin", claims.Role)
	assert.Equal(t, "password", claims.LoginType)
}

func TestParseToken_InvalidSignature(t *testing.T) {
	log := &mockLogger{}

	jwtAuth1, _ := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "secret_key_1",
			AccessTokenExpire:  "15m",
			RefreshTokenExpire: "168h",
		},
	}, log)

	accessToken, _, _, _, err := jwtAuth1.GenerateToken(
		"123", "testuser", "测试用户", "user", "password", "",
	)
	assert.NoError(t, err)

	jwtAuth2, _ := a.SetupJWTAuth(&config.Config{
		JWT: config.JWTConfig{
			SecretKey:          "different_secret_key_2",
			AccessTokenExpire:  "15m",
			RefreshTokenExpire: "168h",
		},
	}, log)

	_, err = jwtAuth2.ParseToken(accessToken)
	assert.Error(t, err)
}

func TestParseToken_InvalidToken(t *testing.T) {
	jwtAuth := newTestJWTAuth(t)

	_, err := jwtAuth.ParseToken("invalid.token.string")
	assert.Error(t, err)
}

func TestParseToken_WrongAlgorithm(t *testing.T) {
	jwtAuth := newTestJWTAuth(t)

	token := jwt.NewWithClaims(jwt.SigningMethodNone, &a.Claims{
		UserID:   "123",
		Username: "testuser",
	})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := jwtAuth.ParseToken(tokenString)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "意外的签名方法")
}

func TestRefreshToken(t *testing.T) {
	jwtAuth := newTestJWTAuth(t)

	_, refreshToken, _, _, err := jwtAuth.GenerateToken(
		"123", "testuser", "测试用户", "user", "password", "",
	)
	assert.NoError(t, err)

	newAccessToken, newRefreshToken, newAccessExpire, newRefreshExpire, err := jwtAuth.RefreshToken(refreshToken)
	assert.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)
	assert.NotEmpty(t, newRefreshToken)
	assert.False(t, newAccessExpire.IsZero())
	assert.False(t, newRefreshExpire.IsZero())
}

func TestContextFunctions(t *testing.T) {
	ctx := context.Background()

	ctx = a.GetContextWithUserID(ctx, "123")
	userID, ok := a.GetUserIDFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, "123", userID)
}

func TestContextWithClaims(t *testing.T) {
	ctx := context.Background()

	claims := &a.Claims{
		UserID:   "456",
		Username: "claimuser",
	}

	ctx = a.GetContextWithClaims(ctx, claims)
	retrievedClaims, ok := a.GetClaimsFromContext(ctx)
	assert.True(t, ok)
	assert.NotNil(t, retrievedClaims)
	assert.Equal(t, "456", retrievedClaims.UserID)
	assert.Equal(t, "claimuser", retrievedClaims.Username)
}

func TestGetUserIDFromContext_Empty(t *testing.T) {
	ctx := context.Background()

	_, ok := a.GetUserIDFromContext(ctx)
	assert.False(t, ok)
}

func TestGetClaimsFromContext_Empty(t *testing.T) {
	ctx := context.Background()

	_, ok := a.GetClaimsFromContext(ctx)
	assert.False(t, ok)
}

func TestClaims_Struct(t *testing.T) {
	claims := &a.Claims{
		UserID:    "789",
		Username:  "structuser",
		Nickname:  "昵称",
		Role:      "editor",
		LoginType: "wechat",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	assert.Equal(t, "789", claims.UserID)
	assert.Equal(t, "structuser", claims.Username)
	assert.Equal(t, "昵称", claims.Nickname)
	assert.Equal(t, "editor", claims.Role)
	assert.Equal(t, "wechat", claims.LoginType)
}

func TestContextKeyConstants(t *testing.T) {
	assert.Equal(t, a.ContextKey("userID"), a.UserIDKey)
	assert.Equal(t, a.ContextKey("claims"), a.ClaimsKey)
	assert.Equal(t, a.ContextKey("role"), a.RoleKey)
	assert.Equal(t, a.ContextKey("userName"), a.UsernameKey)
	assert.Equal(t, a.ContextKey("loginType"), a.LoginTypeKey)
	assert.Equal(t, a.ContextKey("userButtons"), a.UserButtonsKey)
}
