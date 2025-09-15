package logger

import (
	"os"

	"github.com/ix-pay/ixpay-pro/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 定义日志接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
}

// zapLogger 实现Logger接口
type zapLogger struct {
	logger *zap.Logger
}

// NewLogger 创建新的日志记录器
func NewLogger(cfg *config.Config) Logger {
	// 设置日志级别
	var level zapcore.Level
	switch cfg.Logging.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	case "fatal":
		level = zap.FatalLevel
	default:
		level = zap.InfoLevel
	}

	// 创建编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建编码器
	var encoder zapcore.Encoder
	if cfg.Logging.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// 创建文件写入器
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   cfg.Logging.FilePath,
		MaxSize:    cfg.Logging.MaxSize, // MB
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge, // days
		Compress:   true,
	})

	// 创建核心
	core := zapcore.NewCore(
		encoder,
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter), zapcore.Lock(os.Stdout)),
		level,
	)

	// 创建日志记录器
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))

	return &zapLogger{
		logger: logger,
	}
}

// Debug 记录调试信息
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, parseFields(fields)...)
}

// Info 记录信息
func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, parseFields(fields)...)
}

// Warn 记录警告信息
func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, parseFields(fields)...)
}

// Error 记录错误信息
func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.logger.Error(msg, parseFields(fields)...)
}

// Fatal 记录致命错误信息并退出
func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.logger.Fatal(msg, parseFields(fields)...)
}

// parseFields 解析日志字段
func parseFields(fields []interface{}) []zap.Field {
	var zapFields []zap.Field

	for i := 0; i < len(fields); i += 2 {
		if i+1 >= len(fields) {
			break
		}

		key, ok := fields[i].(string)
		if !ok {
			continue
		}

		value := fields[i+1]
		zapFields = append(zapFields, zap.Any(key, value))
	}

	return zapFields
}
