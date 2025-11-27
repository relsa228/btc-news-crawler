package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init() {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      true,
		Encoding:         "console",
		OutputPaths:      []string{"stdout", "app.log"},
		ErrorOutputPaths: []string{"stderr", "app.log"},
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
	}
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	Log = logger
}
