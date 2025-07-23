package logger

import (
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger 初始化zap日志记录器
func InitLogger() error {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return err
	}

	// 配置编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 彩色级别编码
		EncodeTime:     customTimeEncoder,                // 自定义时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建控制台编码器（用于输出到终端，带颜色）
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// 创建文件编码器（用于输出到文件，JSON格式）
	fileEncoderConfig := encoderConfig
	fileEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder // 文件中不使用颜色
	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)

	// 创建日志文件
	logFile, err := os.OpenFile(
		filepath.Join("logs", "app.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	// 创建错误日志文件
	errorFile, err := os.OpenFile(
		filepath.Join("logs", "error.log"),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0644,
	)
	if err != nil {
		return err
	}

	// 配置日志级别
	level := zapcore.InfoLevel
	if os.Getenv("LOG_LEVEL") == "debug" {
		level = zapcore.DebugLevel
	}

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),           // 控制台输出
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logFile), level),               // 所有日志文件
		zapcore.NewCore(fileEncoder, zapcore.AddSync(errorFile), zapcore.ErrorLevel), // 错误日志文件
	)

	// 创建logger
	Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

// Sync 刷新日志缓冲区
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}
