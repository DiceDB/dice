package querywatcher

import (
	"fmt"
	"testing"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/config"
	"github.com/dicedb/dice/internal/constants"
	dstore "github.com/dicedb/dice/internal/store"
	"github.com/xwb1989/sqlparser"
)

var benchmarkDataSizes = []int{100, 1000, 10000, 100000, 1000000}
var benchmarkDataSizesStackQueue = []int{100, 1000, 10000}
var benchmarkDataSizesJSON = []int{100, 1000, 10000, 100000}

var jsonList = map[string]string{
	"smallJSON": `{"score":10,"id":%d,"field1":{"field2":{"field3":{"score":10.36}}}}`,
	"largeJSON": `{"score":10,"id":%d,"field1":{"field2":{"field3":{"score":10.36}}},"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`,
}

func generateBenchmarkData(count int, store *dstore.Store) {
	config.KeysLimit = 2000000 // Set a high limit for benchmarking
	store.ResetStore()

	data := make(map[string]*dstore.Obj, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		data[key] = store.NewObj(value, -1, dstore.ObjTypeString, dstore.ObjEncodingRaw)
	}
	store.PutAll(data)
}

func BenchmarkExecuteQueryOrderBykey(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "$key",
				Order:   constants.Asc,
			},
		}
		// Reset the timer to exclude the setup time from the benchmark
		b.ResetTimer()

		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryBasicOrderByValue(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "$value",
				Order:   constants.Asc,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryLimit(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   false,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
				OrderBy: "$key",
				Order:   constants.Asc,
			},
			Limit: v / 3,
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryNoMatch(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "x*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhere(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithComplexWhere(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCompareWhereKeyandValue(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithBasicWhereNoMatch(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithNullValues(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "nullKey",
			Selection: QuerySelection{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithCaseSesnsitivity(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithClauseOnKey(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: "k*",
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
			OrderBy: QueryOrder{
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
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkExecuteQueryWithEmptyKeyRegex(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)
		defer store.ResetStore()

		query := DSQLQuery{
			KeyRegex: constants.EmptyStr,
			Selection: QuerySelection{
				KeySelection:   true,
				ValueSelection: true,
			},
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func generateBenchmarkJSONData(b *testing.B, count int, json string, store *dstore.Store) {
	config.KeysLimit = 2000000 // Set a high limit for benchmarking
	store.ResetStore()

	data := make(map[string]*dstore.Obj, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf(json, i)

		var jsonValue interface{}
		if err := sonic.UnmarshalString(value, &jsonValue); err != nil {
			b.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		data[key] = store.NewObj(jsonValue, -1, dstore.ObjTypeJSON, dstore.ObjEncodingJSON)
	}
	store.PutAll(data)
}

func BenchmarkExecuteQueryWithJSON(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)
			defer store.ResetStore()

			query := DSQLQuery{
				KeyRegex: "k*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     sqlparser.NewStrVal([]byte("_value.id")),
					Operator: "=",
					Right:    sqlparser.NewIntVal([]byte("3")),
				},
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkExecuteQueryWithNestedJSON(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)
			defer store.ResetStore()

			query := DSQLQuery{
				KeyRegex: "k*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     sqlparser.NewStrVal([]byte("_value.field1.field2.field3.score")),
					Operator: ">",
					Right:    sqlparser.NewFloatVal([]byte("10.1")),
				},
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkExecuteQueryWithJsonInLeftAndRightExpressions(b *testing.B) {
	store := dstore.NewStore(nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)
			defer store.ResetStore()

			query := DSQLQuery{
				KeyRegex: "k*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     sqlparser.NewStrVal([]byte("_value.id")),
					Operator: "=",
					Right:    sqlparser.NewStrVal([]byte("_value.score")),
				},
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkExecuteQueryWithJsonNoMatch(b *testing.B) {
	for _, v := range benchmarkDataSizesJSON {
		store := dstore.NewStore(nil)
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)
			defer store.ResetStore()

			query := DSQLQuery{
				KeyRegex: "k*",
				Selection: QuerySelection{
					KeySelection:   true,
					ValueSelection: true,
				},
				Where: &sqlparser.ComparisonExpr{
					Left:     sqlparser.NewStrVal([]byte("_value.id")),
					Operator: "=",
					Right:    sqlparser.NewIntVal([]byte("-1")),
				},
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}
