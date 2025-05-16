package logger

import (
	"fmt"
	"github.com/seanbit/nrpc/logger/interfaces"
	slog "log"
	"os"
	"sync"
)

// Logger 实现了 ILogger 接口
type Logger struct {
	logger *slog.Logger
	mu     sync.Mutex
	prefix string
}

// NewDefaultLogger 创建一个新的 Logger 实例
func NewDefaultLogger() *Logger {
	return &Logger{
		logger: slog.New(os.Stdout, "", slog.LstdFlags),
	}
}

func (l *Logger) WithKv(key, value string) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("%s=%s ", key, value),
	}
}

func (l *Logger) WithKvs(key string, vals []string) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("%s=%v ", key, vals),
	}
}

func (l *Logger) WithJson(key string, b []byte) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("%s=%s ", key, string(b)),
	}
}

func (l *Logger) WithError(err error) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("error=%v ", err),
	}
}

func (l *Logger) WithErrors(key string, errs []error) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("%s=%v ", key, errs),
	}
}

func (l *Logger) WithFields(fields map[string]interface{}) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("fields=%v ", fields),
	}
}

func (l *Logger) WithField(key string, value interface{}) interfaces.Logger {
	return &Logger{
		logger: l.logger,
		prefix: l.prefix + fmt.Sprintf("%s=%v ", key, value),
	}
}

func (l *Logger) logf(level, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetPrefix(fmt.Sprintf("%s[%s] ", l.prefix, level))
	l.logger.Printf(format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logf("DEBUG", format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logf("INFO", format, args...)
}

func (l *Logger) Warnf(format string, args ...interface{}) {
	l.logf("WARN", format, args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.Warnf(format, args...)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logf("ERROR", format, args...)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logf("FATAL", format, args...)
	os.Exit(1)
}

func (l *Logger) Panicf(format string, args ...interface{}) {
	l.logf("PANIC", format, args...)
	panic(fmt.Sprintf(format, args...))
}

func (l *Logger) log(level string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetPrefix(fmt.Sprintf("%s[%s] ", l.prefix, level))
	l.logger.Println(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.log("WARN", args...)
}

func (l *Logger) Warning(args ...interface{}) {
	l.Warn(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *Logger) Fatal(args ...interface{}) {
	l.log("FATAL", args...)
	os.Exit(1)
}

func (l *Logger) Panic(args ...interface{}) {
	l.log("PANIC", args...)
	panic(args)
}

func (l *Logger) Debugln(args ...interface{}) {
	l.log("DEBUG", args...)
}

func (l *Logger) Infoln(args ...interface{}) {
	l.log("INFO", args...)
}

func (l *Logger) Warnln(args ...interface{}) {
	l.log("WARN", args...)
}

func (l *Logger) Warningln(args ...interface{}) {
	l.Warn(args...)
}

func (l *Logger) Errorln(args ...interface{}) {
	l.log("ERROR", args...)
}

func (l *Logger) Fatalln(args ...interface{}) {
	l.log("FATAL", args...)
	os.Exit(1)
}

func (l *Logger) Panicln(args ...interface{}) {
	l.log("PANIC", args...)
	panic(args)
}
