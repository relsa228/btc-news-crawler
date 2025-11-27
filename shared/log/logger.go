package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init() {
	Log = zap.Must(zap.NewDevelopment(zap.IncreaseLevel(zap.InfoLevel)))
}
