package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config 全局配置结构体
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Swagger  SwaggerConfig  `mapstructure:"swagger"`
	WebSocket WebSocketConfig `mapstructure:"websocket"`
	LLM      LLMConfig      `mapstructure:"llm"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"timezone"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

type SwaggerConfig struct {
	Enable bool   `mapstructure:"enable"`
	Host   string `mapstructure:"host"`
}

type WebSocketConfig struct {
	ReadTimeout    int `mapstructure:"read_timeout"`
	WriteTimeout   int `mapstructure:"write_timeout"`
	PingPeriod     int `mapstructure:"ping_period"`
	MaxMessageSize int `mapstructure:"max_message_size"`
}

type LLMConfig struct {
	DefaultProvider string                 `mapstructure:"default_provider"`
	Providers       map[string]LLMProvider `mapstructure:"providers"`
}

type LLMProvider struct {
	APIKey      string `mapstructure:"api_key"`
	BaseURL     string `mapstructure:"base_url"`
	Model       string `mapstructure:"model"`
	MaxTokens   int    `mapstructure:"max_tokens"`
	Temperature float64 `mapstructure:"temperature"`
}

// GlobalConfig 全局配置实例
var GlobalConfig *Config

// InitConfig 初始化配置
func InitConfig() error {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 设置配置文件路径
	configPath := filepath.Join(workDir, ".env")

	// 如果.env文件不存在，使用默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return initDefaultConfig()
	}

	// 设置viper配置
	viper.SetConfigFile(configPath)
	viper.SetConfigType("env")

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	if err := viper.Unmarshal(&GlobalConfig); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	return nil
}

// initDefaultConfig 初始化默认配置
func initDefaultConfig() error {
	GlobalConfig = &Config{
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "lineuser",
			Password: "linepass",
			DBName:   "line_management",
			SSLMode:  "disable",
			TimeZone: "Asia/Shanghai",
		},
		Redis: RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:     "your-secret-key",
			ExpireHour: 24,
		},
		Log: LogConfig{
			Level:      "info",
			FilePath:   "./logs/app.log",
			MaxSize:    100,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		},
		Swagger: SwaggerConfig{
			Enable: true,
			Host:   "localhost:8080",
		},
		WebSocket: WebSocketConfig{
			ReadTimeout:    60,
			WriteTimeout:   60,
			PingPeriod:     54,
			MaxMessageSize: 4096,
		},
		LLM: LLMConfig{
			DefaultProvider: "openai",
			Providers: map[string]LLMProvider{
				"openai": {
					APIKey:      "",
					BaseURL:     "https://api.openai.com/v1",
					Model:       "gpt-3.5-turbo",
					MaxTokens:   2000,
					Temperature: 0.7,
				},
			},
		},
	}

	// 设置viper默认值
	setViperDefaults()

	return nil
}

// setViperDefaults 设置viper默认值
func setViperDefaults() {
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "lineuser")
	viper.SetDefault("database.password", "linepass")
	viper.SetDefault("database.dbname", "line_management")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.timezone", "Asia/Shanghai")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("jwt.secret", "your-secret-key")
	viper.SetDefault("jwt.expire_hour", 24)
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.file_path", "./logs/app.log")
	viper.SetDefault("log.max_size", 100)
	viper.SetDefault("log.max_backups", 3)
	viper.SetDefault("log.max_age", 28)
	viper.SetDefault("log.compress", true)
	viper.SetDefault("swagger.enable", true)
	viper.SetDefault("swagger.host", "localhost:8080")
	viper.SetDefault("websocket.read_timeout", 60)
	viper.SetDefault("websocket.write_timeout", 60)
	viper.SetDefault("websocket.ping_period", 54)
	viper.SetDefault("websocket.max_message_size", 4096)
	viper.SetDefault("llm.default_provider", "openai")
}
