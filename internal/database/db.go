package database

import (
	"MikoNews/internal/config"
	"MikoNews/internal/pkg/logger"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DB 是数据库连接的封装
type DB struct {
	*gorm.DB
}

// New 创建一个新的数据库连接
func New(conf *config.DatabaseConfig) (*DB, error) {
	// 构建DSN连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.User, conf.Password, conf.Host, conf.Port, conf.DBName)

	// 配置GORM日志
	gormLogLevel := gormlogger.Info
	// 根据应用日志级别调整GORM日志级别
	appLogLevel := logger.GetLogger().Core().Enabled(zap.DebugLevel) // Check if Debug is enabled in zap
	if !appLogLevel {                                                // If zap debug is not enabled, set gorm to Warn
		gormLogLevel = gormlogger.Warn
	}

	// 使用自定义的 Zap Writer
	zapWriter := &ZapGormWriter{Logger: logger.GetLogger().WithOptions(zap.AddCallerSkip(4))}

	gormLogger := gormlogger.New(
		zapWriter,
		gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond, // 慢查询阈值
			LogLevel:                  gormLogLevel,           // GORM日志级别
			IgnoreRecordNotFoundError: true,                   // 忽略记录未找到错误
			Colorful:                  false,                  // Zap会处理颜色
		},
	)

	// 打开数据库连接
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}

	// 获取底层SQL连接以设置连接池
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{DB: gormDB}, nil
}

// Close 关闭数据库连接
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// --- Zap GORM Writer Implementation ---

// ZapGormWriter 实现了 gormlogger.Writer 接口，将 GORM 日志写入 Zap
type ZapGormWriter struct {
	Logger *zap.Logger
}

// Printf 实现了 gormlogger.Writer 接口的 Printf 方法
func (z *ZapGormWriter) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)

	// 根据日志内容判断级别（这是一个简化的判断，可以根据需要调整）
	switch {
	case msg[0:7] == "[error]":
		z.Logger.Error(msg)
	case msg[0:6] == "[warn]":
		z.Logger.Warn(msg)
	default:
		z.Logger.Info(msg) // 默认Info级别，GORM的Trace信息会进入这里
	}
}

// -- ZapGormLogger（可选，如果需要更精细控制Trace等） --
// type zapGormLogger struct {
// 	log *zap.Logger
// }
//
// func NewZapGormLogger(log *zap.Logger) gormlogger.Interface {
// 	return &zapGormLogger{log: log.WithOptions(zap.AddCallerSkip(4))}
// }
//
// func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
// 	return l
// }
//
// func (l *zapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
// 	l.log.Info(fmt.Sprintf(msg, data...))
// }
//
// func (l *zapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
// 	l.log.Warn(fmt.Sprintf(msg, data...))
// }
//
// func (l *zapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
// 	l.log.Error(fmt.Sprintf(msg, data...))
// }
//
// func (l *zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
// 	elapsed := time.Since(begin)
// 	sql, rows := fc()
// 	fields := []zap.Field{
// 		zap.Duration("elapsed", elapsed),
// 		zap.Int64("rows", rows),
// 		zap.String("sql", sql),
// 	}
//
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		l.log.Error("GORM Trace", append(fields, zap.Error(err))...)
// 	} else {
// 		l.log.Debug("GORM Trace", fields...)
// 	}
// }
