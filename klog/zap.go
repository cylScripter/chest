package klog

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"io"
)

type ZapLogger struct {
	log     *zap.Logger
	level   Level
	options []zap.Option // zap 配置项
}

func (ll *ZapLogger) SetOption(opts ...zap.Option) {
	ll.options = append(ll.options, opts...)
	ll.log.WithOptions(ll.options...)
}

func (ll *ZapLogger) Fatal(v ...interface{}) {
}

func (ll *ZapLogger) SetOutput(w io.Writer) {

}

func (ll *ZapLogger) SetLevel(lv Level) {

}

func (ll *ZapLogger) Error(v ...interface{}) {
}

func (ll *ZapLogger) Warn(v ...interface{}) {

}

func (ll *ZapLogger) Notice(v ...interface{}) {

}

func (ll *ZapLogger) Info(v ...interface{}) {
	sugar := ll.log.Sugar()
	sugar.Info(v...)

}

func (ll *ZapLogger) Debug(v ...interface{}) {
	ll.log.Debug(fmt.Sprint(v...))
}

func (ll *ZapLogger) Trace(v ...interface{}) {

}

func (ll *ZapLogger) Fatalf(format string, v ...interface{}) {

}

func (ll *ZapLogger) Errorf(format string, v ...interface{}) {

}

func (ll *ZapLogger) Warnf(format string, v ...interface{}) {

}

func (ll *ZapLogger) Noticef(format string, v ...interface{}) {

}

func (ll *ZapLogger) Infof(format string, v ...interface{}) {
	ll.log.Info(fmt.Sprintf(format, v...))
}

func (ll *ZapLogger) Debugf(format string, v ...interface{}) {

}

func (ll *ZapLogger) Tracef(format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxFatalf(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxErrorf(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxWarnf(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxNoticef(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxInfof(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxDebugf(ctx context.Context, format string, v ...interface{}) {

}

func (ll *ZapLogger) CtxTracef(ctx context.Context, format string, v ...interface{}) {

}

func NewZapLogger() *ZapLogger {
	loggers, _ := zap.NewDevelopment()
	return &ZapLogger{
		log:     loggers,
		level:   LevelDebug,
		options: []zap.Option{},
	}
}
