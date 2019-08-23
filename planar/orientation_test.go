package planar

import (
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"strconv"
	"testing"
)

func TestIsCCW(t *testing.T) {
	type tcase struct {
		desc       string
		p1, p2, p3 geom.Point
		is         bool
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := IsCCW(tc.p1, tc.p2, tc.p3)
			if got != tc.is {
				t.Errorf(
					"%v:%v:%v, expected %v got %v",
					wkt.MustEncode(tc.p1), wkt.MustEncode(tc.p2), wkt.MustEncode(tc.p3),
					tc.is, got,
				)
				return
			}
		}
	}

	tests := []tcase{
		{
			p1: geom.Point{0, 0},
			p2: geom.Point{1, 0},
			p3: geom.Point{1, 1},
			is: false,
		},
		{
			p1: geom.Point{204, 694},
			p2: geom.Point{-2511, -3640},
			p3: geom.Point{3462, -3640},
			is: false,
		},
		{
			p1: geom.Point{3462, -3640},
			p2: geom.Point{204, 694},
			p3: geom.Point{-2511, -3640},
			is: false,
		},
		{
			p1: geom.Point{-2511, -3640},
			p2: geom.Point{3462, -3640},
			p3: geom.Point{204, 694},
			is: false,
		},
		{
			p1: geom.Point{-2511, -3640},
			p2: geom.Point{204, 694},
			p3: geom.Point{3462, -3640},
			is: true,
		},
		{
			p1: geom.Point{204, 694},
			p2: geom.Point{3462, -3640},
			p3: geom.Point{-2511, -3640},
			is: true,
		},
		{
			desc: "from n america",
			// POINT (854.210 1424.142) POINT (853.491 1424.329) POINT (852.395 1424.635)
			p1: geom.Point{854.210, 1424.142},
			p2: geom.Point{853.491, 1424.329},
			p3: geom.Point{852.395, 1424.635},
			is: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, fn(tc))
	}
}

func TestOrientationInRegardsTo(t *testing.T) {
	type tcase struct {
		Desc               string
		origin, p1, p2, p3 geom.Point
		orientation        Orientation
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			t.Run("", func(t *testing.T) {
				got := OrientationInRegardsTo(tc.origin, tc.p1, tc.p2, tc.p3)
				if got != tc.orientation {
					t.Logf("points: %v %v %v",
						wkt.MustEncode(tc.p1),
						wkt.MustEncode(tc.p2),
						wkt.MustEncode(tc.p3),
					)
					t.Errorf("orientation, expected %v, got %v", tc.orientation, got)
				}
			})

		}
	}

	tests := []tcase{
		// subtests
		{
			Desc:        "oooo-cl",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{0, 0},
			p2:          geom.Point{0, 0},
			p3:          geom.Point{0, 0},
			orientation: CoLinearOrientation,
		},
		{
			Desc:        "oaaa-cl",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{1, 1},
			p2:          geom.Point{1, 1},
			p3:          geom.Point{1, 1},
			orientation: CoLinearOrientation,
		},
		{
			Desc:        "oaac-cl",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{1, 1},
			p2:          geom.Point{1, 1},
			p3:          geom.Point{-2, -2},
			orientation: CoLinearOrientation,
		},
		{
			Desc:        "oabc-cl",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{1, 1},
			p2:          geom.Point{-1, -1},
			p3:          geom.Point{-2, -2},
			orientation: CoLinearOrientation,
		},
		{
			Desc:        "oobc-ccw",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{0, 0},
			p2:          geom.Point{0, 1},
			p3:          geom.Point{-1, 0},
			orientation: CounterClockwiseOrientation,
		},
		{
			Desc:        "oobc-ccw",
			origin:      geom.Point{0, 0},
			p1:          geom.Point{0, 0},
			p2:          geom.Point{0, 1},
			p3:          geom.Point{1, 0},
			orientation: CounterClockwiseOrientation,
		},
	}

	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}

}

func TestIsCCWWithRegardsTo(t *testing.T) {
	type tcase struct {
		Desc               string
		origin, p1, p2, p3 geom.Point
		is                 bool
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			for i, pts := range [][3]geom.Point{{tc.p1, tc.p2, tc.p3}, {tc.p2, tc.p3, tc.p1}, {tc.p3, tc.p1, tc.p2}} {
				t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) {
					got := IsCCWWithRegardsTo(tc.origin, pts[0], pts[1], pts[2])
					if got != tc.is {
						t.Logf("points: %v %v %v",
							wkt.MustEncode(pts[0]),
							wkt.MustEncode(pts[1]),
							wkt.MustEncode(pts[2]),
						)
						t.Errorf("counter clockwise, expected %v, got %v", tc.is, got)
					}
				})
			}

		}
	}

	tests := []tcase{
		// subtests
		{
			Desc:   "oooo-cl",
			origin: geom.Point{0, 0},
			p1:     geom.Point{0, 0},
			p2:     geom.Point{0, 0},
			p3:     geom.Point{0, 0},
			is:     false,
		},
		{
			Desc:   "oaaa-cl",
			origin: geom.Point{0, 0},
			p1:     geom.Point{1, 1},
			p2:     geom.Point{1, 1},
			p3:     geom.Point{1, 1},
			is:     false,
		},
		{
			Desc:   "oaac-cl",
			origin: geom.Point{0, 0},
			p1:     geom.Point{1, 1},
			p2:     geom.Point{1, 1},
			p3:     geom.Point{-2, -2},
			is:     false,
		},
		{
			Desc:   "oabc-cl",
			origin: geom.Point{0, 0},
			p1:     geom.Point{1, 1},
			p2:     geom.Point{-1, -1},
			p3:     geom.Point{-2, -2},
			is:     false,
		},
		{
			Desc:   "oabc-cl",
			origin: geom.Point{0, 0},
			p1:     geom.Point{1, 1},
			p2:     geom.Point{-1, -1},
			p3:     geom.Point{-2, -2},
			is:     false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.Desc, fn(tc))
	}
}
