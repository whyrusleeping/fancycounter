package fancycounter

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/RoaringBitmap/roaring/roaring64"
)

type mapCounter struct {
	m      map[uint64]int
	thresh int
}

func (mc *mapCounter) Add(v uint64) {
	mc.m[v]++
}

func (mc *mapCounter) AddN(v uint64, n int) {
	mc.m[v] += n
}

func (mc *mapCounter) GetTopBits() *roaring64.Bitmap {
	out := roaring64.New()
	for k, v := range mc.m {
		if v >= mc.thresh {
			out.Add(k)
		}
	}
	return out
}

type counterIf interface {
	Add(uint64)
	AddN(uint64, int)
	GetTopBits() *roaring64.Bitmap
}

func benchmarkBitCounter(b *testing.B, newcounter func() counterIf) {
	for i := 0; i < b.N; i++ {
		c := newcounter()

		for n := 0; n < 1000000; n++ {
			v := rand.Intn(1_500_000)
			if v > 1_000_000 {
				v = v % 10_000
			}

			c.Add(uint64(v))
		}

		c.GetTopBits().GetCardinality()
	}
}

func BenchmarkMapBitCounter(b *testing.B) {
	mk := func() counterIf {
		return &mapCounter{
			m:      make(map[uint64]int),
			thresh: 4,
		}
	}

	b.ReportAllocs()
	benchmarkBitCounter(b, mk)
}

func BenchmarkFancyBitCounter(b *testing.B) {
	mk := func() counterIf {
		return NewFancyCounter(3)
	}

	b.ReportAllocs()
	benchmarkBitCounter(b, mk)
}

func benchmarkBitCounterAddN(b *testing.B, newcounter func() counterIf) {
	for i := 0; i < b.N; i++ {
		c := newcounter()

		for n := 0; n < 1000000; n++ {
			v := rand.Intn(1_500_000)
			if v > 1_000_000 {
				v = v % 10_000
			}
			n := v % 11

			c.AddN(uint64(v), n)
		}

		c.GetTopBits().GetCardinality()
	}
}

func BenchmarkMapBitCounterAddN(b *testing.B) {
	mk := func() counterIf {
		return &mapCounter{
			m:      make(map[uint64]int),
			thresh: 4,
		}
	}

	b.ReportAllocs()
	benchmarkBitCounterAddN(b, mk)
}

func BenchmarkFancyBitCounterAddN(b *testing.B) {
	mk := func() counterIf {
		return NewFancyCounter(3)
	}

	b.ReportAllocs()
	benchmarkBitCounterAddN(b, mk)
}

func TestFancyCounterMul(t *testing.T) {
	c := NewFancyCounter(6)

	c.AddN(1, 4)
	c.AddN(2, 6)
	c.AddN(3, 15)
	c.AddN(4, 17)
	c.AddN(5, 25)
	c.AddN(6, 31)
	c.AddN(7, 33)

	for i := 0; i < len(c.maps); i++ {
		fmt.Printf("%d: %d\n", i, c.maps[i].GetCardinality())
	}

	fmt.Println(c.DebugGetVals())
	c.MulAllByPow2(2)

	for i := 0; i < len(c.maps); i++ {
		fmt.Printf("%d: %d\n", i, c.maps[i].GetCardinality())
	}
	fmt.Println(c.DebugGetVals())

}

func TestAddMany(t *testing.T) {
	vals := make(map[uint64]int)
	var maps []*roaring64.Bitmap
	for mi := 0; mi < 10; mi++ {
		bm := roaring64.New()
		for i := 0; i < 400; i++ {
			v := uint64(rand.Intn(500))
			if bm.CheckedAdd(v) {
				vals[v]++
			}
		}
		maps = append(maps, bm)
	}
	fmt.Println(vals)

	fc := NewFancyCounter(5)

	for _, m := range maps {
		fc.AddMany(m)
	}

	ovals := fc.DebugGetVals()
	fmt.Println(ovals)
	if len(ovals) != len(vals) {
		t.Fatal("value count mismatch", len(ovals), len(vals))
	}
	for k, v := range ovals {
		if vals[k] != v {
			t.Fatal("value mismatch: ", k, v, vals[k])
		}
	}
}

func TestAddCounters(t *testing.T) {
	n := 30
	vals1 := make(map[uint64]int)
	fc1 := NewFancyCounter(5)
	for i := 0; i < n; i++ {
		k := uint64(rand.Intn(n))
		v := rand.Intn(5) + 1

		fc1.AddN(k, v)
		vals1[k] += v
	}

	vals2 := make(map[uint64]int)
	fc2 := NewFancyCounter(5)
	for i := 0; i < n; i++ {
		k := uint64(rand.Intn(n))
		v := rand.Intn(5) + 1

		fc2.AddN(k, v)
		vals2[k] += v
	}

	fc1dbg := fc1.DebugGetVals()
	fmt.Println("fc1: ", len(fc1dbg))

	fc2dbg := fc2.DebugGetVals()
	fmt.Println("fc2: ", len(fc2dbg))

	fmt.Println(vals1)
	fmt.Println(fc1dbg)
	fmt.Println(vals2)
	fmt.Println(fc2dbg)

	fc1.AddFromCounter(fc2)

	exp := make(map[uint64]int)
	for k, v := range vals1 {
		exp[k] = v
	}
	for k, v := range vals2 {
		exp[k] += v
	}

	outvals := fc1.DebugGetVals()
	fmt.Println("outvals: ", outvals)

	if len(outvals) != len(exp) {
		t.Fatal("set size mismatch", len(outvals), len(exp))
	}

	for k, v := range outvals {
		if exp[k] != v {
			t.Fatal("value mismatch: ", k, v, exp[k])
		}
	}
}
