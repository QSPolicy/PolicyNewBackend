package config

import (
	"log"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	DatabaseURL      string `koanf:"database_url"`
	ServerAddress    string `koanf:"server_address"`
	JWTSecretKey     string `koanf:"jwt_secret_key"`
	JWTTokenDuration int    `koanf:"jwt_token_duration"` // 单位：小时

	// MySQL连接池配置（仅MySQL时有效）
	MySQLMaxIdleConns int `koanf:"mysql_max_idle_conns"`
	MySQLMaxOpenConns int `koanf:"mysql_max_open_conns"`
}

var k = koanf.New(".")

func LoadConfig() *Config {
	// 默认值
	k.Set("database_url", "sqlite3://policy.db")
	k.Set("server_address", ":8080")
	k.Set("jwt_secret_key", "default_jwt_secret_key_change_in_production")
	k.Set("jwt_token_duration", 24)    // 默认24小时
	k.Set("mysql_max_idle_conns", 10)  // MySQL连接池最大空闲连接数
	k.Set("mysql_max_open_conns", 100) // MySQL连接池最大打开连接数

	// 从文件读取
	if err := k.Load(file.Provider("config.yaml"), yaml.Parser()); err != nil {
		log.Printf("warn: error loading config.yaml: %v", err)
	}

	// 从环境变量读取
	// 从大写转小写 DATABASE_URL database_url
	if err := k.Load(env.Provider("", ".", func(s string) string {
		return strings.ToLower(s)
	}), nil); err != nil {
		log.Printf("error loading env: %v", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		log.Fatalf("error unmarshalling config: %v", err)
	}

	return &cfg
}
