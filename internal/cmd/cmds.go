package cmd

type RedisCmd struct {
	RequestID uint32
	Cmd       string
	Args      []string
}

type RedisCmds struct {
	Cmds      []*RedisCmd
	RequestID uint32
}
