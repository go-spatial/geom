package planar

import (
	"math"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

// TestDistanceToLineSegment tests the distance from a line segment to a point.
func TestDistanceToLineSegment(t *testing.T) {
	type tcase struct {
		p        geom.Pointer
		v        geom.Pointer
		w        geom.Pointer
		expected float64
	}

	fn := func(t *testing.T, tc tcase) {
		r := DistanceToLineSegment(tc.p, tc.v, tc.w)
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{geom.Point{-1, 0}, geom.Point{0, 0}, geom.Point{1, 1}, 1},
		{geom.Point{0, 1}, geom.Point{.5, .5}, geom.Point{.5, .5}, 0.7071067811865476},
		{geom.Point{0, 1}, geom.Point{.5, .5}, geom.Point{.6, .4}, 0.7071067811865476},
		{geom.Point{0, 1}, geom.Point{.6, .4}, geom.Point{.5, .5}, 0.7071067811865476},
		{geom.Point{0, 1}, geom.Point{0, 0}, geom.Point{1, 1}, 0.7071067811865476},
		{geom.Point{1, 1}, geom.Point{0, 2}, geom.Point{2, 2}, 1},
		{geom.Point{0, 2}, geom.Point{0, 2}, geom.Point{2, 2}, 0},
		{geom.Point{2, 2}, geom.Point{0, 2}, geom.Point{2, 2}, 0},
		{geom.Point{2, 3}, geom.Point{1, 1}, geom.Point{1, 1}, 2.23606797749979},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestPointDistance(t *testing.T) {
	type tcase struct {
		point1   geom.Point
		point2   geom.Point
		expected float64
		err      error
	}

	fn := func(t *testing.T, tc tcase) {
		d := PointDistance(tc.point1, tc.point2)
		if d != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, d)
			return
		}

		d = PointDistance2(tc.point1, tc.point2)
		if math.Abs(d-tc.expected*tc.expected) > 1e-6 {
			t.Errorf("error, expected %v got %v", tc.expected*tc.expected, d)
			return
		}
	}
	testcases := []tcase{
		{
			point1:   geom.Point{2, 2},
			point2:   geom.Point{3, 4},
			expected: 2.23606797749979,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
