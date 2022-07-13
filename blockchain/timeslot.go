package blockchain

import (
	"github.com/incognitochain/incognito-chain/config"
)

type anchorTime struct {
	StartTime     int64
	StartTimeslot int64
	Timeslot      int
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

func (s *TSManager) updateNewAnchor(startTime int64, startTS int64, timeslot int) {
	s.Anchors = append(s.Anchors, anchorTime{startTime, startTS, timeslot})
}

func (s *TSManager) getCurrentTS() int64 {
	lastAnchor := s.getLatestAnchor()
	if lastAnchor.Timeslot == 0 {
		return config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]
	}
	return int64(lastAnchor.Timeslot)
}

func (s *TSManager) calculateTimeslot(t int64) int64 {
	lastAnchor := s.getLatestAnchor()
	if lastAnchor.Timeslot == 0 {
		lastAnchor = anchorTime{
			0, 0, int(config.Param().BlockTimeParam[BLOCKTIME_DEFAULT]),
		}
	}
	rangeTS := t - lastAnchor.StartTime
	return lastAnchor.StartTimeslot + rangeTS/int64(lastAnchor.Timeslot)
}
