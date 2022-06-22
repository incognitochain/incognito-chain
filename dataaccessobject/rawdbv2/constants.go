package rawdbv2

const (
	NotPruneYetStatus = byte(iota)
	WaitingPruneByHeightStatus
	WaitingPruneByHashStatus
	ProcessingPruneByHeightStatus
	ProcessingPruneByHashStatus
	FinishPruneStatus
)
