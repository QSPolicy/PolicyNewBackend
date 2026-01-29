package utils

import (
	"log"
	"os"
	"path/filepath"
	"policy-backend/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

func InitLogger(cfg *config.Config) {
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

	// 编码器配置
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	var core zapcore.Core

	if cfg.LogFile != "" {
		// 配置 Lumberjack 实现日志切割
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   cfg.LogFile,
			MaxSize:    10,   // 每个日志文件保存10MB
			MaxBackups: 5,    // 保留5个备份
			MaxAge:     30,   // 保留30天
			Compress:   true, // 是否压缩
		})

		// 同时输出到控制台
		consoleWriter := zapcore.AddSync(os.Stdout)

		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.NewMultiWriteSyncer(fileWriter, consoleWriter),
			level,
		)
	} else {
		// 仅输出到控制台
		consoleWriter := zapcore.AddSync(os.Stdout)
		core = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			consoleWriter,
			level,
		)
	}

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
