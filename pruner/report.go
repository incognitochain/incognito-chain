package pruner

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"time"
)

type ShardPrunerReport struct {
	ChainID              int
	LastTriggerTime      time.Time
	BloomSize            uint64
	Status               string
	Error                string
	LastProcessingHeight uint64
	LastProcessingMode   string
	TotalNodePrune       uint64
	TotalStoragePrune    uint64
}

func (s *ShardPruner) Report() ShardPrunerReport {
	res := ShardPrunerReport{}
	res.LastTriggerTime = s.lastTriggerTime
	switch s.status {
	case IDLE:
		res.Status = "IDLE"
	case PRUNING:
		res.Status = "PRUNING"
	case CHECKING:
		res.Status = "CHECKING"
	}
	res.ChainID = s.shardID
	res.Error = s.lastError
	res.LastProcessingHeight = s.lastProcessingHeight
	res.LastProcessingMode = s.lastProcessingMode
	res.TotalNodePrune = s.nodes
	res.TotalStoragePrune = s.storage
	res.BloomSize = s.bloomSize
	return res
}

func (s *ShardPruner) saveStatus() {
	b, _ := json.Marshal(s.Report())
	rawdbv2.StorePruneStatus(s.db, b)
}

func (s *ShardPruner) restoreStatus() {
	b, err := rawdbv2.GetPruneStatus(s.db)
	if err != nil {
		return
	}
	report := &ShardPrunerReport{}
	err = json.Unmarshal(b, report)
	if err != nil {
		return
	}
	s.lastTriggerTime = report.LastTriggerTime
	s.lastProcessingMode = report.LastProcessingMode
	s.lastError = report.Error
	s.lastProcessingHeight = report.LastProcessingHeight
	s.storage = report.TotalStoragePrune
	s.nodes = report.TotalNodePrune
}
