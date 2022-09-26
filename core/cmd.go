package core

type RedisCmd struct {
	Cmd  string
	Args []string
}

type RedisCmds []*RedisCmd
