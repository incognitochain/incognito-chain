package blockchain

import (
	"github.com/incognitochain/incognito-chain/config"
)

type anchorTime struct {
	PreviousEndTime int64
	StartTime       int64
	StartTimeslot   int64
	Timeslot        int
	Feature         string
	BlockHeight     uint64
}

type TSManager struct {
	Anchors []anchorTime
	//below fields save the last enablefeature block for shard
	//when we see update version in shardstate, these value help us retrieve checkpoint information
	CurrentBlockVersion int
	CurrentBlockTS      int64
	CurrentProposeTime  int64
}

func (s *TSManager) getLatestAnchor() anchorTime {
	if len(s.Anchors) == 0 {
		return anchorTime{}
	}
	return s.Anchors[len(s.Anchors)-1]
}

func (s *TSManager) updateNewAnchor(previousEndTime int64, startTime int64, startTS int64, timeslot int, feature string, blockHeight uint64) {
	s.Anchors = append(s.Anchors, anchorTime{previousEndTime, startTime, startTS, timeslot, feature, blockHeight})
}

func (s *TSManager) updateCurrentInfo(version int, currentTS int64, currentProposeTime int64) {
	s.CurrentBlockVersion = version
	s.CurrentBlockTS = currentTS
	s.CurrentProposeTime = currentProposeTime
}

func (s *TSManager) getCurrentTS() int64 {
	lastAnchor := s.getLatestAnchor()
	if lastAnchor.Timeslot == 0 {
		return config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	}
	return int64(lastAnchor.Timeslot)
}

func (s *TSManager) calculateTimeslot(t int64) int64 {
	checkpoint := s.getLatestAnchor()
	for i := len(s.Anchors) - 1; i >= 0; i-- {
		if t >= s.Anchors[i].PreviousEndTime {
			checkpoint = s.Anchors[i]
			break
		}
	}

	if checkpoint.Timeslot == 0 {
		checkpoint = anchorTime{
			0, 0, 0, int(config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]), BLOCKTIME_DEFAULT, 1,
		}
	}
	rangeTS := int64(0)
	if t > checkpoint.StartTime {
		rangeTS = t - checkpoint.StartTime
	}

	return checkpoint.StartTimeslot + rangeTS/int64(checkpoint.Timeslot)
}
