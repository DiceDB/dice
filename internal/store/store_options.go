// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

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
