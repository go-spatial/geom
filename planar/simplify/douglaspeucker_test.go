package simplify

import (
	"context"
	"flag"
	"math"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/encoding/wkt"
	gtesting "github.com/go-spatial/geom/testing"
)

var ignoreSanityCheck bool

func init() {
	flag.BoolVar(&ignoreSanityCheck, "ignoreSanityCheck", false, "ignore sanity checks in test cases.")
}

// string2line
func s2l(s string) [][2]float64 {
	g, err := wkt.NewDecoder(strings.NewReader(s)).Decode()
	if err != nil {
		panic(err)
	}

	return ([][2]float64)(g.(geom.LineString))
}

func TestDouglasPeucker(t *testing.T) {
	flag.Parse()

	type tcase struct {
		l  [][2]float64
		dp DouglasPeucker
		el [][2]float64
	}

	fn := func(t *testing.T, tc tcase) {
		ctx := context.Background()
		gl, err := tc.dp.Simplify(ctx, tc.l, false)
		// Douglas Peucker should never return an error.
		// This is more of a sanity check.
		if err != nil {
			t.Errorf("Douglas Peucker error, expected nil got %v", err)
			return
		}

		if !cmp.LineStringEqual(tc.el, gl) {
			t.Errorf("simplified points, expected\n%v\n\tgot\n%v", tc.el, gl)
			return
		}

		if ignoreSanityCheck {
			return
		}

		// Let's try it with true, it should not matter, as DP does not care.
		// More sanity checking.
		gl, _ = tc.dp.Simplify(ctx, tc.l, true)

		if !cmp.LineStringEqual(tc.el, gl) {
			t.Errorf("simplified points (true), expected %v got %v", tc.el, gl)
			return
		}
	}

	tests := map[string]tcase{
		"simple box": {
			l: [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
			dp: DouglasPeucker{
				Tolerance: 0.001,
			},
			el: [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
		},
		"x axis": {
			l: gtesting.FuncLineString(0, 100, 100, func(t float64) [2]float64 {
				return [2]float64{t, 0} // all points on x-axis
			}),
			dp: DouglasPeucker{
				Tolerance: 0.001,
			},
			el: [][2]float64{{0, 0}, {100, 0}},
		},
		"line": {
			l: gtesting.FuncLineString(0, 100, 100, func(t float64) [2]float64 {
				return [2]float64{t, t} // all points on x-axis
			}),
			dp: DouglasPeucker{
				Tolerance: 0.001,
			},
			el: [][2]float64{{0, 0}, {100, 100}},
		},
		"sin": { // should be left with a zigzag
			l: gtesting.SinLineString(1, 0, 2*math.Pi, 9),
			dp: DouglasPeucker{
				Tolerance: 0.5,
			},
			el: [][2]float64{{0, 0}, {math.Pi / 2, 1}, {3 * math.Pi / 2, -1}, {2 * math.Pi, 0}},
		},
		"natural earth line string 0": {
			l: gtesting.NaturalEarthLineStrings[0],
			dp: DouglasPeucker{
				Tolerance: 500,
			},
			el: s2l("LINESTRING(-7785560.894 5112305.653,-7784854.276 5122268.298,-7786050.091 5139676.21,-7790380.39 5154033.469,-7793922.539 5160820.971,-7798053.535 5166936.297,-7805482.082 5172042.522,-7813762.194 5173879.48,-7817897.72 5173061.654)"),
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func BenchmarkDouglasPeuckerCircle(b *testing.B) {
	rdp := DouglasPeucker{
		Tolerance: 0.001,
	}
	ctx := context.Background()

	circleFn := func(t float64) [2]float64 {
		return [2]float64{math.Cos(t), math.Sin(t)}
	}

	circle := gtesting.FuncLineString(0, 3, 10000, circleFn)

	for i := 0; i < b.N; i++ {
		rdp.Simplify(ctx, circle, false)
	}

	g, err := rdp.Simplify(ctx, circle, false)
	if err != nil {
		panic(err)
	}

	if !ignoreSanityCheck {
		b.Logf("simplified/initial points: %d/%d", len(g), len(circle))
	}
}

func BenchmarkDouglasPeuckerWavyCircle0(b *testing.B) {
	rdp := DouglasPeucker{
		Tolerance: 0.01,
	}
	ctx := context.Background()

	circleFn := func(t float64) [2]float64 {
		return [2]float64{
			math.Sin(10*t)*math.Cos(t) + 3*math.Cos(t),
			math.Sin(10*t)*math.Sin(t) + 3*math.Sin(t),
		}
	}

	circle := gtesting.FuncLineString(0, 3, 10000, circleFn)

	for i := 0; i < b.N; i++ {
		rdp.Simplify(ctx, circle, false)
	}

	g, err := rdp.Simplify(ctx, circle, false)
	if err != nil {
		panic(err)
	}

	if !ignoreSanityCheck {
		b.Logf("simplified/initial points: %d/%d", len(g), len(circle))
	}
}

func BenchmarkDouglasPeuckerWavyCircle1(b *testing.B) {
	rdp := DouglasPeucker{
		Tolerance: 0.01,
	}
	ctx := context.Background()

	circleFn := func(t float64) [2]float64 {
		return [2]float64{
			math.Sin(10*t)*math.Cos(t) + 3*t*math.Cos(t),
			math.Sin(10*t)*math.Sin(t) + 3*t*math.Sin(t),
		}
	}

	circle := gtesting.FuncLineString(0, 10, 100000, circleFn)

	for i := 0; i < b.N; i++ {
		rdp.Simplify(ctx, circle, false)
	}

	g, err := rdp.Simplify(ctx, circle, false)
	if err != nil {
		panic(err)
	}

	if !ignoreSanityCheck {
		b.Logf("simplified/initial points: %d/%d", len(g), len(circle))
	}
}

// BenchmarkDouglasPeuckerZigZag is the worst case scenario. No points
// are dropped, so the recursion eliminates 1 point each time resulting
// in an n^2 time complexity
// https://stackoverflow.com/questions/31516058/line-that-triggers-worst-case-for-douglas-peucker-algorithm
func BenchmarkDouglasPeuckerZigZag(b *testing.B) {
	rdp := DouglasPeucker{
		Tolerance: 0.01,
	}
	ctx := context.Background()

	zigZagFn := func(t float64) [2]float64 {
		var y float64
		if int(t)%2 == 0 {
			y = -1
		} else {
			y = 1
		}
		y *= math.Log(t)
		return [2]float64{t, y}
	}

	zigZag := gtesting.FuncLineString(1, 1000, 1000, zigZagFn)

	for i := 0; i < b.N; i++ {
		rdp.Simplify(ctx, zigZag, false)
	}

	g, err := rdp.Simplify(ctx, zigZag, false)
	if err != nil {
		panic(err)
	}

	if !ignoreSanityCheck {
		b.Logf("simplified/initial points: %d/%d", len(g), len(zigZag))
	}
}

func BenchmarkDouglasPeuckerSA(b *testing.B) {
	rdp := DouglasPeucker{
		Tolerance: 20000,
	}
	ctx := context.Background()

	sa := gtesting.SouthAfrica[0]

	for i := 0; i < b.N; i++ {
		rdp.Simplify(ctx, sa, false)
	}

	g, err := rdp.Simplify(ctx, sa, false)
	if err != nil {
		panic(err)
	}

	if !ignoreSanityCheck {
		b.Logf("simplified/initial points: %d/%d", len(g), len(sa))
	}
}
