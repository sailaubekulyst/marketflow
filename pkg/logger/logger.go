// marketflow/pkg/logger/logger.go
package logger

import (
	"log/slog"
	"os"
	"time" // <-- Добавлен для форматирования времени
)

// Logger - это обертка над slog.Logger для удобного использования.
type Logger struct {
	*slog.Logger
}

// NewLogger создает новый экземпляр логгера.
// levelStr может быть "DEBUG", "INFO", "WARN", "ERROR".
func NewLogger(levelStr string) *Logger {
	var level slog.Level
	switch levelStr {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	loc, err := time.LoadLocation("Asia/Almaty")
	if err != nil {
		loc = time.Local
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Any().(time.Time)
				a.Value = slog.StringValue(t.In(loc).Format("2006-01-02T15:04:05.000Z07:00"))
			}
			return a
		},
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &Logger{slog.New(handler)}
}

// Fatal логирует ошибку на уровне ERROR и затем завершает приложение.
func (l *Logger) Fatal(msg string, args ...any) {
	// Используем log/slog для логирования, а затем os.Exit для завершения.
	l.Error(msg, args...)
	os.Exit(1)
}

// Debug логирует сообщение на уровне DEBUG.
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Info логирует сообщение на уровне INFO.
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Warn логирует сообщение на уровне WARN.
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Error логирует сообщение на уровне ERROR.
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}
