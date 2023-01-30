package logger

//不完全封装，暂时做不到链式调用封装，保留封装接口
import "github.com/rs/zerolog"

var std Logger // 标准输出

//func Trace() interface{} {
//	return std.Trace()
//}
//func Debug() interface{} {
//	return std.Debug()
//}
//func Info() interface{} {
//	return std.Info()
//}
//func Warn() interface{} {
//	return std.Warn()
//}
//func Error() interface{} {
//	return std.Error()
//}
//func Panic() interface{} {
//	return std.Panic()
//}
//func Fatal() interface{} {
//	return std.Fatal()
//}

func Trace() *zerolog.Event {
	return std.Trace()
}
func Debug() *zerolog.Event {
	return std.Debug()
}
func Info() *zerolog.Event {
	return std.Info()
}
func Warn() *zerolog.Event {
	return std.Warn()
}
func Error() *zerolog.Event {
	return std.Error()
}
func Panic() *zerolog.Event {
	return std.Panic()
}
func Fatal() *zerolog.Event {
	return std.Fatal()
}

func Tracef(format string, args ...interface{}) {
	std.Tracef(format, args...)
}
func Debugf(format string, args ...interface{}) {
	std.Debugf(format, args...)
}
func Infof(format string, args ...interface{}) {
	std.Infof(format, args...)
}
func Warnf(format string, args ...interface{}) {
	std.Warnf(format, args...)
}
func Errorf(format string, args ...interface{}) {
	std.Errorf(format, args...)
}
func Panicf(format string, args ...interface{}) {
	std.Panicf(format, args...)
}
func Fatalf(format string, args ...interface{}) {
	std.Fatalf(format, args...)
}

// 初始化全局日志
func InitGlobal(s Logger) {
	std = s
}

// 初始化全局日志
func GetLogger() *Logger {
	return &std
}
