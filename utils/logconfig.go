package utils

// LogConfig 日志模块配置
type LogConfig struct {
	LogLevel string `koanf:"log_level"`
	LogFile  string `koanf:"log_file"`
}

// DefaultLogConfig 返回日志模块的默认配置
func DefaultLogConfig() LogConfig {
	return LogConfig{
		LogLevel: "info",
		LogFile:  "logs/app.log",
	}
}
