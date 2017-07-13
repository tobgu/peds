package peds_testing

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"testing"
)

func TestLenOfNewMap(t *testing.T) {
	m := NewStringIntMap()
	assertEqual(t, 0, m.Len())

	m2 := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1})
	assertEqual(t, 1, m2.Len())

	m3 := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1}, StringIntMapItem{Key: "b", Value: 2})
	assertEqual(t, 2, m3.Len())
}

func TestLoadAndStore(t *testing.T) {
	m := NewStringIntMap()

	m2 := m.Store("a", 1)
	assertEqual(t, 0, m.Len())
	assertEqual(t, 1, m2.Len())

	v, ok := m.Load("a")
	assertEqual(t, 0, v)
	assertEqualBool(t, false, ok)

	v, ok = m2.Load("a")
	assertEqual(t, 1, v)
	assertEqualBool(t, true, ok)
}

func TestLoadAndStoreIntKey(t *testing.T) {
	m := NewIntStringMap()

	m2 := m.Store(1, "")
	v, _ := m.Load(2)
	assertEqualString(t, "", v)

	v, _ = m2.Load(1)
	assertEqualString(t, "", v)
}

func TestLoadAndDeleteExistingItem(t *testing.T) {
	m := NewStringIntMap()
	m2 := m.Store("a", 1)
	m3 := m.Delete("a")

	assertEqual(t, 0, m3.Len())
	assertEqual(t, 1, m2.Len())

	v, ok := m2.Load("a")
	assertEqualBool(t, true, ok)
	assertEqual(t, 1, v)

	v, ok = m3.Load("a")
	assertEqualBool(t, false, ok)
	assertEqual(t, 0, v)
}

func TestLoadAndDeleteNonExistingItem(t *testing.T) {
	m := NewStringIntMap()
	m2 := m.Store("a", 1)
	m3 := m2.Delete("b")

	assertEqual(t, 1, m3.Len())
	assertEqual(t, 1, m2.Len())

	v, ok := m2.Load("a")
	assertEqualBool(t, true, ok)
	assertEqual(t, 1, v)

	if m2 != m3 {
		t.Errorf("m2 and m3 are not the same object: %p != %p", m2, m3)
	}
}

func TestRangeAllItems(t *testing.T) {
	m := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1}, StringIntMapItem{Key: "b", Value: 2}, StringIntMapItem{Key: "c", Value: 3})
	sum := 0
	m.Range(func(key string, value int) bool {
		sum += value
		return true
	})
	assertEqual(t, 6, sum)
}

func TestRangeStopOnKey(t *testing.T) {
	m := NewStringIntMap(StringIntMapItem{Key: "a", Value: 1}, StringIntMapItem{Key: "b", Value: 2}, StringIntMapItem{Key: "c", Value: 3})
	count := 0
	m.Range(func(key string, value int) bool {
		if key == "c" || key == "b" {
			return false
		}

		count++
		return true
	})

	if count > 1 {
		t.Errorf("Did not expect count to be more than 1")
	}
}

func TestLargeInsertAndLookup(t *testing.T) {
	m := NewStringIntMap()
	for j := 0; j < 100000; j++ {
		m = m.Store(fmt.Sprintf("%d", j), j)
	}

	for j := 0; j < 100000; j++ {
		v, ok := m.Load(fmt.Sprintf("%d", j))
		assertEqualBool(t, ok, true)
		assertEqual(t, v, j)
	}
}

//////////////////
/// Benchmarks ///
//////////////////

func BenchmarkInsertMap(b *testing.B) {
	// 5 - 6 times slower than native map
	// ~50% in store, ofBenchmarkInsertMap which ~14% in hash and ~20% in vector.Set()
	// ~50 in runtime._ExternalCode (memory allocation?)
	length := 0
	for i := 0; i < b.N; i++ {
		m := NewIntStringMap()
		for j := 0; j < 1000; j++ {
			m = m.Store(j, "a")
		}

		length += m.Len()
	}

	fmt.Println("Total length", length)
}

func BenchmarkInsertNativeMap(b *testing.B) {
	length := 0
	for i := 0; i < b.N; i++ {
		m := map[int]string{}
		for j := 0; j < 1000; j++ {
			m[j] = "a"
		}

		length += len(m)
	}

	fmt.Println("Total length", length)
}

/*
Results with generic/interface hash function:
BenchmarkAccessMap-2         	    5000	    296257 ns/op
BenchmarkAccessNativeMap-2   	   50000	     29424 ns/op

Results with specialized crc32 hash function (~3x overall improvement):
BenchmarkAccessMap-2         	   20000	     95464 ns/op
BenchmarkAccessNativeMap-2   	   50000	     30085 ns/op
*/

func BenchmarkAccessMap(b *testing.B) {
	// 11 - 12 times slower than native map
	// ~85% of time spent in generic pos function
	m := NewIntStringMap()
	for j := 0; j < 1000; j++ {
		m = m.Store(j, "a")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			_, _ = m.Load(j)
		}
	}
}

func BenchmarkAccessNativeMap(b *testing.B) {
	m := map[int]string{}
	for j := 0; j < 1000; j++ {
		m[j] = "a"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 1000; j++ {
			_, _ = m[j]
		}
	}
}

