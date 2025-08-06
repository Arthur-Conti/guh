package logger

import (
	applicationpackage "github.com/Arthur-Conti/guh/packages/log/application_package"
	loglevels "github.com/Arthur-Conti/guh/packages/log/log_levels"
	"github.com/Arthur-Conti/guh/packages/log/outputs"
)

var levelParse = map[string]int{
	"debug":   1,
	"warning": 2,
	"info":    3,
	"error":   4,
}

type Logger struct {
	opts LoggerOpts
}

type LoggerOpts struct {
	OutputType          outputs.OutputInterface
	SecondaryOutputType outputs.OutputInterface
	Level               int
	LevelStr            string
	ApplicationPackage  applicationpackage.PackageLevel
}

func NewLogger(opts LoggerOpts) *Logger {
	opts.Level = levelParse[opts.LevelStr]
	return &Logger{
		opts: opts,
	}
}

func (l *Logger) Debug(message LogMessage) {
	if l.opts.Level > 1 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Log(packageName, loglevels.DebugLevel, message.Message)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Log(packageName, loglevels.DebugLevel, message.Message)
	}
}

func (l *Logger) Debugf(message LogMessage) {
	if l.opts.Level > 1 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Logf(packageName, loglevels.DebugLevel, message.Message, message.Vals...)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Logf(packageName, loglevels.DebugLevel, message.Message, message.Vals...)
	}
}

func (l *Logger) Warning(message LogMessage) {
	if l.opts.Level > 2 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Log(packageName, loglevels.WarningLevel, message.Message)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Log(packageName, loglevels.WarningLevel, message.Message)
	}
}

func (l *Logger) Warningf(message LogMessage) {
	if l.opts.Level > 2 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Logf(packageName, loglevels.WarningLevel, message.Message, message.Vals...)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Logf(packageName, loglevels.WarningLevel, message.Message, message.Vals...)
	}
}

func (l *Logger) Info(message LogMessage) {
	if l.opts.Level > 3 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Log(packageName, loglevels.InfoLevel, message.Message)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Log(packageName, loglevels.InfoLevel, message.Message)
	}
}

func (l *Logger) Infof(message LogMessage) {
	if l.opts.Level > 3 {
		return
	}
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Logf(packageName, loglevels.InfoLevel, message.Message, message.Vals...)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Logf(packageName, loglevels.InfoLevel, message.Message, message.Vals...)
	}
}

func (l *Logger) Error(message LogMessage) {
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Log(packageName, loglevels.ErrorLevel, message.Message)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Log(packageName, loglevels.ErrorLevel, message.Message)
	}
}

func (l *Logger) Errorf(message LogMessage) {
	packageName := l.opts.ApplicationPackage.Style(message.ApplicationPackage)
	l.opts.OutputType.Logf(packageName, loglevels.ErrorLevel, message.Message, message.Vals...)
	if l.opts.SecondaryOutputType != nil {
		l.opts.SecondaryOutputType.Logf(packageName, loglevels.ErrorLevel, message.Message, message.Vals...)
	}
}
