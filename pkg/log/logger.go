package log

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

// Init 初始化全局 Zap Logger
func Init(mode string) {
	var err error
	if mode == "release" {
		Logger, err = zap.NewProduction()
	} else {
		Logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic("初始化 Zap Logger 失败: " + err.Error())
	}
}

// Sync 刷新缓冲区
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Info 输出 Info 级别日志
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Error 输出 Error 级别日志
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal 输出 Fatal 级别日志并退出
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}
