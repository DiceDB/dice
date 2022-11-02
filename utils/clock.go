package utils

import "time"

func GetCurrentClock() uint32 {
	return uint32(time.Now().Unix()) & 0x00FFFFFF
}
