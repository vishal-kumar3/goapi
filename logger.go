package main

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func InitLogger() {
	_ = os.MkdirAll("./logs", 0o755)

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	consoleEncoder := zapcore.NewJSONEncoder(encCfg)

	logFile, err := os.OpenFile("./logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		panic(err)
	}

	fileWS := zapcore.AddSync(logFile)
	consoleWS := zapcore.AddSync(os.Stdout)

	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, consoleWS, zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, fileWS, zap.InfoLevel),
	)

	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

func SyncLogger() {
	if Log != nil {
		_ = Log.Sync()
	}
}
