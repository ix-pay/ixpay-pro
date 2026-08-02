package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) Logger
	WithContext(ctx context.Context) Logger
	Sync() error
}

type LoggerType string

const (
	DefaultLogger LoggerType = "default"
	ErrorLogger   LoggerType = "error"
	TaskLogger    LoggerType = "task"
	RequestLogger LoggerType = "request"
	AuditLogger   LoggerType = "audit"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

type zapLogger struct {
	logger        *zap.Logger
	errorLogger   *zap.Logger
}

type MultiLogger struct {
	defaultLogger Logger
	errorLogger   Logger
	taskLogger    Logger
	requestLogger Logger
	auditLogger   Logger
}

var globalMultiLogger *MultiLogger

func SetGlobalMultiLogger(ml *MultiLogger) {
	globalMultiLogger = ml
}

func GetGlobalLogger(loggerType LoggerType) Logger {
	if globalMultiLogger == nil {
		return nil
	}
	return globalMultiLogger.GetLogger(loggerType)
}

func (ml *MultiLogger) GetLogger(loggerType LoggerType) Logger {
	switch loggerType {
	case ErrorLogger:
		return ml.errorLogger
	case TaskLogger:
		return ml.taskLogger
	case RequestLogger:
		return ml.requestLogger
	case AuditLogger:
		return ml.auditLogger
	default:
		return ml.defaultLogger
	}
}

func SetupLogger(cfg *config.Config) Logger {
	defaultLogger := setupLoggerWithType(cfg, DefaultLogger)
	errorLogger := setupLoggerWithType(cfg, ErrorLogger)

	if defaultZapLogger, ok := defaultLogger.(*zapLogger); ok {
		if errorZapLogger, ok := errorLogger.(*zapLogger); ok {
			defaultZapLogger.errorLogger = errorZapLogger.logger
		}
	}

	return defaultLogger
}

func SetupMultiLogger(cfg *config.Config) *MultiLogger {
	defaultLogger := setupLoggerWithType(cfg, DefaultLogger)
	errorLogger := setupLoggerWithType(cfg, ErrorLogger)

	if defaultZapLogger, ok := defaultLogger.(*zapLogger); ok {
		if errorZapLogger, ok := errorLogger.(*zapLogger); ok {
			defaultZapLogger.errorLogger = errorZapLogger.logger
		}
	}

	return &MultiLogger{
		defaultLogger: defaultLogger,
		errorLogger:   errorLogger,
		taskLogger:    setupLoggerWithType(cfg, TaskLogger),
		requestLogger: setupLoggerWithType(cfg, RequestLogger),
		auditLogger:   setupLoggerWithType(cfg, AuditLogger),
	}
}

func setupLoggerWithType(cfg *config.Config, loggerType LoggerType) Logger {
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

	if loggerType == ErrorLogger {
		level = zap.ErrorLevel
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var encoder zapcore.Encoder
	if cfg.Logging.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = newOrderedJSONEncoder(encoderConfig)
	}

	var filePath string
	switch loggerType {
	case DefaultLogger:
		filePath = cfg.Logging.FilePath
	case ErrorLogger:
		filePath = cfg.Logging.ErrorFile
	case TaskLogger:
		filePath = cfg.Logging.TaskFile
	case RequestLogger:
		filePath = cfg.Logging.RequestFile
	case AuditLogger:
		filePath = cfg.Logging.AuditFile
	}

	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.Logging.MaxSize,
		MaxBackups: cfg.Logging.MaxBackups,
		MaxAge:     cfg.Logging.MaxAge,
		Compress:   true,
	})

	var core zapcore.Core
	if loggerType == DefaultLogger {
		core = zapcore.NewCore(
			encoder,
			zapcore.NewMultiWriteSyncer(zapcore.AddSync(fileWriter), zapcore.Lock(os.Stdout)),
			level,
		)
	} else {
		core = zapcore.NewCore(
			encoder,
			zapcore.AddSync(fileWriter),
			level,
		)
	}

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	errorLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(2))

	return &zapLogger{
		logger: logger,
		errorLogger: errorLogger,
	}
}

