package utils

// for exit code
const (
	ExitCodeUnknow = iota
	ExitByOs
	ExitByLogging
	ExitCodeForceUpdate
)

const (
	EmptyString = ""
)

var (
	EmptyBytesSlice   = []byte{}     // DO NOT EDIT OR CHANGE VALUE
	EmptyStringArray  = []string{}   // DO NOT EDIT OR CHANGE VALUE
	EmptyStringMatrix = [][]string{} // DO NOT EDIT OR CHANGE VALUE
)
