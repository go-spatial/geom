package subdivision_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/subdivision"
)

func TestClassify(t *testing.T) {

	type tcase struct {
		a, b, c  geometry.Point
		expected subdivision.QType
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := subdivision.Classify(tc.a, tc.b, tc.c)
			if got != tc.expected {
				t.Errorf("error, expected %v got %v", tc.expected, got)
				return
			}
		}
	}

	testcases := []tcase{
		{
			a:        geometry.NewPoint(1.1, 2.5),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.RIGHT,
		},
		{
			a:        geometry.NewPoint(0.9, 2.5),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.LEFT,
		},
		{
			a:        geometry.NewPoint(1, 1),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.BEHIND,
		},
		{
			a:        geometry.NewPoint(1, 4),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.BEYOND,
		},
		{
			a:        geometry.NewPoint(1, 2),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.ORIGIN,
		},
		{
			a:        geometry.NewPoint(1, 3),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.DESTINATION,
		},
		{
			a:        geometry.NewPoint(1, 2.5),
			b:        geometry.NewPoint(1, 2),
			c:        geometry.NewPoint(1, 3),
			expected: subdivision.BETWEEN,
		},
	}
	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}
