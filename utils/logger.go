package utils

import (
	"log"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// InitLogger 初始化日志，使用模块化配置
func InitLogger(cfg *LogConfig) {
	// 确保日志目录存在
	if cfg.LogFile != "" {
		logDir := filepath.Dir(cfg.LogFile)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			_ = os.MkdirAll(logDir, 0755)
		}
	}

	// 设置日志级别
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	// 编码器配置（文件用 JSON）
	jsonEncoderConfig := zap.NewProductionEncoderConfig()
	jsonEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	jsonEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 创建多个 Core
	var cores []zapcore.Core

	// 1. 控制台 Core (使用自定义 PtermCore 做富文本渲染)
	// NewPtermCore 实现了 zapcore.Core 接口，拦截 Write 操作并用 pterm 格式化
	consoleCore := NewPtermCore(level)
	cores = append(cores, consoleCore)

	// 2. 文件 Core (JSON Encoder)
	if cfg.LogFile != "" {
		// 配置 Lumberjack 实现日志切割
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    10,   // 每个日志文件保存10MB
			MaxBackups: 5,    // 保留5个备份
			MaxAge:     30,   // 保留30天
			Compress:   true, // 是否压缩
		})

		fileCore := zapcore.NewCore(
			zapcore.NewJSONEncoder(jsonEncoderConfig),
			fileWriter,
			level,
		)
		cores = append(cores, fileCore)
	}

	// 使用 NewTee 将多个 Core 合并
	core := zapcore.NewTee(cores...)

	// 创建 Logger，添加调用者信息
	Log = zap.New(core, zap.AddCaller())

	// 替换全局 Logger
	zap.ReplaceGlobals(Log)

	// 重定向标准库 log 输出到 zap
	// 这样 log.Println, log.Printf 都会被 zap 接管并输出到日志文件
	zap.RedirectStdLog(Log)

	// 移除标准库 log 的默认前缀（如时间戳），因为 Zap 会自己添加
	log.SetFlags(0)
}
