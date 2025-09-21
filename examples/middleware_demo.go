// Package examples
// 中间件测试示例

package main

import (
	"fmt"
	"net/http"
	"time"

	"ranking/internal/middleware"
	"ranking/pkg/logger"

	netcore "github.com/phuhao00/netcore-go"
	netcorehttp "github.com/phuhao00/netcore-go/pkg/http"
)

// TestHandler 测试处理器
type TestHandler struct{}

// ServeHTTP 实现HTTPHandler接口
func (h *TestHandler) ServeHTTP(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse) {
	ctx.JSON(resp, http.StatusOK, map[string]interface{}{
		"message": "Hello, World!",
		"path":    ctx.Path(),
		"method":  ctx.Method(),
		"time":    time.Now().Format(time.RFC3339),
	})
}

func main() {
	// 创建日志器
	log, err := logger.NewDevelopment()
	if err != nil {
		fmt.Printf("创建日志器失败: %v\n", err)
		return
	}

	// 创建HTTP服务器
	server := netcore.NewHTTPServer()

	// 添加中间件
	server.Use(middleware.NewLoggerMiddleware(log))
	server.Use(middleware.NewRecoveryMiddleware(log))
	server.Use(middleware.NewCORSMiddleware())
	server.Use(middleware.NewRequestIDMiddleware())
	server.Use(middleware.NewSecurityMiddleware())
	server.Use(middleware.NewRateLimitMiddleware(10)) // 每秒10个请求

	// 添加测试路由
	testHandler := &TestHandler{}
	server.GET("/test", testHandler.ServeHTTP)
	server.POST("/test", testHandler.ServeHTTP)

	// 添加健康检查路由
	server.GET("/health", func(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse) {
		ctx.JSON(resp, http.StatusOK, map[string]interface{}{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 添加限流测试路由
	server.GET("/rate-limit-test", func(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse) {
		ctx.JSON(resp, http.StatusOK, map[string]interface{}{
			"message": "Rate limit test - try calling this endpoint rapidly",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// 添加panic测试路由
	server.GET("/panic-test", func(ctx *netcorehttp.HTTPContext, resp *netcorehttp.HTTPResponse) {
		panic("This is a test panic!")
	})

	fmt.Println("中间件测试服务器启动中...")
	fmt.Println("测试端点:")
	fmt.Println("  GET  /health          - 健康检查")
	fmt.Println("  GET  /test            - 基本测试")
	fmt.Println("  POST /test            - POST测试")
	fmt.Println("  GET  /rate-limit-test - 限流测试")
	fmt.Println("  GET  /panic-test      - 恢复中间件测试")
	fmt.Println("")
	fmt.Println("服务器将在 :8080 端口启动")

	// 启动服务器
	if err := server.Start(); err != nil {
		log.Error("服务器启动失败", "error", err)
	}
}