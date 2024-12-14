// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
