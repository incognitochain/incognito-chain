package rawdbv2

const (
	NotPruneYetStatus = iota
	WaitingPruneByHeightStatus
	WaitingPruneByHashStatus
	ProcessingPruneByHeightStatus
	ProcessingPruneByHashStatus
	PausePruneByHeightStatus
	PausePruneByHashStatus
	FinishPruneStatus
)
