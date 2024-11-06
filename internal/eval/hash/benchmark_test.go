package hash

import (
	"testing"
	"fmt"
	"strconv"
	diceerrors "github.com/dicedb/dice/internal/errors"
	"math"
	"strings"
)


// Benchmark for Set method
func benchmarkSimpleMapSet(b *testing.B, encoding int) {
	m := NewHashMap[int,int](encoding)
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

// Benchmark for Get method
func benchmarkSimpleMapGet(b *testing.B, encoding int) {
	m := NewHashMap[int,int](encoding)
	// Pre-fill the map with some data
	for i := 0; i < 1000; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := i % 1000 // Ensure keys are within the pre-filled range
		m.Get(key)
	}
}

// Benchmark for Delete method
func benchmarkSimpleMapDelete(b *testing.B, encoding int) {
	m := NewHashMap[int,int](encoding)
	// Pre-fill the map with some data
	for i := 0; i < 1000; i++ {
		m.Set(i, i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := i % 1000 // Ensure keys are within the pre-filled range
		m.Delete(key)
	}
}

// Benchmark for Set method
func benchmarkSimpleMapSetString(b *testing.B, encoding int) {
	m := NewHashMap[string,string](encoding)
	for i := 0; i < b.N; i++ {
		m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

// Benchmark for Get method
func benchmarkSimpleMapGetString(b *testing.B, encoding int) {
	m := NewHashMap[string,string](encoding)
	// Pre-fill the map with some data
	for i := 0; i < 1000; i++ {
		m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(fmt.Sprintf("key%d", i))
	}
}

// Benchmark for Delete method
func benchmarkSimpleMapDelString(b *testing.B, encoding int) {
	m := NewHashMap[string,string](encoding)
	// Pre-fill the map with some data
	for i := 0; i < 1000; i++ {
		m.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(fmt.Sprintf("key%d", i))
	}
}

func benchmarkSimpleMapALENString(b *testing.B, encoding int) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000}


	for _, size := range sizes {
		b.Run(fmt.Sprintf("HashSize_%d", size), func(b *testing.B) {
			m := NewHashMap[string,string](encoding)
			for i := 0; i < size; i++ {
				m.Set(fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i))
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.ALen()
			}
		})
	}
}

func benchmarkSimpleMapALENInt(b *testing.B, encoding int) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000}


	for _, size := range sizes {
		b.Run(fmt.Sprintf("HashSize_%d", size), func(b *testing.B) {
			m := NewHashMap[int,int](encoding)
			for i := 0; i < size; i++ {
				m.Set(i,i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.ALen()
			}
		})
	}
}

func benchmarkSimpleMapLENString(b *testing.B, encoding int) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000}


	for _, size := range sizes {
		b.Run(fmt.Sprintf("HashSize_%d", size), func(b *testing.B) {
			m := NewHashMap[string,string](encoding)
			for i := 0; i < size; i++ {
				m.Set(fmt.Sprintf("field%d", i), fmt.Sprintf("value%d", i))
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.Len()
			}
		})
	}
}

func benchmarkSimpleMapLENInt(b *testing.B) {
	sizes := []int{0, 10, 100, 1000, 10000, 100000}


	for _, size := range sizes {
		b.Run(fmt.Sprintf("HashSize_%d", size), func(b *testing.B) {
			m := NewHashMap[int,int](6)
			for i := 0; i < size; i++ {
				m.Set(i,i)
			}
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				m.Len()
			}
		})
	}
}

func IncrementInt(val int64, incr int64) (int64, error) {
	if (incr < 0 && val < 0 && incr < (math.MinInt64-val)) ||
		(incr > 0 && val > 0 && incr > (math.MaxInt64-val)) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}

func IncrementFloat(val float64, incr float64) (float64, error) {
	if math.IsInf(val+incr, 1) || math.IsInf(val+incr, -1) {
		return -1, diceerrors.NewErr(diceerrors.IncrDecrOverflowErr)
	}
	return val + incr, nil
}

func formatFloat(f float64, b bool) string {
	formatted := strconv.FormatFloat(f, 'f', -1, 64)
	if b {
		parts := strings.Split(formatted, ".")
		if len(parts) == 1 {
			formatted += ".0"
		}
	}
	return formatted
}

