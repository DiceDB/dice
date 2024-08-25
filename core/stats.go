package core

import (
	"sync"
)

type KeyspaceStat struct {
	stats map[string]int
	mutex sync.Mutex
}

func CreateKeyspaceStat() *KeyspaceStat {
	keyspace := &KeyspaceStat{
		stats: make(map[string]int),
	}
	return keyspace
}

func (ks *KeyspaceStat) UpdateDBStat(metric string, value int) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	ks.stats[metric] = value
}

func (ks *KeyspaceStat) IncrStat(metric string) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	ks.stats[metric]++
}

func (ks *KeyspaceStat) DecrKeys(metric string) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	ks.stats[metric]--
}

func (ks *KeyspaceStat) GetStat(metric string) int {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()
	keyCount := ks.stats[metric]
	return keyCount
}
