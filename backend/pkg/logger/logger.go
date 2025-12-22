package logger

import (
	"fmt"
	"os"
	"path/filepath"

	"line-management/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Logger *zap.Logger
var Sugar *zap.SugaredLogger

// InitLogger 初始化日志
func InitLogger() error {
	cfg := config.GlobalConfig.Log

	// 创建日志目录
	logDir := filepath.Dir(cfg.FilePath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 日志轮转配置
	lumberjackLogger := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,    // MB
		MaxBackups: cfg.MaxBackups, // 保留的备份数量
		MaxAge:     cfg.MaxAge,     // 保留的天数
		Compress:   cfg.Compress,   // 是否压缩
	}

	// 解析日志级别
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:      zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 控制台输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleWriter := zapcore.AddSync(os.Stdout)

	// 文件输出
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileWriter := zapcore.AddSync(lumberjackLogger)

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWriter, level),
		zapcore.NewCore(fileEncoder, fileWriter, level),
	)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	Sugar = Logger.Sugar()

	return nil
}

// Sync 同步日志缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// WithField 添加字段
func WithField(key string, value interface{}) *zap.SugaredLogger {
	return Sugar.With(key, value)
}

// WithFields 添加多个字段
func WithFields(fields map[string]interface{}) *zap.SugaredLogger {
	var args []interface{}
	for k, v := range fields {
		args = append(args, k, v)
	}
	return Sugar.With(args...)
}

// Debug 调试日志
func Debug(args ...interface{}) {
	Sugar.Debug(args...)
}

// Debugf 格式化调试日志
func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

// Info 信息日志
func Info(args ...interface{}) {
	Sugar.Info(args...)
}

// Infof 格式化信息日志
func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

// Warn 警告日志
func Warn(args ...interface{}) {
	Sugar.Warn(args...)
}

// Warnf 格式化警告日志
func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

// Error 错误日志
func Error(args ...interface{}) {
	Sugar.Error(args...)
}

// Errorf 格式化错误日志
func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

// Fatal 致命错误日志
func Fatal(args ...interface{}) {
	Sugar.Fatal(args...)
}

// Fatalf 格式化致命错误日志
func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

// Panic 恐慌日志
func Panic(args ...interface{}) {
	Sugar.Panic(args...)
}

// Panicf 格式化恐慌日志
func Panicf(template string, args ...interface{}) {
	Sugar.Panicf(template, args...)
}
