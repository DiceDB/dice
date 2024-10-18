package sql_test

import (
	"fmt"
	"testing"

	"github.com/dicedb/dice/internal/object"
	"github.com/dicedb/dice/internal/sql"

	"github.com/bytedance/sonic"
	"github.com/dicedb/dice/config"
	dstore "github.com/dicedb/dice/internal/store"
)

var benchmarkDataSizes = []int{100, 1000, 10000, 100000, 1000000}
var benchmarkDataSizesStackQueue = []int{100, 1000, 10000}
var benchmarkDataSizesJSON = []int{100, 1000, 10000, 100000}

var jsonList = map[string]string{
	"smallJSON": `{"score":10,"id":%d,"field1":{"field2":{"field3":{"score":10.36}}}}`,
	"largeJSON": `{"score":10,"id":%d,"field1":{"field2":{"field3":{"score":10.36}}},"inventory":{"mountain_bikes":[{"id":"bike:1","model":"Phoebe","price":1920,"specs":{"material":"carbon","weight":13.1},"colors":["black","silver"]},{"id":"bike:2","model":"Quaoar","price":2072,"specs":{"material":"aluminium","weight":7.9},"colors":["black","white"]},{"id":"bike:3","model":"Weywot","price":3264,"specs":{"material":"alloy","weight":13.8}}],"commuter_bikes":[{"id":"bike:4","model":"Salacia","price":1475,"specs":{"material":"aluminium","weight":16.6},"colors":["black","silver"]},{"id":"bike:5","model":"Mimas","price":3941,"specs":{"material":"alloy","weight":11.6}}]}}`,
}

func generateBenchmarkData(count int, store *dstore.Store) {
	config.DiceConfig.Memory.KeysLimit = 2000000 // Set a high limit for benchmarking
	store.ResetStore()

	data := make(map[string]*object.Obj, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		data[key] = store.NewObj(value, -1, object.ObjTypeString, object.ObjEncodingRaw)
	}
	store.PutAll(data)
}

func BenchmarkExecuteQueryOrderBykey(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key like 'k*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		// Reset the timer to exclude the setup time from the benchmark
		b.ResetTimer()

		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryBasicOrderByValue(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key like 'k*' ORDER BY $value ASC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryLimit(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := fmt.Sprintf("SELECT $key, $value WHERE $key like 'k*' ORDER BY $key ASC LIMIT %d", v/3)
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryNoMatch(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key like 'x*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithBasicWhere(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $value = 'v3' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithComplexWhere(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $value > 'v2' AND $value < 'v100' AND $key like 'k*' ORDER BY $value DESC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithCompareWhereKeyandValue(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key = $value AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithBasicWhereNoMatch(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $value = 'nonexistent' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithCaseSesnsitivity(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $value = 'V9' AND $key like 'k*'"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithClauseOnKey(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key > 'k3' AND $key like 'k*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func BenchmarkExecuteQueryWithAllMatchingKeyRegex(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizes {
		generateBenchmarkData(v, store)

		queryStr := "SELECT $key, $value WHERE $key like '*' ORDER BY $key ASC"
		query, err := sql.ParseQuery(queryStr)
		if err != nil {
			b.Fatal(err)
		}

		b.ResetTimer()
		b.Run(fmt.Sprintf("keys_%d", v), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
					b.Fatal(err)
				}
			}
		})
		store.ResetStore()
	}
}

func generateBenchmarkJSONData(b *testing.B, count int, json string, store *dstore.Store) {
	config.DiceConfig.Memory.KeysLimit = 2000000 // Set a high limit for benchmarking
	store.ResetStore()

	data := make(map[string]*object.Obj, count)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf(json, i)

		var jsonValue interface{}
		if err := sonic.UnmarshalString(value, &jsonValue); err != nil {
			b.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		data[key] = store.NewObj(jsonValue, -1, object.ObjTypeJSON, object.ObjEncodingJSON)
	}
	store.PutAll(data)
}

func BenchmarkExecuteQueryWithJSON(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)

			queryStr := "SELECT $key, $value WHERE $key like 'k*' AND '$value.id' = 3 ORDER BY $key ASC"
			query, err := sql.ParseQuery(queryStr)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
			store.ResetStore()
		}
	}
}

func BenchmarkExecuteQueryWithNestedJSON(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)

			queryStr := "SELECT $key, $value WHERE $key like 'k*' AND '$value.field1.field2.field3.score' > 10.1 ORDER BY $key ASC"
			query, err := sql.ParseQuery(queryStr)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
			store.ResetStore()
		}
	}
}

func BenchmarkExecuteQueryWithJsonInLeftAndRightExpressions(b *testing.B) {
	store := dstore.NewStore(nil, nil)
	for _, v := range benchmarkDataSizesJSON {
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)

			queryStr := "SELECT $key, $value WHERE '$value.id' = '$value.score' AND $key like 'k*' ORDER BY $key ASC"
			query, err := sql.ParseQuery(queryStr)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
			store.ResetStore()
		}
	}
}

func BenchmarkExecuteQueryWithJsonNoMatch(b *testing.B) {
	for _, v := range benchmarkDataSizesJSON {
		store := dstore.NewStore(nil, nil)
		for jsonSize, json := range jsonList {
			generateBenchmarkJSONData(b, v, json, store)

			queryStr := "SELECT $key, $value WHERE '$value.id' = 3 AND $key like 'k*' ORDER BY $key ASC"
			query, err := sql.ParseQuery(queryStr)
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			b.Run(fmt.Sprintf("%s_keys_%d", jsonSize, v), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					if _, err := sql.ExecuteQuery(&query, store.GetStore()); err != nil {
						b.Fatal(err)
					}
				}
			})
			store.ResetStore()
		}
	}
}
