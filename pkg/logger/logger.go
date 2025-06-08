package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	errorLog *os.File
}

func NewLogger() (*Logger, error) {
	errorLog, err := os.OpenFile("error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("無法建立錯誤日誌: %v", err)
	}

	return &Logger{
		errorLog: errorLog,
	}, nil
}

func (l *Logger) Close() error {
	return l.errorLog.Close()
}

func (l *Logger) LogError(path, message string) {
	logEntry := fmt.Sprintf("[%s] %s: %s\n", time.Now().Format("2006-01-02 15:04:05"), path, message)
	l.errorLog.WriteString(logEntry)
}
