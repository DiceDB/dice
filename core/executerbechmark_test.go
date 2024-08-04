package core_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/xwb1989/sqlparser"
)

var itr int

var keys = []struct {
	count int
}{
	{count: 100},
	{count: 1000},
	{count: 10000},
	{count: 100000},
}

func populateData(count int) {

	if itr == 0 {
		config.KeysLimit = 100000000 // override keys limit high for benchmarking

		// need to init again for each round with the overriden buffer size
		// otherwise the watchchannel buffer size will stay as it is with the global keylimits size
		core.WatchChannel = make(chan core.WatchEvent, config.KeysLimit)

		dataset := []keyValue{}

		for i := 0; i < count; i++ {
			dataset = append(dataset, keyValue{fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)})
		}

		// Delete all keys
		for _, data := range dataset {
			obj := core.Get(data.key)
			if obj != nil {
				core.Del(data.key)
			}
		}

		// Insert all keys
		for _, data := range dataset {
			core.Put(data.key, &core.Obj{Value: data.value})
		}
		itr++ // Set count to 1 to indicate data has been populated (Prevent Benchamarker to again populate the data)
	}
}

func BenchmarkExecuteQueryOrderBykey(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   "asc",
			},
		}
		// Reset the timer to exclude the setup time from the benchmark
		b.ResetTimer()

		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryBasicOrderByValue(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$value",
				Order:   "asc",
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryLimit(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   false,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   "asc",
			},
			Limit: v.count / 3,
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryNoMatch(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "x*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhere(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("v3")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithComplexWhere(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$value",
				Order:   "desc",
			},
			Where: &sqlparser.AndExpr{
				Left: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: ">",
					Right:    sqlparser.NewStrVal([]byte("v2")),
				},
				Right: &sqlparser.ComparisonExpr{
					Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
					Operator: "<",
					Right:    sqlparser.NewStrVal([]byte("v100")),
				},
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCompareWhereKeyandValue(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: "=",
				Right:    &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhereNoMatch(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("nonexistent")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithNullValues(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "nullKey",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    &sqlparser.NullVal{},
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCaseSesnsitivity(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_value")},
				Operator: "=",
				Right:    sqlparser.NewStrVal([]byte("V9")), // Uppercase V3
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithClauseOnKey(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   "asc",
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("k3")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithEmptyKeyRegex(b *testing.B) {
	for _, v := range keys {
		itr = 0
		populateData(v.count)

		query := core.DSQLQuery{
			KeyRegex: "",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
