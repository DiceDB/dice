package cmd

type DiceDBCmd struct {
	RequestID uint32
	Cmd       string
	Args      []string
}

type RedisCmds struct {
	Cmds      []*DiceDBCmd
	RequestID uint32
}
