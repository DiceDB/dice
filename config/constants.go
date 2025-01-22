package config

import "time"

const (
	IOBufferLength                   int           = 512
	EvictionRatio                    float64       = 0.9
	DefaultKeysLimit                 int           = 200000000
	WatchChanBufSize                 int           = 20000
	ShardCronFrequency               time.Duration = 1 * time.Second
	AdhocReqChanBufSize              int           = 20
	EnableProfile                    bool          = false
	WebSocketWriteResponseTimeout    time.Duration = 10 * time.Second
	WebSocketMaxWriteResponseRetries int           = 3

	KeepAlive int32 = 300
	Timeout   int32 = 300
)
