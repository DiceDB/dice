package core_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/core"
	"github.com/xwb1989/sqlparser"
)

var itr int

type tupple struct {
	key   string
	value string
}

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
		// fmt.Println("Starting to populate data...")
		dataset := []tupple{}

		for i := 0; i < count; i++ {
			dataset = append(dataset, tupple{fmt.Sprintf("k%d", i), fmt.Sprintf("v%d", i)})
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

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
			Limit: 3,
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
				Right:    sqlparser.NewStrVal([]byte("v9")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
			}
		})
	}
}

func BenchmarkExecuteQueryWithIncompatibleTypes(b *testing.B) {
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
				Right:    sqlparser.NewIntVal([]byte("42")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
			}
		})
	}
}

func BenchmarkExecuteQueryWithUnsupportedOp(b *testing.B) {
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
				Operator: "LIKE",
				Right:    sqlparser.NewStrVal([]byte("%3")),
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v.count), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				core.ExecuteQuery(query)
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
				core.ExecuteQuery(query)
			}
		})
	}
}
