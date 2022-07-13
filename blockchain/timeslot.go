package blockchain

import (
	"github.com/incognitochain/incognito-chain/config"
)

type anchorTime struct {
	PreviousEndTime int64
	StartTime       int64
	StartTimeslot   int64
	Timeslot        int
}

type TSManager struct {
	Anchors []anchorTime
}

func (s *TSManager) getLatestAnchor() anchorTime {
	if len(s.Anchors) == 0 {
		return anchorTime{}
	}
	return s.Anchors[len(s.Anchors)-1]
}

func (s *TSManager) updateNewAnchor(previousEndTime int64, startTime int64, startTS int64, timeslot int) {
	s.Anchors = append(s.Anchors, anchorTime{previousEndTime, startTime, startTS, timeslot})
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
			0, 0, 0, int(config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]),
		}
	}
	rangeTS := int64(0)
	if t > checkpoint.StartTime {
		rangeTS = t - checkpoint.StartTime
	}

	return checkpoint.StartTimeslot + rangeTS/int64(checkpoint.Timeslot)
}
