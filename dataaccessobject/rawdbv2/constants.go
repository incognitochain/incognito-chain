package rawdbv2

const (
	NotPruneYetStatus = byte(iota)
	WaitingPruneByHeightStatus
	WaitingPruneByHashStatus
	ProcessingPruneStatus
	PausePruneStatus
	FinishPruneStatus
)
