package devframework

import (
	"time"
)

type TimeEngine struct {
	time time.Time
}

func NewTimeEngine() *TimeEngine {
	timer := &TimeEngine{}
	return timer
}

func (s *TimeEngine) init(startTime int64) {
	s.time = time.Unix(startTime, 0)
}

func (s *TimeEngine) Now() int64 {
	return s.time.Unix()
}

func (s *TimeEngine) Forward(sec int64) {
	s.time = s.time.Add(time.Duration(sec) * time.Second)
}
