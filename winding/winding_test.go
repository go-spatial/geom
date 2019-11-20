package winding

import (
	"fmt"
	"testing"

	"github.com/go-spatial/geom/cmp"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/testing/must"
)

func TestHelperMethods(t *testing.T) {
	order := Order{}
	val := order.Clockwise()
	if val != Clockwise {
		t.Errorf("clockwise, expected clockwise got %v", val)
	}
	val = order.CounterClockwise()
	if val != CounterClockwise {
		t.Errorf("counter clockwise, expected counter clockwise got %v", val)
	}

	val = order.Colinear()
	if val != Colinear {
		t.Errorf("colinear, expected colinear got %v", val)
	}
	val = order.Collinear()
	if val != Colinear {
		t.Errorf("collinear, expected colinear got %v", val)
	}

}

func TestAttributeMethods(t *testing.T) {

	fn := func(val Winding) func(*testing.T) {
		return func(t *testing.T) {

			var (
				// variables based on the type
				isClockwise        = false
				isCounterClockwise = false
				isColinear         = false
				notDir             = -1 * val
				str                = fmt.Sprintf("unknown(%v)", int8(val))
				sstr               = fmt.Sprintf("{%v}", int8(val))
			)

			switch val {
			case Clockwise:
				isClockwise = true
				isCounterClockwise = false
				isColinear = false
				notDir = CounterClockwise
				str = "clockwise"
				sstr = "⟳"

			case CounterClockwise:
				isClockwise = false
				isCounterClockwise = true
				isColinear = false
				notDir = Clockwise
				str = "counter clockwise"
				sstr = "⟲"

			case Colinear:
				isClockwise = false
				isCounterClockwise = false
				isColinear = true
				notDir = Colinear
				str = "colinear"
				sstr = "O"

			}

			if val.IsClockwise() != isClockwise {
				t.Errorf("is clockwise, expected %v got %v", isClockwise, val.IsClockwise())
			}
			if val.IsCounterClockwise() != isCounterClockwise {
				t.Errorf("is counter clockwise, expected %v got %v", isCounterClockwise, val.IsCounterClockwise())
			}
			if val.IsColinear() != isColinear {
				t.Errorf("is colinear, expected %v got %v", isColinear, val.IsColinear())
			}

			if val.Not() != notDir {
				t.Errorf("not, expected %v got %v", notDir, val.Not())
			}
			if val.Not().Not() != val {
				t.Errorf("not not, expected %v got %v", val, val.Not().Not())
			}
			if val.String() != str {
				t.Errorf("string, expected %v got %v", val.String(), str)
			}
			if val.ShortString() != sstr {
				t.Errorf("string, expected %v got %v", val.ShortString(), sstr)
			}
		}
	}
	tests := []Winding{Clockwise, CounterClockwise, Colinear, 3}
	for i := range tests {
		t.Run(tests[i].String(), fn(tests[i]))
	}
	t.Run("fewer then 3 points", func(t *testing.T) {
		if val := Orient([][2]float64{{0, 0}, {1, 0}}...); val != 0 {
			t.Errorf("less then three point, expected %v got %v", 0, val)
		}
	})
}

