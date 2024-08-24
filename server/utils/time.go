package utils

import "time"

type (
	Clock interface {
		Now() time.Time
	}
	RealClock struct{}
	MockClock struct {
		CurrTime time.Time
	}
)

var (
	CurrentTime Clock = RealClock{}
)

func (RealClock) Now() time.Time {
	return time.Now()
}

func (mc MockClock) Now() time.Time {
	return mc.CurrTime
}

func (mc *MockClock) SetTime(t time.Time) {
	mc.CurrTime = t
}

func GetCurrentTime() time.Time {
	return CurrentTime.Now()
}

func AddSecondsToUnixEpoch(second int64) uint64 {
	return uint64(GetCurrentTime().Unix()) + uint64(second)
}
