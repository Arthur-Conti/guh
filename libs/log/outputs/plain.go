package outputs

import (
	"fmt"

	loglevels "github.com/Arthur-Conti/guh/libs/log/log_levels"
)

type PlainOutput struct {
	opts PlainOutputOpts
}

type PlainOutputOpts struct {
	DebugPattern   string
	WarningPattern string
	InfoPattern    string
	ErrorPattern   string
}

func NewPlainOutput(opts PlainOutputOpts) *PlainOutput {
	return &PlainOutput{
		opts: opts,
	}
}

func (po *PlainOutput) Log(applicationPackage string, level loglevels.LogLevel, message string) {
	switch level {
	case loglevels.DebugLevel:
		fmt.Println(applicationPackage + po.opts.DebugPattern + message)
	case loglevels.WarningLevel:
		fmt.Println(applicationPackage + po.opts.WarningPattern + message)
	case loglevels.InfoLevel:
		fmt.Println(applicationPackage + po.opts.InfoPattern + message)
	case loglevels.ErrorLevel:
		fmt.Println(applicationPackage + po.opts.ErrorPattern + message)
	}
}

func (po *PlainOutput) Logf(applicationPackage string, level loglevels.LogLevel, message string, vals ...any) {
	backSlashN := "\n"
	switch level {
	case loglevels.DebugLevel:
		fmt.Printf(applicationPackage+po.opts.DebugPattern+message+backSlashN, vals...)
	case loglevels.WarningLevel:
		fmt.Printf(applicationPackage+po.opts.WarningPattern+message+backSlashN, vals...)
	case loglevels.InfoLevel:
		fmt.Printf(applicationPackage+po.opts.InfoPattern+message+backSlashN, vals...)
	case loglevels.ErrorLevel:
		fmt.Printf(applicationPackage+po.opts.ErrorPattern+message+backSlashN, vals...)
	}
}
