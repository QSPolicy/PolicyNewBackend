package config

import (
	"log"
	"reflect"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"policy-backend/auth"
	"policy-backend/database"
	"policy-backend/utils"
)

// AppConfig 聚合所有模块的配置（扁平结构，用于 koanf unmarshal）
type AppConfig struct {
	// Server
	ServerAddress string `koanf:"server_address"`

	// Database
	DatabaseURL       string `koanf:"database_url"`
	MySQLMaxIdleConns int    `koanf:"mysql_max_idle_conns"`
	MySQLMaxOpenConns int    `koanf:"mysql_max_open_conns"`

	// Auth
	JWTSecretKey            string `koanf:"jwt_secret_key"`
	JWTAccessTokenDuration  int    `koanf:"jwt_access_token_duration"`
	JWTRefreshTokenDuration int    `koanf:"jwt_refresh_token_duration"`

	// Log
	LogLevel string `koanf:"log_level"`
	LogFile  string `koanf:"log_file"`
}

// Config 对外暴露的配置结构，包含各模块独立的配置
type Config struct {
	Server   ServerConfig
	Database database.Config
	Auth     auth.Config
	Log      utils.LogConfig
}

// defaultAppConfig 聚合所有模块的默认配置
func defaultAppConfig() AppConfig {
	serverDef := DefaultServerConfig()
	dbDef := database.DefaultConfig()
	authDef := auth.DefaultConfig()
	logDef := utils.DefaultLogConfig()

	return AppConfig{
		// Server
		ServerAddress: serverDef.ServerAddress,

		// Database
		DatabaseURL:       dbDef.DatabaseURL,
		MySQLMaxIdleConns: dbDef.MySQLMaxIdleConns,
		MySQLMaxOpenConns: dbDef.MySQLMaxOpenConns,

		// Auth
		JWTSecretKey:            authDef.JWTSecretKey,
		JWTAccessTokenDuration:  authDef.JWTAccessTokenDuration,
		JWTRefreshTokenDuration: authDef.JWTRefreshTokenDuration,

		// Log
		LogLevel: logDef.LogLevel,
		LogFile:  logDef.LogFile,
	}
}

// setDefaults 将默认配置设置到 koanf 实例
func setDefaults(k *koanf.Koanf, defaults AppConfig) {
	v := reflect.ValueOf(defaults)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("koanf")
		if tag != "" {
			k.Set(tag, v.Field(i).Interface())
		}
	}
}

// LoadConfig 加载所有配置
func LoadConfig() *Config {
	k := koanf.New(".")

	// 1. 设置所有模块的默认值
	setDefaults(k, defaultAppConfig())

	// 2. 从文件读取（覆盖默认值）
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Printf("warn: error loading config.yaml: %v", err)
	}

	// 3. 从环境变量读取（最高优先级）
	// DATABASE_URL -> database_url
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(s)
	}), nil); err != nil {
		log.Printf("error loading env: %v", err)
	}

	// 4. Unmarshal 到扁平结构
	var app AppConfig
	if err := k.Unmarshal("", &app); err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	// 5. 分发到各模块配置结构
	return &Config{
		Server: ServerConfig{
			ServerAddress: app.ServerAddress,
		},
		Database: database.Config{
			DatabaseURL:       app.DatabaseURL,
			MySQLMaxIdleConns: app.MySQLMaxIdleConns,
			MySQLMaxOpenConns: app.MySQLMaxOpenConns,
		},
		Auth: auth.Config{
			JWTSecretKey:            app.JWTSecretKey,
			JWTAccessTokenDuration:  app.JWTAccessTokenDuration,
			JWTRefreshTokenDuration: app.JWTRefreshTokenDuration,
		},
		Log: utils.LogConfig{
			LogLevel: app.LogLevel,
			LogFile:  app.LogFile,
		},
	}
}
