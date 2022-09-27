package core

var KeyspaceStat [4]map[string]int

func UpdateDBStat(num int, metric string, value int) {
	KeyspaceStat[num][metric] = value
}
