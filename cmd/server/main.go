// Package main
// Author: HHaou
// Created: 2024-01-20
// Description: 分布式游戏排行榜系统主程序入口

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ranking/internal/app"
	"ranking/internal/config"
	"ranking/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	// Version 版本信息
	Version = "1.0.0"
	// BuildTime 构建时间
	BuildTime = "unknown"
	// GitCommit Git提交哈希
	GitCommit = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ranking-server",
		Short: "分布式游戏排行榜系统",
		Long:  "高性能、高并发的分布式游戏排行榜微服务系统",
		Run:   runServer,
	}

	// 添加版本命令
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", Version)
			fmt.Printf("Build Time: %s\n", BuildTime)
			fmt.Printf("Git Commit: %s\n", GitCommit)
		},
	}

	rootCmd.AddCommand(versionCmd)

	// 添加配置文件参数
	rootCmd.PersistentFlags().StringP("config", "c", "configs/config.yaml", "配置文件路径")
	rootCmd.PersistentFlags().StringP("env", "e", "development", "运行环境 (development, testing, production)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// runServer 启动服务器
func runServer(cmd *cobra.Command, args []string) {
	// 获取配置文件路径
	configPath, _ := cmd.Flags().GetString("config")
	env, _ := cmd.Flags().GetString("env")

	// 加载配置
	cfg, err := config.Load(configPath, env)
	if err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化日志
	logger, err := logger.New(cfg.Log)
	if err != nil {
		log.Fatalf("初始化日志失败: %v", err)
	}
	defer logger.Sync()

	logger.Info("启动排行榜服务",
		"version", Version,
		"build_time", BuildTime,
		"git_commit", GitCommit,
		"env", env,
	)

	// 创建应用实例
	app, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatal("创建应用实例失败", "error", err)
	}

	// 启动服务
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := app.Start(ctx); err != nil {
			logger.Error("服务启动失败", "error", err)
			cancel()
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logger.Info("接收到信号，开始优雅关闭", "signal", sig)
	case <-ctx.Done():
		logger.Info("应用上下文取消，开始关闭")
	}

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := app.Stop(shutdownCtx); err != nil {
		logger.Error("服务关闭失败", "error", err)
	} else {
		logger.Info("服务已优雅关闭")
	}
}