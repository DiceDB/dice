// This file is part of DiceDB.
// Copyright (C) 2024 DiceDB (dicedb.io).
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

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
