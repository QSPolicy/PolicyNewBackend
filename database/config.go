package database

// Config 数据库模块配置
type Config struct {
	DatabaseURL       string `koanf:"database_url"`
	MySQLMaxIdleConns int    `koanf:"mysql_max_idle_conns"`
	MySQLMaxOpenConns int    `koanf:"mysql_max_open_conns"`
}

// DefaultConfig 返回数据库模块的默认配置
func DefaultConfig() Config {
	return Config{
		DatabaseURL:       "sqlite3://policy.db",
		MySQLMaxIdleConns: 10,
		MySQLMaxOpenConns: 100,
	}
}
