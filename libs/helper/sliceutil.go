package helper

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
	if middle > intLookFor {
		return RecursiveSearch(slice[:middle], lookFor)
	} else {
		return RecursiveSearch(slice[middle:], lookFor)
	}
}