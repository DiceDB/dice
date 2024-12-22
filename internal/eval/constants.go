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

package eval

const (
	BYTE = "BYTE"
	BIT  = "BIT"

	AND string = "AND"
	OR  string = "OR"
	XOR string = "XOR"
	NOT string = "NOT"

	Ex              string = "EX"
	Px              string = "PX"
	Persist         string = "PERSIST"
	Pxat            string = "PXAT"
	Exat            string = "EXAT"
	XX              string = "XX"
	NX              string = "NX"
	GT              string = "GT"
	LT              string = "LT"
	CH              string = "CH"
	INCR            string = "INCR"
	KeepTTL         string = "KEEPTTL"
	Sync            string = "SYNC"
	Async           string = "ASYNC"
	Help            string = "HELP"
	Memory          string = "MEMORY"
	Count           string = "COUNT"
	GetKeys         string = "GETKEYS"
	GetKeysandFlags string = "GETKEYSANDFLAGS"
	List            string = "LIST"
	Info            string = "INFO"
	Docs            string = "DOCS"
	null            string = "null"
	WithValues      string = "WITHVALUES"
	WithScores      string = "WITHSCORES"
	WithScore       string = "WITHSCORE"
	REV             string = "REV"
	GET             string = "GET"
	SET             string = "SET"
	INCRBY          string = "INCRBY"
	OVERFLOW        string = "OVERFLOW"
	WRAP            string = "WRAP"
	SAT             string = "SAT"
	FAIL            string = "FAIL"
	SIGNED          string = "SIGNED"
	UNSIGNED        string = "UNSIGNED"
	CAPACITY        string = "CAPACITY"
	SIZE            string = "SIZE"
	FILTERS         string = "FILTER"
	ITEMS           string = "ITEMS"
	EXPANSION       string = "EXPANSION"
)
