package controller

import (
	"net/http"

	"github.com/ix-pay/ixpay-pro/internal/domain/model"
	"github.com/ix-pay/ixpay-pro/internal/infrastructure/logger"

	"github.com/gin-gonic/gin"
)

// UserController 用户控制器
// @Summary 用户相关API
// @Description 提供用户注册、登录、信息查询等功能
// @Tags 用户管理
// @Router /api/v1/user [get]
type UserController struct {
	service model.UserService
	log     logger.Logger
}

// NewUserController 创建用户控制器
func NewUserController(service model.UserService, log logger.Logger) *UserController {
	return &UserController{
		service: service,
		log:     log,
	}
}

// RegisterRequest 注册请求参数
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=30"`
	Email    string `json:"email" binding:"required,email"`
}

// LoginRequest 登录请求参数
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// WechatLoginRequest 微信登录请求参数
type WechatLoginRequest struct {
	Code string `json:"code" binding:"required"`
}

// RefreshTokenRequest 刷新令牌请求参数
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
	Status   int    `json:"status"`
}

// Register 用户注册
// @Summary 用户注册
// @Description 创建新用户账户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "注册请求参数"
// @Success 201 {object} map[string]UserInfoResponse "注册成功"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/auth/register [post]
func (c *UserController) Register(ctx *gin.Context) {
	var req RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.Register(req.Username, req.Password, req.Email)
	if err != nil {
		c.log.Error("User registration failed", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	ctx.JSON(http.StatusCreated, gin.H{"data": response})
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户通过账号密码登录系统
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param login body LoginRequest true "登录请求参数"
// @Success 200 {object} map[string]interface{} "登录成功，包含用户信息和令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /api/v1/auth/login [post]
func (c *UserController) Login(ctx *gin.Context) {
	var req LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, refreshToken, err := c.service.Login(req.Username, req.Password)
	if err != nil {
		c.log.Error("User login failed", "error", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":          response,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// WechatLogin 微信登录
// @Summary 微信登录
// @Description 用户通过微信code进行登录
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param wechat body WechatLoginRequest true "微信登录请求参数"
// @Success 200 {object} map[string]interface{} "登录成功，包含用户信息和令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /api/v1/auth/wechat-login [post]
func (c *UserController) WechatLogin(ctx *gin.Context) {
	var req WechatLoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, accessToken, refreshToken, err := c.service.WechatLogin(req.Code)
	if err != nil {
		c.log.Error("WeChat login failed", "error", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response := UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":          response,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// RefreshToken 刷新令牌
// @Summary 刷新令牌
// @Description 使用refresh token获取新的access token和refresh token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param refresh body RefreshTokenRequest true "刷新令牌请求参数"
// @Success 200 {object} map[string]string "刷新成功，包含新的令牌"
// @Failure 400 {object} map[string]string "请求参数错误"
// @Failure 401 {object} map[string]string "认证失败"
// @Router /api/v1/auth/refresh-token [post]
func (c *UserController) RefreshToken(ctx *gin.Context) {
	var req RefreshTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, err := c.service.RefreshToken(req.RefreshToken)
	if err != nil {
		c.log.Error("Token refresh failed", "error", err)
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
}

// GetUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]UserInfoResponse "用户信息"
// @Failure 401 {object} map[string]string "未授权"
// @Failure 500 {object} map[string]string "服务器内部错误"
// @Router /api/v1/user/info [get]
func (c *UserController) GetUserInfo(ctx *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := c.service.GetUserInfo(userID.(uint))
	if err != nil {
		c.log.Error("Failed to get user info", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := UserInfoResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Email:    user.Email,
		Phone:    user.Phone,
		Avatar:   user.Avatar,
		Role:     user.Role,
		Status:   user.Status,
	}

	ctx.JSON(http.StatusOK, gin.H{"data": response})
}
