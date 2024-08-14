package core_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/core"
	"github.com/dicedb/dice/internal/constants"
	"github.com/xwb1989/sqlparser"
)

var benchmarkDataSizes = []int{100, 1000, 10000, 100000, 1000000}

func generateBenchmarkData(count int) {
	config.KeysLimit = 2000000 // Set a high limit for benchmarking
	core.ResetStore()

	data := make(map[string]*core.Obj, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		data[key] = core.NewObj(value, -1, core.ObjTypeString, core.ObjEncodingRaw)
	}
	core.PutAll(data)
}

func BenchmarkExecuteQueryOrderBykey(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   constants.Asc,
			},
		}
		// Reset the timer to exclude the setup time from the benchmark
		b.ResetTimer()

		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryBasicOrderByValue(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$value",
				Order:   constants.Asc,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryLimit(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   false,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   constants.Asc,
			},
			Limit: v / 3,
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryNoMatch(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: "x*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhere(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithComplexWhere(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCompareWhereKeyandValue(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhereNoMatch(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithNullValues(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCaseSesnsitivity(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

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
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithClauseOnKey(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: "k*",
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: core.QueryOrder{
				OrderBy: "$key",
				Order:   constants.Asc,
			},
			Where: &sqlparser.ComparisonExpr{
				Left:     &sqlparser.ColName{Name: sqlparser.NewColIdent("_key")},
				Operator: ">",
				Right:    sqlparser.NewStrVal([]byte("k3")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithEmptyKeyRegex(b *testing.B) {
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v)
		defer core.ResetStore()

		query := core.DSQLQuery{
			KeyRegex: constants.EmptyStr,
			Selection: core.QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := core.ExecuteQuery(query); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
