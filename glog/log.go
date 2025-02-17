package glog

import (
	"fmt"
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func StdError(logContent string) {
	logContent = strings.TrimSpace(logContent)
	os.Stderr.WriteString(fmt.Sprintf("[%s]%s\n", time.Now().Format("2006-01-02 15:04:05"), logContent))
}

func StdInfo(logContent string) {
	logContent = strings.TrimSpace(logContent)
	os.Stdout.WriteString(fmt.Sprintf("[%s]%s\n", time.Now().Format("2006-01-02 15:04:05"), logContent))
}

var gLogger *zap.Logger
var gLogDir string

func initLogger() {

	if gLogger != nil {
		return
	}

	var logFileDir string

	if len(gLogDir) > 0 {
		logFileDir = gLogDir
	}

	if len(logFileDir) < 1 {

		modFilePath, _ := os.Executable()
		if len(modFilePath) < 1 {
			modFilePath = os.Args[0]
		}

		appFilePath, appErr := filepath.Abs(modFilePath)
		if appErr != nil {
			StdError(appErr.Error())
			return
		}
		logFileDir = path.Join(filepath.Dir(appFilePath), "log")
	}

	logFileInfo, pathErr := os.Stat(logFileDir)
	if pathErr != nil {
		os.MkdirAll(logFileDir, 0755)
		logFileInfo, pathErr = os.Stat(logFileDir)
		if pathErr != nil {
			StdError(pathErr.Error())
			return
		}
	}

	if !logFileInfo.IsDir() {
		dirErr := os.MkdirAll(logFileDir, 0755)
		if dirErr != nil {
			StdError(dirErr.Error())
			return
		}
	}

	logFileFormat := path.Join(logFileDir, "app_%Y%m%d.log")

	logHandle, logErr := rotatelogs.New(logFileFormat,
		rotatelogs.WithClock(rotatelogs.Local),
		rotatelogs.WithMaxAge(24*time.Hour))
	if logErr != nil {
		StdError(logErr.Error())
		return
	}

	logConfig := zap.NewProductionEncoderConfig()
	logConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logConfig.EncodeCaller = func(caller zapcore.EntryCaller, encoder zapcore.PrimitiveArrayEncoder) {
		timePath := caller.TrimmedPath()
		timePath = strings.ReplaceAll(timePath, ".go:", ".cpp:")
		encoder.AppendString(timePath)
	}

	logConfig.EncodeLevel = func(level zapcore.Level, encoder zapcore.PrimitiveArrayEncoder) {
		encoder.AppendString("[" + level.CapitalString() + "]")
	}

	logEncoder := zapcore.NewConsoleEncoder(logConfig)

	logInfoLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel
	})

	logWarnLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel
	})

	logErrorLevel := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel
	})

	var logCore zapcore.Core

	logMode := os.Getenv("glog_run_mode")

	if strings.EqualFold(logMode, "release") {
		logCore = zapcore.NewTee(
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logInfoLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logWarnLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logErrorLevel),
		)

	} else {
		logCore = zapcore.NewTee(
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logInfoLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logWarnLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(logHandle), logErrorLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(os.Stdout), logInfoLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(os.Stderr), logWarnLevel),
			zapcore.NewCore(logEncoder, zapcore.AddSync(os.Stderr), logErrorLevel),
		)
	}

	gLogger = zap.New(logCore,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
	)

	if gLogger != nil {
		zap.ReplaceGlobals(gLogger)
	}
}

func SetLogDir(dirPath string) {
	gLogDir = dirPath
}

func Info(args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprint(args...)
	gLogger.Info(logData)
}

func InfoF(format string, args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprintf(format, args...)
	gLogger.Info(logData)
}

func Warn(args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprint(args...)
	gLogger.Warn(logData)
}

func WarnF(format string, args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprintf(format, args...)
	gLogger.Warn(logData)
}

func Error(args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprint(args...)
	gLogger.Error(logData)
}

func ErrorF(format string, args ...interface{}) {

	initLogger()
	if gLogger == nil {
		return
	}

	logData := fmt.Sprintf(format, args...)
	gLogger.Error(logData)
}
