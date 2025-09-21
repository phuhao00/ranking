// Package logger
// Author: HHaou
// Created: 2024-01-20
// Description: 日志模块，基于spoor v2.0.1日志库的封装

package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"ranking/internal/config"

	spoor "github.com/phuhao00/spoor/v2"
)

// Logger 日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	Sync() error
	With(fields ...interface{}) Logger
}

// spoorLogger spoor日志实现
type spoorLogger struct {
	logger spoor.Logger
	fields map[string]interface{}
}

// New 创建新的日志实例
func New(cfg config.LogConfig) (Logger, error) {
	// 创建日志目录
	if cfg.Output == "file" && cfg.FilePath != "" {
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	// 创建spoor实例
	if cfg.Output == "file" && cfg.FilePath != "" {
		// 使用文件输出
		fileWriter, err := spoor.NewFile(filepath.Dir(cfg.FilePath), spoor.LevelInfo)
		if err != nil {
			return nil, fmt.Errorf("创建文件日志器失败: %w", err)
		}
		logger := fileWriter
		return &spoorLogger{
			logger: logger,
			fields: make(map[string]interface{}),
		}, nil
	} else {
		// 使用控制台输出
		logger := spoor.NewConsole(spoor.LevelInfo)
		return &spoorLogger{
			logger: logger,
			fields: make(map[string]interface{}),
		}, nil
	}
}

// Debug 调试日志
func (l *spoorLogger) Debug(msg string, fields ...interface{}) {
	l.logWithFields("DEBUG", msg, fields...)
}

// Info 信息日志
func (l *spoorLogger) Info(msg string, fields ...interface{}) {
	l.logWithFields("INFO", msg, fields...)
}

// Warn 警告日志
func (l *spoorLogger) Warn(msg string, fields ...interface{}) {
	l.logWithFields("WARN", msg, fields...)
}

// Error 错误日志
func (l *spoorLogger) Error(msg string, fields ...interface{}) {
	l.logWithFields("ERROR", msg, fields...)
}

// Fatal 致命错误日志
func (l *spoorLogger) Fatal(msg string, fields ...interface{}) {
	l.logWithFields("FATAL", msg, fields...)
	os.Exit(1)
}

// Sync 同步日志
func (l *spoorLogger) Sync() error {
	// spoor可能没有Sync方法，返回nil
	return nil
}

// With 添加字段
func (l *spoorLogger) With(fields ...interface{}) Logger {
	newFields := make(map[string]interface{})
	// 复制现有字段
	for k, v := range l.fields {
		newFields[k] = v
	}
	
	// 添加新字段
	parsedFields := l.parseFields(fields...)
	for k, v := range parsedFields {
		newFields[k] = v
	}

	return &spoorLogger{
		logger: l.logger,
		fields: newFields,
	}
}

// logWithFields 带字段的日志记录
func (l *spoorLogger) logWithFields(level string, msg string, fields ...interface{}) {
	// 合并现有字段和新字段
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	
	parsedFields := l.parseFields(fields...)
	for k, v := range parsedFields {
		allFields[k] = v
	}

	// 使用spoor v2的WithFields方法添加字段，然后记录日志
	logger := l.logger
	if len(allFields) > 0 {
		logger = l.logger.WithFields(allFields)
	}

	// 使用spoor记录日志
	switch level {
	case "DEBUG":
		logger.Debug(msg)
	case "INFO":
		logger.Info(msg)
	case "WARN":
		logger.Warn(msg)
	case "ERROR":
		logger.Error(msg)
	case "FATAL":
		logger.Fatal(msg)
	default:
		logger.Info(msg)
	}
}

// parseFields 解析字段
func (l *spoorLogger) parseFields(fields ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	if len(fields)%2 != 0 {
		fields = append(fields, "MISSING_VALUE")
	}

	for i := 0; i < len(fields); i += 2 {
		key, ok := fields[i].(string)
		if !ok {
			key = fmt.Sprintf("%v", fields[i])
		}
		result[key] = fields[i+1]
	}

	return result
}

// NewNop 创建空日志实例（用于测试）
func NewNop() Logger {
	logger := spoor.NewConsole(spoor.LevelInfo)
	
	return &spoorLogger{
		logger: logger,
		fields: make(map[string]interface{}),
	}
}

// NewDevelopment 创建开发环境日志实例
func NewDevelopment() (Logger, error) {
	logger := spoor.NewConsole(spoor.LevelDebug)
	
	return &spoorLogger{
		logger: logger,
		fields: make(map[string]interface{}),
	}, nil
}

// NewProduction 创建生产环境日志实例
func NewProduction() (Logger, error) {
	logger := spoor.NewConsole(spoor.LevelInfo)
	
	return &spoorLogger{
		logger: logger,
		fields: make(map[string]interface{}),
	}, nil
}