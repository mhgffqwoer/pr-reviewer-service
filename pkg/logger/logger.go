package logger

import (
	"os"

	"github.com/mhgffqwoer/pr-service/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var log *zap.SugaredLogger

func InitLogger(cfg *config.LoggingConfig) *zap.SugaredLogger {
	level := zap.InfoLevel
	switch cfg.Level {
	case "debug":
		level = zap.DebugLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	}

	c := zap.NewProductionConfig()
	c.Level = zap.NewAtomicLevelAt(level)
	c.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	if rotation := config.Get().Logging.Rotation; rotation != nil {
		w := zapcore.AddSync(&lumberjack.Logger{
			Filename:   rotation.Name,
			MaxSize:    rotation.MaxSize,
			MaxBackups: rotation.MaxBackups,
			MaxAge:     rotation.MaxAge,
		})

		var syncer zapcore.WriteSyncer
		if rotation.DuplicateToConsole {
			syncer = zapcore.NewMultiWriteSyncer(w, zapcore.AddSync(os.Stderr))
		} else {
			syncer = w
		}

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(c.EncoderConfig),
			syncer,
			c.Level,
		)
		zapLogger := zap.New(core, zap.AddStacktrace(zap.ErrorLevel))
		log = zapLogger.Sugar()
		return log
	}

	zapLogger, _ := c.Build()
	log = zapLogger.Sugar()
	return log
}

func Get() *zap.SugaredLogger {
	if log == nil {
		panic("logger not initialized")
	}
	return log
}
