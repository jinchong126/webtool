package logger

//不完全封装，暂时做不到链式调用封装，保留封装接口
import "github.com/rs/zerolog"

type Logger interface {
	//Trace()interface{}
	//Debug()interface{}
	//Warn()interface{}
	//Info()interface{}
	//Error()interface{}
	//Panic()interface{}
	//Fatal()interface{}
	//不完全封装，暂时做不到链式调用封装
	Trace() *zerolog.Event
	Debug() *zerolog.Event
	Warn() *zerolog.Event
	Info() *zerolog.Event
	Error() *zerolog.Event
	Panic() *zerolog.Event
	Fatal() *zerolog.Event

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Panicf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}
