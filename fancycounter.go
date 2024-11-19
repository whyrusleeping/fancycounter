package fancycounter

import (
	"github.com/RoaringBitmap/roaring/roaring64"
)

type FancyCounter struct {
	limit  int
	thresh int
	maps   []*roaring64.Bitmap
}

func NewFancyCounter(limitPowTwo int) *FancyCounter {
	maps := make([]*roaring64.Bitmap, limitPowTwo)
	for i := 0; i < len(maps); i++ {
		maps[i] = roaring64.New()
	}

	return &FancyCounter{
		limit:  limitPowTwo,
		thresh: 1 << (limitPowTwo - 1),
		maps:   maps,
	}
}

func (fc *FancyCounter) Limit() int {
	return fc.limit
}

func (fc *FancyCounter) Thresh() int {
	return fc.thresh
}

func (fc *FancyCounter) Add(v uint64) {
	if fc.maps[fc.limit-1].Contains(v) {
		return
	}

	for i, m := range fc.maps {
		if m.Contains(v) {
			if i < fc.limit-1 {
				m.Remove(v)
			}
		} else {
			m.Add(v)
			break
		}
	}
}

func (fc *FancyCounter) AddN(v uint64, n int) {
	if fc.maps[fc.limit-1].Contains(v) {
		return
	}

	// adding a value equal to or greater than our threshold, just set the top bit and move on
	if n >= fc.thresh {
		fc.maps[fc.limit-1].Add(v)
		return
	}

	if n == 1 {
		fc.Add(v)
		return
	}

	for i := fc.limit - 2; i >= 0; i-- {
		bit := 1 << i
		if n&bit != 0 {
			for j := i; j < len(fc.maps); j++ {
				m := fc.maps[j]
				if m.Contains(v) {
					if j < fc.limit-1 {
						m.Remove(v)
					}
				} else {
					m.Add(v)
					if j == fc.limit-1 {
						// early exit if we set the top bit
						return
					}
					break
				}
			}
			n = n ^ bit
		}
	}
}

func (fc *FancyCounter) AddMany(inp *roaring64.Bitmap) {
	m := inp.Clone()
	fc.addManyPow2(m, 0)
}

func (fc *FancyCounter) addManyPow2(m *roaring64.Bitmap, powtwo int) {
	for i := powtwo; i < len(fc.maps) && !m.IsEmpty(); i++ {
		fc.maps[i].Xor(m)
		m.AndNot(fc.maps[i])
	}
}

func (fc *FancyCounter) Remove(v uint64) {
	for _, m := range fc.maps {
		m.Remove(v)
	}
}

func (fc *FancyCounter) RemoveLessThanThresh(thresh uint64) {
	for _, m := range fc.maps {
		m.RemoveRange(0, thresh)
	}
}

func (fc *FancyCounter) RemoveMany(rm *roaring64.Bitmap) {
	for _, m := range fc.maps {
		m.AndNot(rm)
	}
}

func (fc *FancyCounter) GetTopBits() *roaring64.Bitmap {
	return fc.maps[fc.limit-1]
}

func (fc *FancyCounter) GetNthTopSet(n int) *roaring64.Bitmap {
	return fc.maps[fc.limit-n]
}

func (fc *FancyCounter) MulAllByPow2(n int) {
	for i := 0; i < n; i++ {
		a := len(fc.maps) - 1
		b := len(fc.maps) - (i + 2)
		fc.maps[a].Or(fc.maps[b])
	}

	for i := 0; i < (len(fc.maps) - (n + 1)); i++ {
		a := len(fc.maps) - (2 + i)
		b := len(fc.maps) - (2 + i + n)
		fc.maps[a] = fc.maps[b]
	}

	for i := 0; i < n; i++ {
		fc.maps[i] = roaring64.New()
	}
}

func (fc *FancyCounter) AddFromCounter(ofc *FancyCounter) {
	for i := len(ofc.maps) - 1; i >= 0; i-- {
		mc := ofc.maps[i].Clone()
		fc.addManyPow2(mc, i)
	}
}

// AddFromCounterDestructive is a faster version of AddFromCounter that uses (and mutates) the input bitmaps
func (fc *FancyCounter) AddFromCounterDestructive(ofc *FancyCounter) {
	for i := len(ofc.maps) - 1; i >= 0; i-- {
		fc.addManyPow2(ofc.maps[i], i)
	}
}

func (fc *FancyCounter) DebugGetVals() map[uint64]int {
	out := make(map[uint64]int)
	for i := 0; i < len(fc.maps); i++ {
		iter := fc.maps[i].Iterator()
		for iter.HasNext() {
			v := iter.Next()
			out[v] += 1 << i
		}
	}
	return out
}

func (fc *FancyCounter) Clear() {
	for _, m := range fc.maps {
		m.Clear()
	}
}
