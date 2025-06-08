package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	logger *zap.Logger
}

func NewLogger() (*Logger, error) {
	// 建立日誌目錄
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("建立日誌目錄失敗: %v", err)
	}

	// 設定日誌檔案路徑
	logPath := filepath.Join(logDir, "app.log")
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("開啟日誌檔案失敗: %v", err)
	}

	// 設定 zap 的編碼器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 建立檔案輸出
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(logFile),
		zapcore.InfoLevel,
	)

	// 建立控制台輸出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zapcore.InfoLevel,
	)

	// 合併兩個輸出
	core := zapcore.NewTee(fileCore, consoleCore)

	// 建立 logger
	logger := zap.New(core, zap.AddCaller())

	return &Logger{
		logger: logger,
	}, nil
}

func (l *Logger) Close() error {
	return l.logger.Sync()
}

func (l *Logger) LogError(path string, errMsg string) {
	l.logger.Error("處理檔案失敗",
		zap.String("path", path),
		zap.String("error", errMsg),
		zap.Time("timestamp", time.Now()),
	)
}

func (l *Logger) LogInfo(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) LogDebug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

func (l *Logger) LogWarn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}
