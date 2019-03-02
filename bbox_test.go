package geom_test

import (
	"math"
	"reflect"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestExtentNew(t *testing.T) {
	type tcase struct {
		points   [][2]float64
		expected *geom.Extent
	}
	var tests map[string]tcase
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		got := geom.NewExtent(tc.points...)
		if !reflect.DeepEqual(got, tc.expected) {
			t.Errorf("failed,  expected %+v got %+v", tc.expected, *got)
		}
	}
	tests = map[string]tcase{

		"a point": {
			points: [][2]float64{
				{1.0, 2.0},
			},
			expected: &geom.Extent{1.0, 2.0, 1.0, 2.0},
		},
		"3 points": {
			points: [][2]float64{
				{0.0, 0.0},
				{6.0, 4.0},
				{3.0, 7.0},
			},
			expected: &geom.Extent{0.0, 0.0, 6.0, 7.0},
		},
		"4 points": {
			points: [][2]float64{
				{0.0, 0.0},
				{-10.0, -10.0},
				{6.0, 4.0},
				{3.0, 7.0},
			},
			expected: &geom.Extent{-10.0, -10.0, 6.0, 7.0},
		},
		"0 points": {
			points:   [][2]float64{},
			expected: nil,
		},
		"2 points":{
			points: [][2]float64{
				{1286931.4293799447, 6138849.430462967},
				{1286970.842222649, 6138810.017620263},
			},
			expected: &geom.Extent{1286931.4293799447, 6138810.017620263, 1286970.842222649, 6138849.430462967},
		},
		"2 points swapped":{
			points: [][2]float64{
				{1.2869314293799447e+06, 6.138810017620263e+06},
				{1.286970842222649e+06,6.138849430462967e+06},
			},
			expected: &geom.Extent{1286931.4293799447, 6138810.017620263, 1286970.842222649, 6138849.430462967},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentAdd(t *testing.T) {
	type tcase struct {
		bb       *geom.Extent
		extent   *geom.Extent
		expected *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		bb := tc.bb
		bb.Add(tc.extent)
		if !cmp.GeomExtent(tc.expected, bb) {
			t.Errorf("failed, expected %+v got %+v", tc.expected, bb)
		}
	}
	tests := map[string]tcase{
		"nil expanded by point": {
			bb:       nil,
			extent:   &geom.Extent{3.0, 3.0, 3.0, 3.0},
			expected: nil,
		},
		"point expanded by nil": {
			bb:       &geom.Extent{1.0, 2.0, 1.0, 2.0},
			extent:   nil,
			expected: nil,
		},
		"point expanded by point": {
			bb:       &geom.Extent{1.0, 2.0, 1.0, 2.0},
			extent:   &geom.Extent{3.0, 3.0, 3.0, 3.0},
			expected: &geom.Extent{1.0, 2.0, 3.0, 3.0},
		},
		"point expanded by enclosing box": {
			bb:       &geom.Extent{1.0, 2.0, 1.0, 2.0},
			extent:   &geom.Extent{0.0, 0.0, 3.0, 3.0},
			expected: &geom.Extent{0.0, 0.0, 3.0, 3.0},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentAddPoints(t *testing.T) {
	type tcase struct {
		bb       *geom.Extent
		points   [][2]float64
		expected *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		bb := tc.bb
		bb.AddPoints(tc.points...)
		if !cmp.GeomExtent(tc.expected, bb) {
			t.Errorf("failed, expected %+v got %+v", tc.expected, bb)
		}
	}
	tests := map[string]tcase{
		"nil expanded by point": {
			bb: nil,
			points: [][2]float64{
				{1.0, 2.0},
			},
			expected: nil,
		},
		"point expanded zero points": {
			bb:       &geom.Extent{1.0, 2.0, 1.0, 2.0},
			points:   [][2]float64{},
			expected: &geom.Extent{1.0, 2.0, 1.0, 2.0},
		},
		"point expanded by point": {
			bb: &geom.Extent{1.0, 2.0, 1.0, 2.0},
			points: [][2]float64{
				{3.0, 3.0},
				{1.0, 1.0},
			},
			expected: &geom.Extent{1.0, 1.0, 3.0, 3.0},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentAddGeometry(t *testing.T) {
	type tcase struct {
		bb       *geom.Extent
		g        geom.Geometry
		expected *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		bb := tc.bb
		err := bb.AddGeometry(tc.g)
		if err != nil {
			t.Errorf("failed, expected nil got %+v", err)
		}
		if !cmp.GeomExtent(tc.expected, bb) {
			t.Errorf("failed, expected %+v got %+v", tc.expected, bb)
		}
	}
	tests := map[string]tcase{
		"point expanded by LineString": {
			bb:       &geom.Extent{1.0, 2.0, 1.0, 2.0},
			g:        &geom.LineString{{0.0, 3.0}, {4.0, 5.0}, {3.0, -1.0}},
			expected: &geom.Extent{0, -1, 4, 5},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentContains(t *testing.T) {
	type tcase struct {
		mm       geom.MinMaxer
		bb       *geom.Extent
		expected bool
	}
	fn := func(t *testing.T, tc tcase) {
		got := tc.bb.Contains(tc.mm)
		if got != tc.expected {
			t.Errorf(" contains, expected %v got %v", tc.expected, got)
		}
	}
	tests := map[string]tcase{
		"nil bb nil mm": {
			expected: true,
		},
		"nil bb non-nil mm": {
			mm:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			expected: true,
		},
		"non-nil bb nil mm": {
			bb:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			expected: false,
		},
		"same": {
			bb:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			mm:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			expected: true,
		},
		"contained": {
			bb:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			mm:       geom.NewExtent([2]float64{1, 1}, [2]float64{5, 5}),
			expected: true,
		},
		"same only at 0,0": {
			bb:       geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			mm:       geom.NewExtent([2]float64{0, 0}, [2]float64{-10, -10}),
			expected: false,
		},
		"overlap not contained": {
			bb:       geom.NewExtent([2]float64{-1, -1}, [2]float64{10, 10}),
			mm:       geom.NewExtent([2]float64{0, 0}, [2]float64{-10, -10}),
			expected: false,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentContainsPoint(t *testing.T) {
	type tcase struct {
		bb       *geom.Extent
		pt       [2]float64
		expected bool
	}
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		bb := tc.bb
		got := bb.ContainsPoint(tc.pt)
		exp := tc.expected
		does := "does "
		if !exp {
			does = "does not "
		}
		if got != exp {
			t.Errorf(does+" contain, expected %v got %v", exp, got)
		}
	}
	tests := map[string]tcase{
		"contained point": {
			bb:       &geom.Extent{0.0, 0.0, 3.0, 3.0},
			pt:       [2]float64{1.0, 1.0},
			expected: true,
		},
		"uncontained point": {
			bb:       &geom.Extent{0.0, 0.0, 3.0, 3.0},
			pt:       [2]float64{-1.0, -1.0},
			expected: false,
		},
		"nil bb": {
			bb:       nil,
			pt:       [2]float64{-1.0, -1.0},
			expected: true,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

// TestExtentAttributes check that the extent is returning the correct values for the different
// attributes that a extent can have.
func TestExtentAttributes(t *testing.T) {
	bblncmp := func(pt [2]float64, x, y float64) bool {
		return pt[0] == x && pt[1] == y
	}

	type tcase struct {
		bb           *geom.Extent
		xspan, yspan float64
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			t.Parallel()
			bb := tc.bb

			{
				vert := bb.Vertices()

				if len(vert) != 4 {
					t.Errorf("vertices length, expected %v got %v", 4, len(vert))
					return
				}
				if !bblncmp(vert[0], bb.MinX(), bb.MinY()) {
					t.Errorf("vert0] top left, expected %v got %v", [2]float64{bb.MinX(), bb.MinY()}, vert[0])
					return
				}
				if !bblncmp(vert[1], bb.MaxX(), bb.MinY()) {
					t.Errorf("vert[1] top right, expected %v got %v", [2]float64{bb.MaxX(), bb.MinY()}, vert[1])
					return
				}
				if !bblncmp(vert[2], bb.MaxX(), bb.MaxY()) {
					t.Errorf("vert[2] bottom right, expected %v, got %v", [2]float64{bb.MaxX(), bb.MaxY()}, vert[2])
					return
				}
				if !bblncmp(vert[3], bb.MinX(), bb.MaxY()) {
					t.Errorf("vert[3], expected %v got %v", [2]float64{bb.MinX(), bb.MaxY()}, vert[3])
					return
				}
				edges := bb.Edges(nil)
				if len(edges) != 4 {
					t.Errorf("edges length, expected 4 got %v", len(edges))
					return
				} else {
					eedge := [][2][2]float64{
						{vert[0], vert[1]},
						{vert[1], vert[2]},
						{vert[2], vert[3]},
						{vert[3], vert[0]},
					}
					if !reflect.DeepEqual(edges, eedge) {
						t.Errorf("edges, expected %v got %v", eedge, edges)
						return
					}
				}
				edges = bb.Edges(func(_ ...[2]float64) bool { return false })
				if len(edges) != 4 {
					t.Errorf("edges length, expected 4 got %v", len(edges))
					return
				} else {
					eedge := [][2][2]float64{
						{vert[3], vert[2]},
						{vert[2], vert[1]},
						{vert[1], vert[0]},
						{vert[0], vert[3]},
					}
					if !reflect.DeepEqual(edges, eedge) {
						t.Errorf("edges, expected %v got %v", eedge, edges)
						return
					}
				}
				poly := bb.AsPolygon()
				epoly := geom.Polygon{vert}
				if !reflect.DeepEqual(epoly, poly) {
					t.Errorf("as polygon, expected %v got %v", epoly, poly)
					return
				}
			}

			minx, miny, maxx, maxy := -math.MaxFloat64, -math.MaxFloat64, math.MaxFloat64, math.MaxFloat64
			if bb != nil {
				minx, miny, maxx, maxy = bb[0], bb[1], bb[2], bb[3]
			}

			if minx > maxx {
				minx, maxx = maxx, minx
			}
			if miny > maxy {
				miny, maxy = maxy, miny
			}

			if maxx != bb.MaxX() {
				t.Errorf("maxx, expected %v, got %v", maxx, bb.MaxX())
			}
			if minx != bb.MinX() {
				t.Errorf("minx, expected %v, got %v", minx, bb.MinX())
			}
			if maxy != bb.MaxY() {
				t.Errorf("maxy, expected %v, got %v", maxy, bb.MaxY())
			}
			if miny != bb.MinY() {
				t.Errorf("miny, expected %v, got %v", miny, bb.MinY())
			}
			cbb := bb.Clone()
			if !cmp.GeomExtent(bb, cbb) {
				t.Errorf("Clone equal, expected (%v) true got (%v) false", bb, cbb)
			}
			xspan := bb.XSpan()
			if tc.xspan == math.Inf(1) && xspan != tc.xspan {
				t.Errorf("xspan, expected ∞ got %v", xspan)
			} else {
				if !cmp.Float(tc.xspan, xspan) {
					t.Errorf("xspan, expected %v got %v", tc.xspan, xspan)
				}
			}
			yspan := bb.YSpan()
			if tc.yspan == math.Inf(1) && yspan != tc.yspan {
				t.Errorf("yspan, expected ∞ got %v", yspan)
			} else {
				if !cmp.Float(tc.yspan, yspan) {
					t.Errorf("yspan, expected %v got %v", tc.yspan, yspan)
				}
			}
		}
	}
	tests := map[string]tcase{
		"std": {
			bb:    &geom.Extent{0.0, 0.0, 10.0, 10.0},
			xspan: 10.0,
			yspan: 10.0,
		},
		"nil": {
			bb:    nil,
			xspan: math.Inf(1),
			yspan: math.Inf(1),
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestExtentScaleBy(t *testing.T) {
	type tcase struct {
		bb    *geom.Extent
		scale float64
		ebb   *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {
		sbb := tc.bb.ScaleBy(tc.scale)
		if !cmp.GeomExtent(tc.ebb, sbb) {
			t.Errorf("Scale by, expected %v got %v", tc.ebb, sbb)
		}
	}
	tests := map[string]tcase{
		"nil": {
			scale: 2.0,
		},
		"1.0 scale": {
			bb:    &geom.Extent{0, 0, 10, 10},
			ebb:   &geom.Extent{0, 0, 10, 10},
			scale: 1.0,
		},
		"2.0 scale": {
			bb:    &geom.Extent{0, 0, 10, 10},
			ebb:   &geom.Extent{0, 0, 20, 20},
			scale: 2.0,
		},
		"-2.0 scale": {
			bb:    &geom.Extent{0, 0, 10, 10},
			ebb:   &geom.Extent{-20, -20, 0, 0},
			scale: -2.0,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentExpandBy(t *testing.T) {
	type tcase struct {
		bb     *geom.Extent
		factor float64
		ebb    *geom.Extent
	}
	fn := func(t *testing.T, tc tcase) {
		sbb := tc.bb.ExpandBy(tc.factor)
		if !cmp.GeomExtent(tc.ebb, sbb) {
			t.Errorf("Expand by, expected %v got %v", tc.ebb, sbb)
		}
	}
	tests := map[string]tcase{
		"nil": {
			factor: 2.0,
		},
		"1.0 factor": {
			bb:     &geom.Extent{0, 0, 10, 10},
			ebb:    &geom.Extent{-1, -1, 11, 11},
			factor: 1.0,
		},
		"-20.1 factor": {
			bb:     &geom.Extent{0, 0, 10, 10},
			ebb:    &geom.Extent{-10.1, -10.1, 20.1, 20.1},
			factor: -20.1,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentIntersect(t *testing.T) {
	type tcase struct {
		// bb is the extent we are going to call Intersect on
		bb *geom.Extent
		// nbb is the extent passed to Intersect
		nbb *geom.Extent
		// ibb is the expected intersect extent
		ibb *geom.Extent
		// does it intersect or not?
		does bool
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			gbb, does := tc.bb.Intersect(tc.nbb)
			if does != tc.does {
				t.Errorf(" Intersect does, expected %v got %v", tc.does, does)
				return
			}
			if !cmp.GeomExtent(tc.ibb, gbb) {
				t.Errorf(" Intersect, expected %v got %v", tc.ibb, gbb)
				return
			}
		}
	}
	tests := map[string]tcase{
		"nil": {
			bb:   nil,
			nbb:  nil,
			ibb:  nil,
			does: true,
		},
		"bb not nil": {
			bb:   &geom.Extent{10, 10, 20, 20},
			nbb:  nil,
			ibb:  &geom.Extent{10, 10, 20, 20},
			does: true,
		},
		"1": {
			bb:   &geom.Extent{10, 10, 20, 20},
			nbb:  &geom.Extent{10, 10, 15, 15},
			ibb:  &geom.Extent{10, 10, 15, 15},
			does: true,
		},
		"2": {
			bb:   &geom.Extent{10, 10, 15, 15},
			nbb:  &geom.Extent{10, 10, 20, 20},
			ibb:  &geom.Extent{10, 10, 15, 15},
			does: true,
		},
		"3": {
			bb:   &geom.Extent{10, 10, 15, 15},
			nbb:  &geom.Extent{15, 15, 20, 20},
			ibb:  nil,
			does: false,
		},
		"4": {
			bb:   &geom.Extent{10, 10, 15, 15},
			nbb:  &geom.Extent{10, 15, 20, 20},
			ibb:  nil,
			does: false,
		},
		"Henry": {
			bb: geom.NewExtent(
				[2]float64{1.28694046060967e+06, 6.13880758852643e+06},
				[2]float64{1.28696919030943e+06, 6.1388302432236e+06},
			),
			nbb: geom.NewExtent(
				[2]float64{1.2869561422558832e+06, 6.138783662354196e+06},
				[2]float64{1.2869942394710402e+06, 6.138821397139203e+06},
			),
			ibb: &geom.Extent{
				1.2869561422558832e+06, 6.13880758852643e+06,
				1.28696919030943e+06, 6.138821397139203e+06,
			},
			does: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

func TestExtentArea(t *testing.T) {
	maxarea := math.Inf(1)
	type tcase struct {
		bb   *geom.Extent
		area float64
	}
	fn := func(t *testing.T, tc tcase) {
		a := tc.bb.Area()
		if !cmp.Float(tc.area, a) {
			t.Errorf("area, expected %v got %v", tc.area, a)
		}
	}
	tests := map[string]tcase{
		"nil": {
			bb:   nil,
			area: maxarea,
		},
		"simple 10x10": {
			bb:   geom.NewExtent([2]float64{0, 0}, [2]float64{10, 10}),
			area: 100,
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtentContainsLine(t *testing.T) {
	type tcase struct {
		bb *geom.Extent
		l  [2][2]float64
		e  bool
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			if got := tc.bb.ContainsLine(tc.l); got != tc.e {
				t.Errorf("contains line, expected %v got %v", tc.e, got)
			}
		}
	}
	tests := map[string]tcase{
		"nil": {
			l: [2][2]float64{{0, 0}, {10, 10}},
			e: true,
		},
		"contained": {
			bb: &geom.Extent{-1, -1, 20, 20},
			l:  [2][2]float64{{0, 0}, {10, 10}},
			e:  true,
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}

}
func TestNewExtentFromGeometry(t *testing.T) {
	type tcase struct {
		g   geom.Geometry
		e   *geom.Extent
		err error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			e, err := geom.NewExtentFromGeometry(tc.g)
			if (tc.err != nil && err == nil) ||
				(tc.err == nil && err != nil) {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}
			if tc.err != nil {
				if tc.err.Error() != err.Error() {
					t.Errorf("error, expected %v got %v", tc.err, err)
					return
				}
				return
			}
			if (tc.e != nil && e == nil) ||
				(tc.e == nil && e != nil) {
				t.Errorf("extent, expected %v got %v", tc.e, e)
				return
			}
			if tc.e != nil {
				if tc.e[0] != e[0] ||
					tc.e[1] != e[1] ||
					tc.e[2] != e[2] ||
					tc.e[3] != e[3] {
					t.Errorf("extent, expected %v got %v", tc.e, e)
					return
				}
			}

		}
	}

	tests := map[string]tcase{

		"Henery Circle One": {
			g: &geom.MultiPolygon{
				{ // Polygon
					{ // Ring
						{1286956.1422558832, 6138803.15957211},
						{1286957.5138675969, 6138809.6399925},
						{1286961.0222077654, 6138815.252628375},
						{1286966.228733862, 6138819.3396373615},
						{1286972.5176202222, 6138821.397139203},
						{1286979.1330808033, 6138821.173193399},
						{1286985.2820067848, 6138818.695793352},
						{1286990.1992814348, 6138814.272866236},
						{1286993.3157325392, 6138808.436285537},
						{1286994.2394710402, 6138801.885883152},
						{1286992.8678593265, 6138795.40546864},
						{1286989.3781805448, 6138789.792845784},
						{1286984.1623237533, 6138785.719847463},
						{1286977.864106701, 6138783.662354196},
						{1286971.2486461198, 6138783.872302467},
						{1286965.1183815224, 6138786.349692439},
						{1286960.1824454917, 6138790.7726051165},
						{1286957.084655768, 6138796.623170342},
						{1286956.1422558832, 6138803.15957211},
					},
				},
			},
			e: &geom.Extent{1.2869561422558832e+06, 6.138783662354196e+06, 1.2869942394710402e+06, 6.138821397139203e+06},
		},
	}

	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