func benchmarkSimpleMapHINCRBYString(b *testing.B, encoding int) {
	m := NewHashMap[string,string](encoding)
	// creating new fields
	for i := 0; i < b.N; i++ {
		incr := int64(i)
		m.CreateOrModify(fmt.Sprintf("FIELD_%d", i), func(element string) (altered string, err error) {
			if element == "" {
				element = "0"
			}
			i, err := strconv.ParseInt(element, 10, 64)
			if err != nil {
				return "-1", diceerrors.NewErr(diceerrors.HashValueNotIntegerErr)
			}
			newVal, err := IncrementInt(i, incr)
			return fmt.Sprintf("%v", newVal), err
		})
	}

	// updating the existing fields
	for i := 0; i < b.N; i++ {
		incr := int64(i)
		m.CreateOrModify(fmt.Sprintf("FIELD_%d", i), func(element string) (altered string, err error) {
			if element == "" {
				element = "0"
			}
			i, err := strconv.ParseInt(element, 10, 64)
			if err != nil {
				return "-1", diceerrors.NewErr(diceerrors.HashValueNotIntegerErr)
			}
			newVal, err := IncrementInt(i, incr)
			return fmt.Sprintf("%v", newVal), err
		})
	}
}

func benchmarkSimpleMapHINCRBYInt(b *testing.B, encoding int) {
	m := NewHashMap[string,int64](encoding)
	// creating new fields
	for i := 0; i < b.N; i++ {
		incr := int64(i)
		m.CreateOrModify(fmt.Sprintf("FIELD_%d", i), func(element int64) (altered int64, err error) {
			newVal, err := IncrementInt(element, incr)
			return newVal, err
		})
	}

	// updating the existing fields
	for i := 0; i < b.N; i++ {
		incr := int64(i)
		m.CreateOrModify(fmt.Sprintf("FIELD_%d", i), func(element int64) (altered int64, err error) {
			newVal, err := IncrementInt(element, incr)
			return newVal, err
		})
	}
}

func benchmarkSimpleMapHINCRBYFLOATString(b *testing.B, encoding int) {
	m := NewHashMap[string,string](encoding)

	// Setting initial fields with some values
	m.Set("field1", "1.0")
	m.Set("field2", "1.2")

	inputs := []struct {
		field string
		incr  string
	}{
		{"field1", "0.1"},
		{"field1", "-0.1"},
		{"field2", "1000000.1"},
		{ "field2", "-1000000.1"},
		{ "field1", "-10.1234"},
		{ "field1", "1.5"},  // testing with non-existing key
		{ "field2", "2.75"}, // testing with non-existing field in existing key
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("HINCRBYFLOAT %s %s",input.field, input.incr), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				incr, _ := strconv.ParseFloat(input.incr, 64)
				m.CreateOrModify(input.field, func(element string) (altered string, err error) {
					if element == "" {
						element = "0"
					}
					i, err := strconv.ParseFloat(element, 64)
					if err != nil {
						return "-1", diceerrors.NewErr(diceerrors.IntOrFloatErr)
					}
					newVal, err := IncrementFloat(i, incr)
					strValue := formatFloat(newVal, false)
					return strValue, err
				})
			}
		})
	}
}


func benchmarkSimpleMapHINCRBYFLOATFloat(b *testing.B, encoding int) {
	m := NewHashMap[string,float64](encoding)

	// Setting initial fields with some values
	m.Set("field1", float64(1.0))
	m.Set("field2", float64(1.2))

	inputs := []struct {
		field string
		incr  float64
	}{
		{"field1", float64(0.1)},
		{"field1", float64(-0.1)},
		{"field2", float64(1000000.1)},
		{ "field2",float64(-1000000.1)},
		{ "field1", float64(-10.1234)},
		{ "field1", float64(1.5)},  // testing with non-existing key
		{ "field2", float64(2.75)}, // testing with non-existing field in existing key
	}

	for _, input := range inputs {
		b.Run(fmt.Sprintf("HINCRBYFLOAT %s %f",input.field, input.incr), func(b *testing.B) {
			for i := 0; i < b.N; i++ {

				m.CreateOrModify(input.field, func(element float64) (altered float64, err error) {
					newVal, err := IncrementFloat(element, input.incr)
					return newVal, err
				})
			}
		})
	}
}

func BenchmarkAllHashMaps(b *testing.B) {
	encodings := map[int]string{
		6: "simple_map",
		7:"ComplexMap",
	}
	for encoding, name := range encodings {
		b.Run(name, func(b *testing.B) {
			benchmarkSimpleMapSet(b, encoding)
			benchmarkSimpleMapGet(b, encoding)
			benchmarkSimpleMapDelete(b, encoding)
			benchmarkSimpleMapSetString(b, encoding)
			benchmarkSimpleMapGetString(b, encoding)
			benchmarkSimpleMapDelString(b, encoding)
			benchmarkSimpleMapALENString(b, encoding)
			benchmarkSimpleMapALENInt(b, encoding)
			benchmarkSimpleMapLENString(b, encoding)
			benchmarkSimpleMapLENInt(b)
			benchmarkSimpleMapHINCRBYString(b, encoding)
			benchmarkSimpleMapHINCRBYInt(b, encoding)
			benchmarkSimpleMapHINCRBYFLOATString(b, encoding)
			benchmarkSimpleMapHINCRBYFLOATFloat(b, encoding)
		})
	}
}