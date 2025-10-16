package adapters

import (
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/hashicorp/go-hclog"
	"github.com/sirupsen/logrus"
)

type HclogAdapter struct {
	logger *logrus.Logger
	fields []interface{}
	name   string
}

func NewHclogAdapter(logger *logrus.Logger, fields ...interface{}) *HclogAdapter {
	return &HclogAdapter{logger: logger, fields: fields, name: "logrusAdapter"}
}

func (a *HclogAdapter) Log(level hclog.Level, msg string, args ...interface{}) {
	fields := make(logrus.Fields)
	if len(a.fields) > 0 && len(args)%2 != 0 {
		args = append(args, errors.New("adapter: odd number of arguments"))
	}
	for i := 0; i < len(args); i += 2 {
		key, value := args[i], args[i+1]
		fields[fmt.Sprintf("%s", key)] = value
	}
	a.logger.WithFields(fields).Log(toLogrusLevel(level), msg)
}

func (a *HclogAdapter) Trace(msg string, args ...interface{}) {
	a.Log(hclog.Trace, msg, args...)
}

func (a *HclogAdapter) Debug(msg string, args ...interface{}) {
	a.Log(hclog.Debug, msg, args...)
}

func (a *HclogAdapter) Info(msg string, args ...interface{}) {
	a.Log(hclog.Info, msg, args...)
}

func (a *HclogAdapter) Warn(msg string, args ...interface{}) {
	a.Log(hclog.Warn, msg, args...)
}

func (a *HclogAdapter) Error(msg string, args ...interface{}) {
	a.Log(hclog.Error, msg, args...)
}

func (a *HclogAdapter) IsTrace() bool {
	return a.logger.IsLevelEnabled(logrus.TraceLevel)
}

func (a *HclogAdapter) IsDebug() bool {
	return a.logger.IsLevelEnabled(logrus.DebugLevel)
}

func (a *HclogAdapter) IsInfo() bool {
	return a.logger.IsLevelEnabled(logrus.InfoLevel)
}

func (a *HclogAdapter) IsWarn() bool {
	return a.logger.IsLevelEnabled(logrus.WarnLevel)
}

func (a *HclogAdapter) IsError() bool {
	return a.logger.IsLevelEnabled(logrus.ErrorLevel)
}

func (a *HclogAdapter) ImpliedArgs() []interface{} {
	return a.fields
}

func (a *HclogAdapter) With(args ...interface{}) hclog.Logger {
	return NewHclogAdapter(a.logger, append(a.fields, args...)...)
}

func (a *HclogAdapter) Name() string {
	return a.name
}

func (a *HclogAdapter) Named(name string) hclog.Logger {
	a.name = name
	return a
}

func (a *HclogAdapter) ResetNamed(name string) hclog.Logger {
	a.name = "logrusAdapter"
	return a
}

func (a *HclogAdapter) SetLevel(level hclog.Level) {
	a.logger.SetLevel(toLogrusLevel(level))
}

func (a *HclogAdapter) GetLevel() hclog.Level {
	return toHclogLevel(a.logger.GetLevel())
}

func (a *HclogAdapter) StandardLogger(opts *hclog.StandardLoggerOptions) *log.Logger {
	return log.New(a.StandardWriter(opts), "", log.LstdFlags)
}

func (a *HclogAdapter) StandardWriter(opts *hclog.StandardLoggerOptions) io.Writer {
	return a.logger.Out
}

func toLogrusLevel(level hclog.Level) logrus.Level {
	switch level {
	case hclog.Trace:
		return logrus.TraceLevel
	case hclog.Debug:
		return logrus.DebugLevel
	case hclog.Info:
		return logrus.InfoLevel
	case hclog.Warn:
		return logrus.WarnLevel
	case hclog.Error:
		return logrus.ErrorLevel
	}
	return logrus.InfoLevel
}

func toHclogLevel(level logrus.Level) hclog.Level {
	switch level {
	case logrus.TraceLevel:
		return hclog.Trace
	case logrus.DebugLevel:
		return hclog.Debug
	case logrus.InfoLevel:
		return hclog.Info
	case logrus.WarnLevel:
		return hclog.Warn
	case logrus.ErrorLevel:
		return hclog.Error
	}
	return hclog.Info
}
