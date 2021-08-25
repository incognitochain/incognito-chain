package common

import (
	"math"
)

func CalculateTimeSlot(time int64) int64 {
	return int64(math.Floor(float64(time / int64(TIMESLOT))))
}
