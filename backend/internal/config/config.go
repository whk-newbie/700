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
	// 设置viper默认值
	setViperDefaults()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取工作目录失败: %w", err)
	}

	// 设置配置文件路径
	configPath := filepath.Join(workDir, ".env")

	// 如果.env文件存在，先读取配置文件
	if _, err := os.Stat(configPath); err == nil {
		viper.SetConfigFile(configPath)
		viper.SetConfigType("env")
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}
		// 将.env文件中的大写键名映射到结构体路径
		// 因为.env文件使用DATABASE_PASSWORD格式，但结构体使用database.password
		if val := viper.GetString("DATABASE_PASSWORD"); val != "" {
			viper.Set("database.password", val)
		}
		if val := viper.GetString("DATABASE_HOST"); val != "" {
			viper.Set("database.host", val)
		}
		if val := viper.GetInt("DATABASE_PORT"); val != 0 {
			viper.Set("database.port", val)
		}
		if val := viper.GetString("DATABASE_USER"); val != "" {
			viper.Set("database.user", val)
		}
		if val := viper.GetString("DATABASE_DBNAME"); val != "" {
			viper.Set("database.dbname", val)
		}
		if val := viper.GetString("DATABASE_SSLMODE"); val != "" {
			viper.Set("database.sslmode", val)
		}
		if val := viper.GetString("DATABASE_TIMEZONE"); val != "" {
			viper.Set("database.timezone", val)
		}
		// Redis配置
		if val := viper.GetString("REDIS_HOST"); val != "" {
			viper.Set("redis.host", val)
		}
		if val := viper.GetInt("REDIS_PORT"); val != 0 {
			viper.Set("redis.port", val)
		}
		if val := viper.GetString("REDIS_PASSWORD"); val != "" {
			viper.Set("redis.password", val)
		}
		if val := viper.GetInt("REDIS_DB"); val != 0 || viper.IsSet("REDIS_DB") {
			viper.Set("redis.db", viper.GetInt("REDIS_DB"))
		}
		// 服务器配置
		if val := viper.GetString("SERVER_PORT"); val != "" {
			viper.Set("server.port", val)
		}
		if val := viper.GetString("SERVER_MODE"); val != "" {
			viper.Set("server.mode", val)
		}
		// JWT配置
		if val := viper.GetString("JWT_SECRET"); val != "" {
			viper.Set("jwt.secret", val)
		}
		if val := viper.GetInt("JWT_EXPIRE_HOUR"); val != 0 {
			viper.Set("jwt.expire_hour", val)
		}
	}

	// 设置环境变量自动读取（环境变量优先级高于配置文件）
	viper.AutomaticEnv()
	viper.SetEnvPrefix("")

	// 绑定环境变量（支持下划线和点号两种格式）
	bindEnvVars()

	// 解析配置到结构体
	GlobalConfig = &Config{}
	if err := viper.Unmarshal(GlobalConfig); err != nil {
		return fmt.Errorf("解析配置失败: %w", err)
	}

	// 如果配置为空，使用默认配置
	if GlobalConfig.Server.Port == "" {
		return initDefaultConfig()
	}

	// 调试：检查数据库密码是否正确读取（生产环境应移除）
	if GlobalConfig.Database.Password == "" || GlobalConfig.Database.Password == "linepass" {
		fmt.Printf("警告: 数据库密码可能未正确读取，当前值: %s\n", GlobalConfig.Database.Password)
	}

	return nil
}

// bindEnvVars 绑定环境变量
func bindEnvVars() {
	// 服务器配置
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "SERVER_MODE")

	// 数据库配置
	viper.BindEnv("database.host", "DATABASE_HOST")
	viper.BindEnv("database.port", "DATABASE_PORT")
	viper.BindEnv("database.user", "DATABASE_USER")
	viper.BindEnv("database.password", "DATABASE_PASSWORD")
	viper.BindEnv("database.dbname", "DATABASE_DBNAME")
	viper.BindEnv("database.sslmode", "DATABASE_SSLMODE")
	viper.BindEnv("database.timezone", "DATABASE_TIMEZONE")

	// Redis配置
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")

	// JWT配置
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("jwt.expire_hour", "JWT_EXPIRE_HOUR")

	// 日志配置
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.file_path", "LOG_FILE_PATH")
	viper.BindEnv("log.max_size", "LOG_MAX_SIZE")
	viper.BindEnv("log.max_backups", "LOG_MAX_BACKUPS")
	viper.BindEnv("log.max_age", "LOG_MAX_AGE")
	viper.BindEnv("log.compress", "LOG_COMPRESS")

	// Swagger配置
	viper.BindEnv("swagger.enable", "SWAGGER_ENABLE")
	viper.BindEnv("swagger.host", "SWAGGER_HOST")

	// WebSocket配置
	viper.BindEnv("websocket.read_timeout", "WEBSOCKET_READ_TIMEOUT")
	viper.BindEnv("websocket.write_timeout", "WEBSOCKET_WRITE_TIMEOUT")
	viper.BindEnv("websocket.ping_period", "WEBSOCKET_PING_PERIOD")
	viper.BindEnv("websocket.max_message_size", "WEBSOCKET_MAX_MESSAGE_SIZE")

	// LLM配置
	viper.BindEnv("llm.default_provider", "LLM_DEFAULT_PROVIDER")
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