func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Debug(msg, parseFields(fields)...)
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.logger.Info(msg, parseFields(fields)...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Warn(msg, parseFields(fields)...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	if l.errorLogger != nil {
		l.errorLogger.Error(msg, parseFields(fields)...)
	} else {
		l.logger.Error(msg, parseFields(fields)...)
	}
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	if l.errorLogger != nil {
		l.errorLogger.Fatal(msg, parseFields(fields)...)
	} else {
		l.logger.Fatal(msg, parseFields(fields)...)
	}
}

func (l *zapLogger) With(fields ...interface{}) Logger {
	return &zapLogger{
		logger: l.logger.With(parseFields(fields)...),
	}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	traceID := GetTraceID(ctx)
	if traceID != "" {
		return &zapLogger{
			logger:      l.logger.With(zap.String("trace_id", traceID)),
			errorLogger: l.errorLogger.With(zap.String("trace_id", traceID)),
		}
	}
	return l
}

func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

func NewTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func parseFields(fields []interface{}) []zap.Field {
	var zapFields []zap.Field

	for _, field := range fields {
		if f, ok := field.(zap.Field); ok {
			zapFields = append(zapFields, f)
			continue
		}
	}

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

var keyPriority = map[string]int{
	"id":        1,
	"level":     2,
	"timestamp": 3,
	"caller":    4,
	"msg":       5,
}

func newOrderedJSONEncoder(cfg zapcore.EncoderConfig) zapcore.Encoder {
	return &orderedJSONEncoder{
		jsonEncoder: zapcore.NewJSONEncoder(cfg),
	}
}

type orderedJSONEncoder struct {
	jsonEncoder zapcore.Encoder
}

func (e *orderedJSONEncoder) Clone() zapcore.Encoder {
	return &orderedJSONEncoder{
		jsonEncoder: e.jsonEncoder.Clone(),
	}
}

func (e *orderedJSONEncoder) EncodeEntry(ent zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf, err := e.jsonEncoder.EncodeEntry(ent, fields)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		return buf, nil
	}

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		pi, oki := keyPriority[keys[i]]
		pj, okj := keyPriority[keys[j]]

		if oki && okj {
			return pi < pj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}

		return strings.Compare(keys[i], keys[j]) < 0
	})

	ordered := make([]string, 0, len(keys))
	ordered = append(ordered, keys...)

	var sb strings.Builder
	sb.WriteString("{")
	for i, k := range ordered {
		if i > 0 {
			sb.WriteString(",")
		}
		v := data[k]
		b, _ := json.Marshal(k)
		sb.Write(b)
		sb.WriteString(":")
		b, _ = json.Marshal(v)
		sb.Write(b)
	}
	sb.WriteString("}")
	sb.WriteString("\n")

	newBuf := buffer.NewPool().Get()
	newBuf.AppendString(sb.String())

	buf.Free()
	return newBuf, nil
}

func (e *orderedJSONEncoder) AddArray(key string, val zapcore.ArrayMarshaler) error {
	return e.jsonEncoder.AddArray(key, val)
}

func (e *orderedJSONEncoder) AddObject(key string, val zapcore.ObjectMarshaler) error {
	return e.jsonEncoder.AddObject(key, val)
}

func (e *orderedJSONEncoder) AddBinary(key string, val []byte) {
	e.jsonEncoder.AddBinary(key, val)
}

func (e *orderedJSONEncoder) AddBool(key string, val bool) {
	e.jsonEncoder.AddBool(key, val)
}

func (e *orderedJSONEncoder) AddByteString(key string, val []byte) {
	e.jsonEncoder.AddByteString(key, val)
}

func (e *orderedJSONEncoder) AddComplex128(key string, val complex128) {
	e.jsonEncoder.AddComplex128(key, val)
}

func (e *orderedJSONEncoder) AddComplex64(key string, val complex64) {
	e.jsonEncoder.AddComplex64(key, val)
}

func (e *orderedJSONEncoder) AddFloat64(key string, val float64) {
	e.jsonEncoder.AddFloat64(key, val)
}

func (e *orderedJSONEncoder) AddFloat32(key string, val float32) {
	e.jsonEncoder.AddFloat32(key, val)
}

func (e *orderedJSONEncoder) AddInt(key string, val int) {
	e.jsonEncoder.AddInt(key, val)
}

func (e *orderedJSONEncoder) AddInt64(key string, val int64) {
	e.jsonEncoder.AddInt64(key, val)
}

func (e *orderedJSONEncoder) AddInt32(key string, val int32) {
	e.jsonEncoder.AddInt32(key, val)
}

func (e *orderedJSONEncoder) AddInt16(key string, val int16) {
	e.jsonEncoder.AddInt16(key, val)
}

func (e *orderedJSONEncoder) AddInt8(key string, val int8) {
	e.jsonEncoder.AddInt8(key, val)
}

func (e *orderedJSONEncoder) AddString(key string, val string) {
	e.jsonEncoder.AddString(key, val)
}

func (e *orderedJSONEncoder) AddTime(key string, val time.Time) {
	e.jsonEncoder.AddTime(key, val)
}

func (e *orderedJSONEncoder) AddUint(key string, val uint) {
	e.jsonEncoder.AddUint(key, val)
}

func (e *orderedJSONEncoder) AddUint64(key string, val uint64) {
	e.jsonEncoder.AddUint64(key, val)
}

func (e *orderedJSONEncoder) AddUint32(key string, val uint32) {
	e.jsonEncoder.AddUint32(key, val)
}

func (e *orderedJSONEncoder) AddUint16(key string, val uint16) {
	e.jsonEncoder.AddUint16(key, val)
}

func (e *orderedJSONEncoder) AddUint8(key string, val uint8) {
	e.jsonEncoder.AddUint8(key, val)
}

func (e *orderedJSONEncoder) AddUintptr(key string, val uintptr) {
	e.jsonEncoder.AddUintptr(key, val)
}

func (e *orderedJSONEncoder) AddReflected(key string, val interface{}) error {
	return e.jsonEncoder.AddReflected(key, val)
}

func (e *orderedJSONEncoder) OpenNamespace(key string) {
	e.jsonEncoder.OpenNamespace(key)
}

func (e *orderedJSONEncoder) AddDuration(key string, val time.Duration) {
	e.jsonEncoder.AddDuration(key, val)
}
