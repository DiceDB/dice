// This file is part of DiceDB.
// Copyright (C) 2025  DiceDB (dicedb.io).
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

package querymanager

// CreatePushResponse creates a push response. Push responses refer to messages that the server sends to clients without
// the client explicitly requesting them. These are typically seen in scenarios where the client has subscribed to some
// kind of event or data feed and is notified in real-time when changes occur.
// `key` is the unique key that identifies the push response.
func GenericWatchResponse(cmd, key string, result interface{}) (response []interface{}) {
	response = make([]interface{}, 3)
	response[0] = cmd
	response[1] = key
	response[2] = result
	return
}
