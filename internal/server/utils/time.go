// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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

func (mc *MockClock) GetTime() time.Time {
	return mc.CurrTime
}

func GetCurrentTime() time.Time {
	return CurrentTime.Now()
}

func AddSecondsToUnixEpoch(second int64) int64 {
	return GetCurrentTime().Unix() + second
}
