package dashboard

import "math/big"

// All chart arithmetic is integer arithmetic. Values are millionths (int64),
// coordinates are hundredths of a pixel (int64). No float and no
// transcendental function is involved, so output is identical on every
// platform.

// Scale is a y axis: Lo and Hi are multiples of Step, in millionths.
type Scale struct {
	Lo, Hi, Step int64
}

// Intervals returns the number of steps between Lo and Hi.
func (s Scale) Intervals() int64 { return (s.Hi - s.Lo) / s.Step }

// Ticks returns the gridline values from Lo to Hi.
func (s Scale) Ticks() []int64 {
	var t []int64
	for v := s.Lo; v <= s.Hi; v += s.Step {
		t = append(t, v)
	}
	return t
}

func floorDiv(a, b int64) int64 {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func ceilDiv(a, b int64) int64 { return -floorDiv(-a, b) }

// NiceScale returns the axis for the range lo0..hi0 (lo0 < hi0). The step is
// the smallest of 1, 2 and 5 times a power of ten (in millionths) for which the
// axis, widened outward to multiples of the step, has at most 5 intervals. If
// that leaves fewer than 3 intervals, whole steps are added until there are 3.
// The result has 3 to 5 intervals, so 4 to 6 gridlines. A range narrower than
// 3 millionths is first widened to 3 so the lower bound holds there too.
func NiceScale(lo0, hi0 int64) Scale {
	if hi0-lo0 < 3 {
		hi0 = lo0 + 3
	}
	step := int64(1)
	for k := 0; k <= 18; k++ {
		for _, m := range []int64{1, 2, 5} {
			step = m
			for i := 0; i < k; i++ {
				step *= 10
			}
			lo := floorDiv(lo0, step) * step
			hi := ceilDiv(hi0, step) * step
			if (hi-lo)/step <= 5 {
				// Widening by whole steps keeps at least 3 intervals. It happens
				// only when the range straddles aligned values, such as -5..5
				// with step 5; the side with less slack gets the extra step.
				for (hi-lo)/step < 3 {
					if hi-hi0 <= lo0-lo {
						hi += step
					} else {
						lo -= step
					}
				}
				return Scale{Lo: lo, Hi: hi, Step: step}
			}
		}
	}
	// Unreachable for magnitudes below 1e12 (1e18 millionths).
	return Scale{Lo: lo0, Hi: hi0, Step: hi0 - lo0}
}

// BarRange is the data range for a bar chart: always includes zero.
func BarRange(dmin, dmax int64) (int64, int64) {
	lo, hi := dmin, dmax
	if lo > 0 {
		lo = 0
	}
	if hi < 0 {
		hi = 0
	}
	if lo == hi {
		hi = lo + 1_000_000
	}
	return lo, hi
}

// LineRange is the data range for a line chart. Rule: if every value is
// non-negative and the smallest is at most half the largest, the axis starts
// at 0; if every value is non-positive and the largest is at least half the
// smallest, it ends at 0; otherwise it spans the data and the nice scale
// rounds the ends outward to multiples of the step, so the lower bound is
// not always 0. A flat series spans 0 to its value.
func LineRange(dmin, dmax int64) (int64, int64) {
	lo, hi := dmin, dmax
	if dmin >= 0 && dmin*2 <= dmax {
		lo = 0
	}
	if dmax <= 0 && dmax*2 >= dmin {
		hi = 0
	}
	if lo == hi {
		return BarRange(lo, hi)
	}
	return lo, hi
}

// roundDiv returns num/den rounded half up for num >= 0, den > 0.
func roundDiv(num, den *big.Int) *big.Int {
	n := new(big.Int).Lsh(num, 1)
	n.Add(n, den)
	return n.Div(n, new(big.Int).Lsh(den, 1))
}

// yOf maps a value to a y coordinate in hundredths of a pixel. top and bottom
// are the plot edges; Hi maps to top and Lo to bottom.
func (s Scale) yOf(v, top, bottom int64) int64 {
	num := new(big.Int).Mul(big.NewInt(v-s.Lo), big.NewInt(bottom-top))
	d := roundDiv(num, big.NewInt(s.Hi-s.Lo))
	return bottom - d.Int64()
}

// hund writes hundredths with exactly two decimals.
func hund(v int64) string {
	sign := ""
	if v < 0 {
		sign, v = "-", -v
	}
	return sign + itoa(v/100) + "." + pad2(v%100)
}

func pad2(v int64) string {
	if v < 10 {
		return "0" + itoa(v)
	}
	return itoa(v)
}

func itoa(v int64) string {
	if v == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for v > 0 {
		i--
		b[i] = byte('0' + v%10)
		v /= 10
	}
	return string(b[i:])
}
