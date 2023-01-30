package logger

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"webtool/pkg/define"

	"github.com/rs/zerolog"
)

const (
	//默认控制台输出
	// LogFileDir 文件路径
	LogFileDir = "/var/log"
	// LogFileMaxSize 每个日志文件最大 MB
	LogFileMaxSize = 100
	// LogFileMaxBackups 保留日志文件个数
	LogFileMaxBackups = 5
	// LogFileMaxAge 保留日志最大天数
	LogFileMaxAge = 30

	// LogPeriod LogSampled 配置: 每 1 秒最多输出 3 条日志
	LogPeriod = time.Second
	LogBurst  = 3

	// LogLevel 日志级别: -1Trace 0Debug 1Info 2Warn 3Error(默认) 4Fatal 5Panic 6NoLevel 7Off
	LogLevel = zerolog.DebugLevel
)

// 文件名不为空，输入到文件
func Init(FileName string) {

	if err := InitLogger(FileName); err != nil {
		os.Exit(1)
	}

	// 路径脱敏
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}

}

// InitLogger 配置热加载等场景调用, 重载日志环境
func InitLogger(FileName string) error {

	basicLog := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: define.TimeFormat}

	if len(FileName) > 0 {
		//切换文件输出
		println(LogFileDir[len(LogFileDir)-1])
		file := ""
		if LogFileDir[len(LogFileDir)-1] == '/' {
			file = LogFileDir + FileName
		} else {
			file = LogFileDir + "/" + FileName
		}

		println("log file:%s", file)
		logFileLocation, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
		if err != nil {
			return err
		}
		basicLog.Out = logFileLocation
		basicLog.NoColor = true
	} else {
		//控制台默认开启颜色
		basicLog.NoColor = false
	}

	logger := zerolog.New(basicLog).With().Timestamp().Caller().Logger()
	logger = logger.Level(LogLevel)

	//初始化前设置好日志属性
	InitGlobal(NewZeroLogger(logger))
	return nil
}

type zeroLogger struct {
	l *zerolog.Logger
}

func NewZeroLogger(l zerolog.Logger) Logger {
	return &zeroLogger{
		l: &l,
	}
}

func (this *zeroLogger) Trace() *zerolog.Event {
	return this.l.Trace()
}

func (this *zeroLogger) Debug() *zerolog.Event {
	return this.l.Debug()
}

func (this *zeroLogger) Warn() *zerolog.Event {
	return this.l.Warn()
}

func (this *zeroLogger) Info() *zerolog.Event {
	return this.l.Info()
}

func (this *zeroLogger) Error() *zerolog.Event {
	return this.l.Error()
}

func (this *zeroLogger) Panic() *zerolog.Event {
	return this.l.Panic()
}

func (this *zeroLogger) Fatal() *zerolog.Event {
	return this.l.Fatal()
}

func (this *zeroLogger) Tracef(format string, a ...interface{}) {
	this.l.Trace().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Debugf(format string, a ...interface{}) {
	this.l.Debug().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Warnf(format string, a ...interface{}) {
	this.l.Warn().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Infof(format string, a ...interface{}) {
	this.l.Info().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Errorf(format string, a ...interface{}) {
	this.l.Error().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Panicf(format string, a ...interface{}) {
	this.l.Panic().CallerSkipFrame(2).Msgf(format, a...)
}

func (this *zeroLogger) Fatalf(format string, a ...interface{}) {
	this.l.Fatal().CallerSkipFrame(2).Msgf(format, a...)
}
