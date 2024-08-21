package core

type RedisCmd struct {
	ID   int32
	Cmd  string
	Args []string
}

type RedisCmds []*RedisCmd