func TestOfPoints(t *testing.T) {
	type tcase struct {
		Desc         string
		pts          [][2]float64
		order        Winding
		int64NotSame bool
	}
	order := Order{
		YPositiveDown: false,
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			got := order.OfPoints(tc.pts...)
			if got != tc.order {
				t.Errorf("order.OfPoints, expected %v got %v", tc.order, got)
				for i := range tc.pts {
					str, err := wkt.EncodeString(geom.Point(tc.pts[i]))
					if err != nil {
						panic(err)
					}
					t.Logf("%03v:%v", i, str)
				}
				return
			}
			if !tc.int64NotSame {
				int64pts := make([][2]int64, len(tc.pts))
				for i := range tc.pts {
					int64pts[i][0], int64pts[i][1] = int64(tc.pts[i][0]), int64(tc.pts[i][1])
				}
				got = order.OfInt64Points(int64pts...)
				if got != tc.order {
					t.Errorf("OfInt64Points, expected %v got %v", tc.order, got)
					for i := range int64pts {
						str, err := wkt.EncodeString(geom.Point{float64(int64pts[i][0]), float64(int64pts[i][1])})
						if err != nil {
							panic(err)
						}
						t.Logf("%03v:%v", i, str)
					}
					return
				}
			}

			got = OfPoints(tc.pts...)
			if got != tc.order {
				t.Errorf("OfPoints, expected %v got %v", tc.order, got)
				for i := range tc.pts {
					str, err := wkt.EncodeString(geom.Point(tc.pts[i]))
					if err != nil {
						panic(err)
					}
					t.Logf("%03v:%v", i, str)
				}
				return
			}

			points := make([]geom.Point, len(tc.pts))
			for i := range tc.pts {
				points[i] = geom.Point(tc.pts[i])
			}

			got = order.OfGeomPoints(points...)
			if got != tc.order {
				t.Errorf("order.OfGeomPoints, expected %v got %v", tc.order, got)
			}

			got = OfGeomPoints(points...)
			if got != tc.order {
				t.Errorf("OfGeomPoints, expected %v got %v", tc.order, got)
			}

			// Test with yPostiveDown set to false
			got = Orientation(!order.YPositiveDown, tc.pts...)
			if got != tc.order.Not() {
				t.Errorf("Orientation y-false, expected %v got %v", tc.order.Not(), got)
			}

		}
	}
	tests := [...]tcase{
		{
			Desc:  "simple points",
			pts:   [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			order: CounterClockwise,
		},
		{
			Desc:  "counter simple points",
			pts:   [][2]float64{{0, 10}, {10, 10}, {10, 0}, {0, 0}},
			order: Clockwise,
		},
		{
			Desc:  "not colinear",
			pts:   [][2]float64{{20, 10}, {20, 0}, {0, 10}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {10, 0}, {0, 10}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {1, 0}, {0, 1}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {0, 10}, {10, 0}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {0, 1}, {1, 0}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{10, 0}, {10, 10}, {0, 10}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 10}, {10, 10}, {10, 0}},
			order: Clockwise,
		},
		{
			Desc:  "colinear",
			pts:   [][2]float64{{0, 0}, {0, 1}, {0, 2}},
			order: Colinear, // This is really colinear
		},
		{
			Desc:  "colinear",
			pts:   [][2]float64{{0, 0}, {0, 2}, {0, 1}},
			order: Colinear, // This is really colinear
		},
		{
			Desc:  "empty",
			order: Colinear,
		},
		{
			Desc:  "one",
			pts:   [][2]float64{{0, 0}},
			order: Colinear,
		},
		{
			Desc:  "two",
			pts:   [][2]float64{{0, 0}, {0, 1}},
			order: Colinear,
		},
		{
			Desc:  "3-true",
			pts:   [][2]float64{{0, 0}, {0, 1}, {0, 2}},
			order: Colinear,
		},
		{
			Desc:  "3-false",
			pts:   [][2]float64{{0, 0}, {0, 1}, {1, 2}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {1, 0}, {1, 1}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{204, 694}, {-2511, -3640}, {3462, -3660}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{-2511, -3640}, {204, 694}, {3462, -3660}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{204, 694}, {3462, -3660}, {-2511, -3640}},
			order: Clockwise,
		},
		{
			Desc: "from n america",
			pts: [][2]float64{
				{854.210, 1424.142},
				{853.491, 1424.329},
				{852.395, 1424.635},
			},
			order:        Clockwise,
			int64NotSame: true,
		},
		{
			Desc: "edge_test initial good",
			pts: [][2]float64{
				{375, 113},
				{372, 114},
				{368, 117},
				{384, 112},
			},
			order: Clockwise,
		},
		{
			Desc: "edge_test initial good",
			pts: [][2]float64{
				{365.513, 116.162},
				{366.318, 117.961},
				{384.939, 111.896},
			},
			order: Clockwise,
		},
		{
			Desc: "test_rectify_polygon #1_polygon_ring_0",
			pts: [][2]float64{
				{0, 0},
				{10, 0},
				{0, 10},
				{0, 0},
			},
			order: CounterClockwise,
		},
		{
			Desc: "test_rectify_polygon #1_polygon_ring_1",
			pts: [][2]float64{
				{1, 1},
				{1, 2},
				{2, 1},
				{1, 1},
			},
			order: Clockwise,
		},
	}
	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestRectifyPolygon(t *testing.T) {
	type tcase struct {
		Polygon  geom.Polygon
		Expected geom.Polygon
	}
	var order Order
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := order.RectifyPolygon([][][2]float64(tc.Polygon))
			if !cmp.PolygonEqual(got, [][][2]float64(tc.Expected)) {
				t.Errorf("polygon, expected: %v got %v", wkt.MustEncode(tc.Expected), wkt.MustEncode(geom.Polygon(got)))
			}
		}
	}

	tests := map[string]tcase{
		"#1": {
			Polygon:  must.AsPolygon(must.Decode(wkt.DecodeString(`POLYGON((0 0,10 0,0 10,0 0),(1 1,2 1,1 2,1 1),(1 1,1 2,1 3,1 1))`))),
			Expected: must.AsPolygon(must.Decode(wkt.DecodeString(`POLYGON((0 0,0 10,10 0,0 0),(1 1,2 1,1 2,1 1))`))),
		},
		"#2": {
			Polygon: must.AsPolygon(must.Decode(wkt.DecodeString(`POLYGON((1 1,1 2,1 3,1 1))`))),
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}
