package UI

// SaveInput saves an input to the inputHistory and updates the current index
func (inH *InputHistory) SaveInput(input string) {
	if input != "" {
		inH.inputs = append(inH.inputs, input)
	}

	inH.first = true
	inH.setCurrentHistoryIndex(-1)
}

// setCurrentHistoryIndex sets the current history index on the given pending or if
// it's to large/small on the other end of the slice to make sure it is always a
// correct index
func (inH *InputHistory) setCurrentHistoryIndex(pending int) {
	switch {
	case pending >= len(inH.inputs):
		inH.current = 0
	case pending < 0:
		inH.current = len(inH.inputs) - 1
	default:
		inH.current = pending
	}
}

// checkFirst checks if the autocompleted input history suggestion is
// the first one after the terminal was empty to keep it consitent
func (inH *InputHistory) checkFirst() bool {
	if inH.first == true {
		inH.first = false
		return true
	}
	return false
}
