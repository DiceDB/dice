package core

type RedisCmd struct {
	ID   uint32
	Cmd  string
	Args []string
}

type RedisCmds []*RedisCmd
