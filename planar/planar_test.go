package planar

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestSlope(t *testing.T) {
	type tcase struct {
		line    [2][2]float64
		m, b    float64
		defined bool
	}

	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		gm, gb, gd := Slope(tc.line)
		if tc.defined != gd {
			t.Errorf("sloped defined, expected %v got %v", tc.defined, gd)
			return
		}
		// if the slope is not defined, line is verticle and m,b don't have good values.
		if !tc.defined {
			return
		}
		if tc.m != gm {
			t.Errorf("sloped, expected %v got %v", tc.m, gm)

		}
		if tc.b != gb {
			t.Errorf("sloped intercept, expected %v got %v", tc.b, gb)
		}
	}
	tests := []tcase{
		{
			line:    [2][2]float64{{0, 0}, {10, 10}},
			m:       1,
			b:       0,
			defined: true,
		},
		{
			line:    [2][2]float64{{1, 7}, {1, 17}},
			defined: false,
		},
	}
	for i := range tests {
		tc := tests[i]
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestIsPointOnLineSegment(t *testing.T) {
	type tcase struct {
		desc     string
		point    geom.Point
		segment  geom.Line
		expected bool
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return fmt.Sprintf("%v on %v", tc.point, tc.segment),
			func(t *testing.T) {
				if tc.expected != IsPointOnLineSegment(tc.point, tc.segment) {
					t.Errorf("got %v, expected %v", !tc.expected, tc.expected)
				}
			}
	}
	tests := [...]tcase{
		{
			// Diagonal line
			point:   geom.Point{1, 1},
			segment: geom.Line{{0, 0}, {1, 10}},
		},
		{
			// Vertical line
			point:   geom.Point{1, 1},
			segment: geom.Line{{0, 0}, {0, 10}},
		},
		{
			// Vertical line
			point:   geom.Point{1, 1},
			segment: geom.Line{{0, 10}, {10, 10}},
		},
		{
			// horizontal line
			point:    geom.Point{1, 1},
			segment:  geom.Line{{1, 0}, {1, 10}},
			expected: true,
		},
		{
			// horizontal line on close to the end point
			point:   geom.Point{-0.5, 0},
			segment: geom.Line{{1, 0}, {1, 10}},
		},
		{
			// horizontal line on the end point
			point:    geom.Point{1, 0},
			segment:  geom.Line{{1, 0}, {1, 10}},
			expected: true,
		},
	}
	for _, tc := range tests {
		t.Run(fn(tc))
	}
}
