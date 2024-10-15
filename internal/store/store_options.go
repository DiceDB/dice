package store

type PutOptions struct {
	KeepTTL bool
	PutCmd  string
}

func getDefaultPutOptions() *PutOptions {
	return &PutOptions{
		KeepTTL: false,
		PutCmd:  Set,
	}
}

type PutOption func(*PutOptions)

func WithKeepTTL(value bool) PutOption {
	return func(po *PutOptions) {
		po.KeepTTL = value
	}
}

func WithPutCmd(cmd string) PutOption {
	return func(po *PutOptions) {
		po.PutCmd = cmd
	}
}

type DelOptions struct {
	DelCmd string
}

func getDefaultDelOptions() *DelOptions {
	return &DelOptions{
		DelCmd: Del,
	}
}

type DelOption func(*DelOptions)

func WithDelCmd(cmd string) DelOption {
	return func(po *DelOptions) {
		po.DelCmd = cmd
	}
}
