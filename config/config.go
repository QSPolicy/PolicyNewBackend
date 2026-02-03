package config

import (
	"encoding/json"
	"log"
	"reflect"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"

	"policy-backend/agent"
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

	// Agent (LLM + 工具)
	LLMConfigs        []agent.LLMConfig `koanf:"llm_configs"`
	BaiduSearchAPIKey string            `koanf:"baidu_search_api_key"`
}

// Config 对外暴露的配置结构，包含各模块独立的配置
type Config struct {
	Server   ServerConfig
	Database database.Config
	Auth     auth.Config
	Log      utils.LogConfig
	Agent    agent.Config
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

	// 尝试读取本地覆盖配置
	if err := k.Load(file.Provider("config.local.yaml"), yaml.Parser()); err != nil {
		// 忽略文件不存在的错误，其他错误则打印
		// 由于 koanf file provider 不容易区分文件不存在，这里简单处理，如果加载失败通常就是文件不存在或格式错误
		// 也可以显式检查文件是否存在，但 koanf 设计是松散的
	}

	// 从环境变量读取
	// 从大写转小写 DATABASE_URL database_url
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
		Agent: agent.Config{
			LLMConfigs:        app.LLMConfigs,
			BaiduSearchAPIKey: app.BaiduSearchAPIKey,
		},
	}
}

// GetDebugConfig 返回已加载配置的 Map 视图，并自动对敏感字段进行模糊处理
// 通过将结构体转为 map 并递归遍历实现，无需手动维护字段列表
func (c *Config) GetDebugConfig() map[string]interface{} {
	var m map[string]interface{}
	// 利用 JSON 序列化将结构体转换为 map
	// 注意：Config 结构体字段需要是 Exported 的（首字母大写），这是 Go 的要求，我们已经满足
	b, _ := json.Marshal(c)
	_ = json.Unmarshal(b, &m)

	return maskMap(m)
}

// maskMap 递归遍历 map，对敏感 key 的 value 进行遮蔽
func maskMap(m map[string]interface{}) map[string]interface{} {
	for k, v := range m {
		// 如果值是嵌套的 map（对应嵌套结构体），递归处理
		if innerMap, ok := v.(map[string]interface{}); ok {
			m[k] = maskMap(innerMap)
			continue
		}

		// 检查 key 是否包含敏感词
		if isSensitiveKey(k) {
			m[k] = "***MASKED***"
		}
	}
	return m
}

// isSensitiveKey 判断字段名是否敏感
func isSensitiveKey(key string) bool {
	k := strings.ToLower(key)
	// 敏感关键词列表
	sensitivePatterns := []string{
		"password",
		"secret",
		"apikey",
		"token_key", // 避免误杀 TokenDuration
		"access_key",
	}

	for _, pattern := range sensitivePatterns {
		if strings.Contains(k, pattern) {
			return true
		}
	}

	// 特殊处理：DatabaseURL 虽然不带 password 字样，但也包含敏感信息
	if strings.Contains(k, "databaseurl") || strings.Contains(k, "dsn") {
		return true
	}

	return false
}
