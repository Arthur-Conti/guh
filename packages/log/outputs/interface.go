package outputs

import loglevels "github.com/Arthur-Conti/guh/packages/log/log_levels"

type OutputInterface interface {
	Log(string, loglevels.LogLevel, string)
	Logf(string,  loglevels.LogLevel, string, ...any)
}