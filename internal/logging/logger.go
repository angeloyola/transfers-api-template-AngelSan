package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var Logger *zap.SugaredLogger

func init() {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	writer := zapcore.AddSync(os.Stdout)
	level := zapcore.InfoLevel

	core := zapcore.NewCore(consoleEncoder, writer, level)
	Logger = zap.New(core, zap.AddCaller()).Sugar()
}
