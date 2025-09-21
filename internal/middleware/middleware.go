// Package middleware
// Author: HHaou
// Created: 2024-01-20
// Description: HTTP中间件实现，基于netcore-go框架

package middleware

import (
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"

	"ranking/pkg/logger"

	netcorehttp "github.com/phuhao00/netcore-go/pkg/http"
	"golang.org/x/time/rate"
)

// LoggerMiddleware 日志中间件
type LoggerMiddleware struct {
	logger logger.Logger
}

// NewLoggerMiddleware 创建日志中间件
func NewLoggerMiddleware(log logger.Logger) *LoggerMiddleware {
	return &LoggerMiddleware{logger: log}
}

// Handle 实现HTTPMiddleware接口
func (m *LoggerMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	start := time.Now()
	path := ctx.Path()
	method := ctx.Method()
	clientIP := getClientIP(ctx)

	// 记录请求开始
	m.logger.Info("HTTP请求开始",
		"method", method,
		"path", path,
		"client_ip", clientIP,
	)

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)

	// 记录请求完成
	latency := time.Since(start)
	m.logger.Info("HTTP请求完成",
		"method", method,
		"path", path,
		"latency", latency,
		"client_ip", clientIP,
	)
}

// getClientIP 获取客户端IP地址
func getClientIP(ctx *netcorehttp.HTTPContext) string {
	// 检查X-Forwarded-For头
	if xff := ctx.Header("X-Forwarded-For"); xff != "" {
		return xff
	}

	// 检查X-Real-IP头
	if xri := ctx.Header("X-Real-IP"); xri != "" {
		return xri
	}

	// 默认返回空字符串，netcore-go会处理
	return ""
}

// RecoveryMiddleware 恢复中间件
type RecoveryMiddleware struct {
	logger logger.Logger
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(log logger.Logger) *RecoveryMiddleware {
	return &RecoveryMiddleware{logger: log}
}

// Handle 实现HTTPMiddleware接口
func (m *RecoveryMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	defer func() {
		if err := recover(); err != nil {
			// 记录panic信息
			m.logger.Error("HTTP请求panic恢复",
				"error", err,
				"path", ctx.Path(),
				"method", ctx.Method(),
				"client_ip", getClientIP(ctx),
				"stack", string(debug.Stack()),
			)

			// 返回500错误
			ctx.Error(resp, http.StatusInternalServerError, "内部服务器错误")
		}
	}()

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}

// CORSMiddleware 跨域中间件
type CORSMiddleware struct{}

// NewCORSMiddleware 创建CORS中间件
func NewCORSMiddleware() *CORSMiddleware {
	return &CORSMiddleware{}
}

// Handle 实现HTTPMiddleware接口
func (m *CORSMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	// 设置CORS头
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers["Access-Control-Allow-Origin"] = "*"
	resp.Headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
	resp.Headers["Access-Control-Allow-Headers"] = "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization"
	resp.Headers["Access-Control-Expose-Headers"] = "Content-Length"
	resp.Headers["Access-Control-Allow-Credentials"] = "true"

	if ctx.Method() == "OPTIONS" {
		resp.StatusCode = http.StatusNoContent
		return
	}

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}

// RateLimiter 限流器
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
}

// GetLimiter 获取客户端限流器
func (rl *RateLimiter) GetLimiter(clientIP string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[clientIP]
	if !exists {
		limiter = rate.NewLimiter(rl.limit, rl.burst)
		rl.limiters[clientIP] = limiter
	}

	return limiter
}

// RateLimitMiddleware 限流中间件
type RateLimitMiddleware struct {
	rateLimiter *RateLimiter
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(rps int) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: NewRateLimiter(rps, rps*2),
	}
}

// Handle 实现HTTPMiddleware接口
func (m *RateLimitMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	clientIP := getClientIP(ctx)
	if clientIP == "" {
		clientIP = "unknown"
	}
	limiter := m.rateLimiter.GetLimiter(clientIP)

	if !limiter.Allow() {
		ctx.JSON(resp, http.StatusTooManyRequests, map[string]interface{}{
			"code":    429,
			"message": "请求过于频繁，请稍后再试",
		})
		return
	}

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}

// RequestIDMiddleware 请求ID中间件
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware 创建请求ID中间件
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// Handle 实现HTTPMiddleware接口
func (m *RequestIDMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	requestID := ctx.Header("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	ctx.Set("request_id", requestID)
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers["X-Request-ID"] = requestID

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return strconv.FormatInt(time.Now().UnixNano(), 36)
}

// AdminAuthMiddleware 管理员认证中间件
type AdminAuthMiddleware struct{}

// NewAdminAuthMiddleware 创建管理员认证中间件
func NewAdminAuthMiddleware() *AdminAuthMiddleware {
	return &AdminAuthMiddleware{}
}

// Handle 实现HTTPMiddleware接口
func (m *AdminAuthMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	token := ctx.Header("Authorization")
	if token == "" {
		ctx.JSON(resp, http.StatusUnauthorized, map[string]interface{}{
			"code":    401,
			"message": "缺少管理员认证令牌",
		})
		return
	}

	// TODO: 验证管理员JWT令牌
	// 这里应该实现管理员JWT令牌验证逻辑

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct{}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{}
}

// Handle 实现HTTPMiddleware接口
func (m *SecurityMiddleware) Handle(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse, next netcorehttp.HTTPHandler) {
	// 设置安全头
	if resp.Headers == nil {
		resp.Headers = make(map[string]string)
	}
	resp.Headers["X-Content-Type-Options"] = "nosniff"
	resp.Headers["X-Frame-Options"] = "DENY"
	resp.Headers["X-XSS-Protection"] = "1; mode=block"
	resp.Headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains"
	resp.Headers["Content-Security-Policy"] = "default-src 'self'"

	// 调用下一个处理器
	next.ServeHTTP(ctx, resp)
}
