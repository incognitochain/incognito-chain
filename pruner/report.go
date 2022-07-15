package pruner

import (
	"time"
)

type ShardPrunerReport struct {
	LastTriggerTime      time.Time
	BloomSize            uint64
	Status               string
	Error                string
	LastProcessingHeight uint64
	TotalNodePrune       uint64
	TotalStoragePrune    uint64
}

func (s *ShardPruner) Report() ShardPrunerReport {
	res := ShardPrunerReport{}
	res.LastTriggerTime = s.lastTriggerTime
	switch s.status {
	case IDLE:
		res.Status = "IDLE"
	case RUNNING:
		res.Status = "RUNNING"
	case ERROR:
		res.Status = "ERROR"
	}

	res.Error = s.lastError
	res.LastProcessingHeight = s.lastProcessingHeight
	res.TotalNodePrune = s.nodes
	res.TotalStoragePrune = s.storage
	res.BloomSize = s.bloomSize
	return res
}
