package config

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// Config 结构体表示整个配置文件
type Config struct {
	Feishu   FeishuConfig   `yaml:"feishu"`   // 飞书相关配置
	Database DatabaseConfig `yaml:"database"` // 数据库相关配置
	Server   ServerConfig   `yaml:"server"`   // 服务器相关配置
	Logger   LoggerConfig   `yaml:"logger"`   // 日志相关配置
}

// FeishuConfig 结构体表示飞书机器人的配置
type FeishuConfig struct {
	AppID             string `yaml:"app_id"`             // 飞书应用的 App ID
	AppSecret         string `yaml:"app_secret"`         // 飞书应用的 App Secret
	VerificationToken string `yaml:"verification_token"` // 事件订阅的验证令牌
	EncryptKey        string `yaml:"encrypt_key"`        // 事件订阅的加密密钥
	GroupChats        []string `yaml:"group_chats"`        // 群聊ID列表
}

// DatabaseConfig 结构体表示数据库配置
type DatabaseConfig struct {
	Host         string `yaml:"host"`           // 数据库主机
	Port         int    `yaml:"port"`           // 数据库端口
	User         string `yaml:"user"`           // 数据库用户名
	Password     string `yaml:"password"`       // 数据库密码
	DBName       string `yaml:"dbname"`         // 数据库名称
	MaxOpenConns int    `yaml:"max_open_conns"` // 最大打开连接数
	MaxIdleConns int    `yaml:"max_idle_conns"` // 最大空闲连接数
}

// ServerConfig 结构体表示服务器配置
type ServerConfig struct {
	Port int `yaml:"port"` // 服务器监听端口
}

// LoggerConfig 结构体表示日志配置
type LoggerConfig struct {
	Level string `yaml:"level"` // 日志级别 (debug, info, warn, error, dpanic, panic, fatal)
	Path  string `yaml:"path"`  // 日志文件路径
}

// LoadConfig 加载配置文件并解析为 Config 结构体
func LoadConfig() (*Config, error) {
	// 打开配置文件
	f, err := os.Open("configs/config.yaml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	// 使用 YAML 解码器解析配置文件
	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	// 从环境变量覆盖配置
	overrideFromEnv(&cfg)

	return &cfg, nil
}

// overrideFromEnv 从环境变量覆盖配置
func overrideFromEnv(cfg *Config) {
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if portStr := os.Getenv("DB_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Database.Port = port
		}
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.User = user
	}
	if password := os.Getenv("DB_PASSWORD"); password != "" {
		cfg.Database.Password = password
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		cfg.Database.DBName = dbName
	}

	// 飞书配置
	if appID := os.Getenv("FEISHU_APP_ID"); appID != "" {
		cfg.Feishu.AppID = appID
	}
	if appSecret := os.Getenv("FEISHU_APP_SECRET"); appSecret != "" {
		cfg.Feishu.AppSecret = appSecret
	}
	if token := os.Getenv("FEISHU_VERIFICATION_TOKEN"); token != "" {
		cfg.Feishu.VerificationToken = token
	}
	if key := os.Getenv("FEISHU_ENCRYPT_KEY"); key != "" {
		cfg.Feishu.EncryptKey = key
	}

	// 服务器配置
	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Server.Port = port
		}
	}

	// 群聊ID列表
	if groupChats := os.Getenv("GROUP_CHATS"); groupChats != "" {
		cfg.Feishu.GroupChats = strings.Split(groupChats, ",")
	}

	// 日志配置
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.Logger.Level = level
	}
	if path := os.Getenv("LOG_PATH"); path != "" {
		cfg.Logger.Path = path
	}
}
