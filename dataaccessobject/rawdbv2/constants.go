package rawdbv2

const (
	NotPruneYetStatus = iota
	WaitingPruneByHeightStatus
	WaitingPruneByHashStatus
	ProcessingPruneStatus
	PausePruneStatus
	FinishPruneStatus
)
