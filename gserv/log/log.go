package log

import (
	"fmt"
	mlog "github.com/ngaut/log"
	"github.com/sirupsen/logrus"
	"io"
	stdlog "log"
	"os"
	"runtime"
)

var (
	logger = logrus.New()
)

func init() {
	//	logrus.SetOutput()
	logrus.SetFormatter(&logrus.TextFormatter{})
	//logger.WriterLevel(logrus.DebugLevel)
	logger.Level = logrus.DebugLevel
}

func fileLineEntry() *logrus.Entry {
	_, file, line, succ := runtime.Caller(2)

	if !succ {
		return logger.WithFields(logrus.Fields{})
	}
	return logger.WithFields(logrus.Fields{
		"file": file,
		"line": line,
	})
}

func SetLevel(levelStr string) (err error) {
	switch levelStr {
	case "debug":
		logger.Level = logrus.DebugLevel
		mlog.SetLevel(mlog.LOG_LEVEL_DEBUG)
		break
	case "info":
		logger.Level = logrus.InfoLevel
		mlog.SetLevel(mlog.LOG_LEVEL_INFO)
		break
	case "warn":
		logger.Level = logrus.WarnLevel
		mlog.SetLevel(mlog.LOG_LEVEL_WARN)
		break
	default:
		logger.Fatalln("log.SetLevel Error,log conf only allow [info, debug, warn], please check your confguire", err)
	}

	return
}
func SetOutput(out io.Writer) {
	logger.Out = out
	stdlog.SetOutput(out)
	mlog.SetOutput(out)
}

func Debug(args ...interface{}) {
	fileLineEntry().Debug(args...)
}

func Debugln(args ...interface{}) {
	fileLineEntry().Debugln(args...)
}

func Debugf(format string, args ...interface{}) {
	fileLineEntry().Debugf(format, args...)
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infoln(args ...interface{}) {
	logger.Infoln(args...)
}

func Infof(format string, args ...interface{}) {
	logger.Infof(format, args...)
}

func Warn(args ...interface{}) {
	fileLineEntry().Warn(args...)
}

func Warnln(args ...interface{}) {
	fileLineEntry().Warnln(args...)
}

func Warnf(format string, args ...interface{}) {
	fileLineEntry().Warnf(format, args...)
}

func Error(args ...interface{}) {
	fileLineEntry().Error(args...)
}

func Errorln(args ...interface{}) {
	fileLineEntry().Errorln(args...)
}

func Errorf(format string, args ...interface{}) {
	fileLineEntry().Errorf(format, args...)
}

func Fatal(args ...interface{}) {
	fileLineEntry().Fatal(args...)
}

func Fatalln(args ...interface{}) {
	fileLineEntry().Fatalln(args...)
}

func Fatalf(format string, args ...interface{}) {
	fileLineEntry().Fatalf(format, args...)
}

func SetLevelAndFile(fileName string, level string) {
	logFile, logErr := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if logErr != nil {
		fmt.Println("FatalError", logErr)
		Fatalln("Fail to find", *logFile, "cServer start Failed")
	}
	SetOutput(logFile)

	SetLevel(level)
}
