package loglevels

type LogLevel string

var (
	DebugLevel    LogLevel = "debug"
	WarningLevel LogLevel = "warning"
	InfoLevel    LogLevel = "info"
	ErrorLevel   LogLevel = "error"
)
