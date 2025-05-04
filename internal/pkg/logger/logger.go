package logger

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.Logger

// InitLogger 初始化 zap 日志记录器
func InitLogger(logPath string, logLevel string) {
	// 配置 lumberjack 用于日志分割
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath, // 日志文件路径
		MaxSize:    100,     // 每个日志文件的最大大小（MB）
		MaxBackups: 5,       // 保留旧文件的最大个数
		MaxAge:     30,      // 保留旧文件的最大天数
		Compress:   false,   // 是否压缩/归档旧文件
		LocalTime:  true,    // 使用本地时间进行日志切割
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	case "dpanic":
		level = zapcore.DPanicLevel
	case "panic":
		level = zapcore.PanicLevel
	case "fatal":
		level = zapcore.FatalLevel
	default:
		level = zapcore.InfoLevel
	}

	// 配置 zap 日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "T", // 使用短Key，减少冗余
		LevelKey:       "L", // 使用短Key
		NameKey:        "N",
		CallerKey:      "C",             // 使用短Key
		FunctionKey:    zapcore.OmitKey, // 省略函数名
		MessageKey:     "M",
		StacktraceKey:  "S",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    CustomLevelEncoder, // 使用自定义的级别编码器
		EncodeTime:     CustomTimeEncoder,  // 使用自定义时间格式
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder, // 短路径编码器
	}

	// 设置核心：同时输出到控制台和文件
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(lumberjackLogger)),
		zap.NewAtomicLevelAt(level),
	)

	// 构建日志记录器，添加调用者信息和堆栈跟踪（仅 Error 级别及以上）
	log = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.Development(), zap.AddStacktrace(zapcore.ErrorLevel))

	Info("Zap Logger initialized successfully", "logLevel", logLevel, "logPath", logPath)
}

// GetLogger 返回全局日志记录器实例
func GetLogger() *zap.Logger {
	return log
}

// Sync 将缓冲区中的日志刷新到磁盘
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}

// CustomTimeEncoder 自定义时间格式化函数
func CustomTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000 MST"))
}

// CustomLevelEncoder 自定义级别编码器，输出 [LEVEL] 格式
func CustomLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(fmt.Sprintf("[%s]", level.CapitalString())) // 输出如 [INFO], [DEBUG] 等
}

// ====== 简化的日志方法 ======

// Debug 输出调试日志，支持k-v对
func Debug(msg string, keysAndValues ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Debugw(msg, keysAndValues...)
}

// Debugf 输出格式化的调试日志
func Debugf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Debugf(format, args...)
}

// Info 输出信息日志，支持k-v对
func Info(msg string, keysAndValues ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Infow(msg, keysAndValues...)
}

// Infof 输出格式化的信息日志
func Infof(format string, args ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Infof(format, args...)
}

// Warn 输出警告日志，支持k-v对
func Warn(msg string, keysAndValues ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Warnw(msg, keysAndValues...)
}

// Warnf 输出格式化的警告日志
func Warnf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Warnf(format, args...)
}

// Error 输出错误日志，支持k-v对
func Error(msg string, keysAndValues ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Errorw(msg, keysAndValues...)
}

// Errorf 输出格式化的错误日志
func Errorf(format string, args ...interface{}) {
	if log == nil {
		return
	}
	sugar := log.Sugar()
	sugar.Errorf(format, args...)
}

// Fatal 输出致命错误日志并退出程序，支持k-v对
func Fatal(msg string, keysAndValues ...interface{}) {
	if log == nil {
		os.Exit(1)
	}
	sugar := log.Sugar()
	sugar.Fatalw(msg, keysAndValues...)
}

// Fatalf 输出格式化的致命错误日志并退出程序
func Fatalf(format string, args ...interface{}) {
	if log == nil {
		os.Exit(1)
	}
	sugar := log.Sugar()
	sugar.Fatalf(format, args...)
}

// Panic 输出严重错误日志并抛出panic，支持k-v对
func Panic(msg string, keysAndValues ...interface{}) {
	if log == nil {
		panic(msg)
	}
	sugar := log.Sugar()
	sugar.Panicw(msg, keysAndValues...)
}

// Panicf 输出格式化的严重错误日志并抛出panic
func Panicf(format string, args ...interface{}) {
	if log == nil {
		panic(fmt.Sprintf(format, args...))
	}
	sugar := log.Sugar()
	sugar.Panicf(format, args...)
}
 