func BenchmarkInterfaceHash(b *testing.B) {
	b.ReportAllocs()
	result := uint32(0)
	for i := 0; i < b.N; i++ {
		result += interfaceHash(i)
	}

	fmt.Println(result)
}

func intHashFunc(x int) uint32 {
	// Adler32 is the quickest by far of the hash functions provided in the stdlib but its distribution is bad.
	// CRC32 has a fairly good distribution and is fairly quick.
	bX := make([]byte, 8)
	binary.LittleEndian.PutUint64(bX, uint64(x))
	return crc32.ChecksumIEEE(bX)
}

func BenchmarkIntHash(b *testing.B) {
	b.ReportAllocs()
	result := uint32(0)
	for i := 0; i < b.N; i++ {
		result += intHashFunc(i)
	}

	fmt.Println(result)
}

func BenchmarkLargeInsertAndLookup(b *testing.B) {
	/*
		Experiment with fixed size vector, the validity of this experiment is questionable...
		320  	      1	1592123509 ns/op	542679008 B/op	 1071516 allocs/op
		3200          2	 564513986 ns/op	240819736 B/op	 1296435 allocs/op
	    6400          2	 533717864 ns/op	221782588 B/op	 1299555 allocs/op
	    16000         2	 591304710 ns/op	211273412 B/op	 1303406 allocs/op
		32000:        2	 652429227 ns/op	208832108 B/op	 1307847 allocs/op
		320000:       2	 710530090 ns/op	281911676 B/op	 1597980 allocs/op

		Dynamic
		1 - 2:        2	 821062415 ns/op	254996996 B/op	 1643698 allocs/op
		2 - 4:        2	 673806304 ns/op	217694020 B/op	 1425024 allocs/op
		2 - 8:        2	 582710461 ns/op	215104792 B/op	 1352542 allocs/op
		4 - 8:        2	 630426666 ns/op	220647548 B/op	 1386341 allocs/op
		2 - 16:       2	 609139537 ns/op	229649368 B/op	 1379827 allocs/op
		20 - 40:      1	1234014496 ns/op	260392080 B/op	 1250811 allocs/op
	*/

	b.ReportAllocs()
	total := 0
	for i := 0; i < b.N; i++ {

		m := NewStringIntMap()
		total = 0
		for j := 0; j < 100000; j++ {
			m = m.Store(fmt.Sprintf("%d", j), j)
		}

		for j := 0; j < 100000; j++ {
			v, _ := m.Load(fmt.Sprintf("%d", j))
			total += v
		}

	}
	fmt.Println(total)
}

func BenchmarkLargeCreateInsertAndLookup(b *testing.B) {
	/*
	 10	 134831379 ns/op	14035978 B/op	  505355 allocs/op
	*/
	b.ReportAllocs()
	total := 0
	for i := 0; i < b.N; i++ {
		input := make([]StringIntMapItem, 0, 100000)
		for j := 0; j < 100000; j++ {
			input = append(input, StringIntMapItem{Key: fmt.Sprintf("%d", j), Value: j})
		}

		m := NewStringIntMap(input...)
		total = 0

		for j := 0; j < 100000; j++ {
			v, _ := m.Load(fmt.Sprintf("%d", j))
			total += v
		}

	}
	fmt.Println(total)
}

/*
$ go test -bench Hash -run=^$
BenchmarkGenericHash-2   	12173717644275345446
 5000000	       302 ns/op	      32 B/op	       4 allocs/op
BenchmarkIntHash-2       	10376819065122364326
20000000	        63.8 ns/op	      16 B/op	       2 allocs/op

Reusing byte buffer between hashes:
30000000	        39.8 ns/op	       8 B/op	       1 allocs/op
PASS

Using crc32.ChecksumIEEE
50000000	        34.3 ns/op	       0 B/op	       0 allocs/op
PASS

Using adler32.Checksum, very quick indeed but seems to have bad collision characteristics,
go with crc32 for now.
100000000	        18.8 ns/op	       0 B/op	       0 allocs/op
PASS

We may want to revisit this later. Using SipHash as done in Python,
Rust, etc may be a good longer term solution for a cryptographically saner/safer solution.
See https://github.com/dchest/siphash.
*/

/*
Profiling commands:

# Run specific benchmark
go test -bench=BenchmarkInsertMap -benchmem -run=^$ -memprofile=insert.mprof -cpuprofile=insert.prof --memprofilerate 1

# CPU
go tool pprof tests.test insert.prof

# Memory
go tool pprof --alloc_objects tests.test insert.mprof

*/

/* TODO:- Constructor from native map?
        - Improve parsing of specs to allow white spaces etc.
        - Dynamic sizing of backing vector depending on size of the map (which thresholds?)
        - More tests, store and load from larger structures
        - ToNativeMap() function (and ToNativeSlice for vectors?)
        - Custom imports?
        - Non comparable types cannot be used as keys (should be detected during compilation)
   	    - Test custom struct as key
   	    - Nicer interface for the vector iterator, see Scanner for an example
*/
