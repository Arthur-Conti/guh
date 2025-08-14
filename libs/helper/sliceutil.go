package helper

import (
	"github.com/Arthur-Conti/guh/config"
	"github.com/Arthur-Conti/guh/libs/log/logger"
)

func Contains(slice []any, lookFor any, isSorted bool) bool {
	if isSorted {
		return RecursiveSearch(slice, lookFor)
	} else {
		return NormalSearch(slice, lookFor)
	}
}

func NormalSearch(slice []any, lookFor any) bool {
	for _, item := range slice {
		if item == lookFor {
			return true
		}
	}
	return false
}

func RecursiveSearch(slice []any, lookFor any) bool {
	middle := len(slice) / 2
	intLookFor, ok := lookFor.(int)
	if !ok {
		return false
	}
	if slice[middle] == intLookFor {
		return true
	}
	if intLookFor < slice[0].(int) || intLookFor > len(slice) {
		return false
	}
	config.Config.Logger.Debugf(logger.LogMessage{Message: "%v", Vals: []any{slice[middle]}})
	if middle > intLookFor {
		return RecursiveSearch(slice[:middle], lookFor)
	} else {
		return RecursiveSearch(slice[middle:], lookFor)
	}
}