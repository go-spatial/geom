package windingorder

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"testing"
)

func TestAttributeMethods(t *testing.T) {

	fn := func(val WindingOrder) func(*testing.T) {
		return func(t *testing.T) {

			var (
				// variables based on the type
				isClockwise        = false
				isCounterClockwise = false
				isColinear         = false
				notDir             = val
				str                = "unknown"
			)

			switch val {
			case Clockwise:
				isClockwise = true
				isCounterClockwise = false
				isColinear = false
				notDir = CounterClockwise
				str = "clockwise"

			case CounterClockwise:
				isClockwise = false
				isCounterClockwise = true
				isColinear = false
				notDir = Clockwise
				str = "counter clockwise"

			case Colinear:
				isClockwise = false
				isCounterClockwise = false
				isColinear = true
				notDir = Colinear
				str = "colinear"

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
		}
	}
	tests := []WindingOrder{Clockwise, CounterClockwise, Colinear, 3}
	for i := range tests {
		t.Run(tests[i].String(), fn(tests[i]))
	}
}

func TestOfPoints(t *testing.T) {
	type tcase struct {
		Desc  string
		pts   [][2]float64
		order WindingOrder
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			got := OfPoints(tc.pts...)
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

			got = OfGeomPoints(points...)
			if got != tc.order {
				t.Errorf("OfGeomPoints, expected %v got %v", tc.order, got)
			}

			// Test with yPostiveDown set to false
			got = Orientation(false, tc.pts...)
			if got != tc.order.Not() {
				t.Errorf("Orientation y-false, expected %v got %v", tc.order.Not(), got)
			}

		}
	}
	tests := [...]tcase{
		{
			Desc:  "simple points",
			pts:   [][2]float64{{0, 0}, {10, 0}, {10, 10}, {0, 10}},
			order: Clockwise,
		},
		{
			Desc:  "counter simple points",
			pts:   [][2]float64{{0, 10}, {10, 10}, {10, 0}, {0, 0}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {10, 0}, {0, 10}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {1, 0}, {0, 1}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {0, 10}, {10, 0}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {0, 1}, {1, 0}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{10, 0}, {10, 10}, {0, 10}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{0, 10}, {10, 10}, {10, 0}},
			order: CounterClockwise,
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
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{0, 0}, {1, 0}, {1, 1}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{204, 694}, {-2511, -3640}, {3462, -3660}},
			order: Clockwise,
		},
		{
			pts:   [][2]float64{{-2511, -3640}, {204, 694}, {3462, -3660}},
			order: CounterClockwise,
		},
		{
			pts:   [][2]float64{{204, 694}, {3462, -3660}, {-2511, -3640}},
			order: CounterClockwise,
		},
		{
			Desc: "from n america",
			pts: [][2]float64{
				{854.210, 1424.142},
				{853.491, 1424.329},
				{852.395, 1424.635},
			},
			order: CounterClockwise,
		},
		{
			Desc: "edge_test initial good",
			pts: [][2]float64{
				{375, 113},
				{372, 114},
				{368, 117},
				{384, 112},
			},
			order: CounterClockwise,
		},
		{
			Desc: "edge_test initial good",
			pts: [][2]float64{
				{365.513, 116.162},
				{366.318, 117.961},
				{384.939, 111.896},
			},
			order: CounterClockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{10,0},
				{0,20},
				{10,20},
			},
			order:CounterClockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{10,20},
				{10,0},
				{0,20},
			},
			order:CounterClockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{0,20},
				{10,20},
				{10,0},
			},
			order:CounterClockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{10,20},
				{0,20},
				{10,0},
			},
			order:Clockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{10,0},
				{10,20},
				{0,20},
			},
			order:Clockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{0,20},
				{10,0},
				{10,20},
			},
			order:Clockwise,
		},
		{
			Desc: "pseudo_polygon_test should not be colinear",
			pts: [][2]float64{
				{0,0},
				{0,-3},
				{-2,0},
			},
			order: CounterClockwise,
		},
		{
			Desc: "North Africa fullset",
			pts: [][2]float64{
				{4160    ,3104.315},
				{25280   ,-42304},
				{4160    ,4160},
				{4151.127,3897.246},
				{4040.404,3144.386},
				{4058.330,3145.647},
				{4086.840,3143.034},
				{4092.542,3142.665},
				{4140.334,3117.767},
				{4143.023,3116.169},
				{4157.268,3107.438},
				{4158.342,3106.261},
			},
			order: Clockwise,
		},
	}
	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}
