// Copyright (c) 2022-present, DiceDB contributors
// All rights reserved. Licensed under the BSD 3-Clause License. See LICENSE file in the project root for full license information.

package ironhawk

import (
	"testing"
	"time"

	"github.com/dicedb/dicedb-go/wire"
	"gotest.tools/v3/assert"
)

func TestGetDel(t *testing.T) {
	client := getLocalConnection()
	defer client.Close()

	testCases := []struct {
		name   string
		cmds   []*wire.Command
		expect []*wire.Response
		delays []time.Duration
	}{
		{
			name:   "GetDel",
			cmds:   []*wire.Command{{Cmd: "SET", Args: []string{"k", "v"}}, {Cmd: "GETDEL", Args: []string{"k"}}, {Cmd: "GETDEL", Args: []string{"k"}}, {Cmd: "GET", Args: []string{"k"}}},
			expect: []*wire.Response{{Value: &wire.Response_VStr{VStr: "OK"}}, {Value: &wire.Response_VStr{VStr: "v"}}, {Value: &wire.Response_VNil{VNil: true}}, {Value: &wire.Response_VNil{VNil: true}}},
			delays: []time.Duration{0, 0, 0, 0},
		},
		{
			name: "GetDel with expiration, checking if key exist and is already expired, then it should return null",
			cmds: []*wire.Command{
				{Cmd: "GETDEL", Args: []string{"k"}},
				{Cmd: "SET", Args: []string{"k", "v", "EX", "2"}},
				{Cmd: "GETDEL", Args: []string{"k"}},
			},
			expect: []*wire.Response{
				{Value: &wire.Response_VNil{VNil: true}},
				{Value: &wire.Response_VStr{VStr: "OK"}},
				{Value: &wire.Response_VNil{VNil: true}},
			},
			delays: []time.Duration{0, 0, 3 * time.Second},
		},
		{
			name: "GetDel with expiration, checking if key exist and is not yet expired, then it should return its value",
			cmds: []*wire.Command{
				{Cmd: "SET", Args: []string{"k", "v", "EX", "40"}},
				{Cmd: "GETDEL", Args: []string{"k"}},
			},
			expect: []*wire.Response{
				{Value: &wire.Response_VStr{VStr: "OK"}},
				{Value: &wire.Response_VStr{VStr: "v"}},
			},
			delays: []time.Duration{0, 2 * time.Second},
		},
		{
			name: "GetDel with invalid command",
			cmds: []*wire.Command{
				{Cmd: "GETDEL", Args: []string{}},
				{Cmd: "GETDEL", Args: []string{"k", "v"}},
			},
			expect: []*wire.Response{
				{},
				{},
			},
			delays: []time.Duration{0, 0},
		},
		// TODO: remove these after confirming if needed
		// {
		// name: "Getdel with value created from Setbit",
		// cmds: []*wire.Command{
		// &wire.Command{Cmd: "SETBIT", Args: []string{"k", "1", "1"}},
		// &wire.Command{Cmd: "GET", Args: []string{"k"}},
		// &wire.Command{Cmd: "GETDEL", Args: []string{"k"}},
		// &wire.Command{Cmd: "GET", Args: []string{"k"}},
		// },
		// expect: []*wire.Response{
		// &wire.Response{Value: &wire.Response_VInt{VInt: 0}},
		// &wire.Response{Value: &wire.Response_VStr{VStr: "@"}},
		// &wire.Response{Value: &wire.Response_VStr{VStr: "@"}},
		// &wire.Response{Value: &wire.Response_VNil{VNil: true}},
		// },
		// delays: []time.Duration{0, 0, 0, 0},
		// },
		// {
		// name: "GetDel with Set object should return wrong type error",
		// cmds: []*wire.Command{
		// &wire.Command{Cmd: "SADD", Args: []string{"myset", "member1"}},
		// &wire.Command{Cmd: "GETDEL", Args: []string{"myset"}},
		// },
		// expect: []*wire.Response{
		// &wire.Response{Value: &wire.Response_VInt{VInt: 1}},
		// &wire.Response{},
		// // &wire.Response{Value: &wire.Response_VErr{VErr: "WRONGTYPE Operation against a key holding the wrong kind of value"}},
		// },
		// delays: []time.Duration{0, 0},
		// },
		// {
		// name: "GetDel with JSON object should return wrong type error",
		// cmds: []*wire.Command{
		// &wire.Command{Cmd: "JSON.SET", Args: []string{"k", "$", "1"}},
		// &wire.Command{Cmd: "GETDEL", Args: []string{"k"}},
		// &wire.Command{Cmd: "JSON.GET", Args: []string{"k"}},
		// },
		// expect: []*wire.Response{
		// &wire.Response{Value: &wire.Response_VStr{VStr: "OK"}},
		// &wire.Response{},
		// // &wire.Response{Value: &wire.Response_VErr{VErr: "WRONGTYPE Operation against a key holding the wrong kind of value"}},
		// &wire.Response{Value: &wire.Response_VStr{VStr: "1"}},
		// },
		// delays: []time.Duration{0, 0, 0},
		// },
	}

	for _, tc := range testCases {
		client.Fire(&wire.Command{Cmd: "del k"})
		t.Run(tc.name, func(t *testing.T) {
			for i, cmd := range tc.cmds {
				if tc.delays[i] > 0 {
					time.Sleep(tc.delays[i])
				}
				result := client.Fire(cmd)
				// Since, we might have any of these 5 types (Response_VNil, Response_VInt, Response_VStr, Response_VFloat, Response_VBytes) in .Value, we do DeepEqual to check actual values
				assert.DeepEqual(t, tc.expect[i].Value, result.Value)
			}
		})
	}
}
