package subdivision_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/subdivision"
)

func init() {
	debugger.DefaultOutputDir = "output"
}

func TestClassify(t *testing.T) {

	type tcase struct {
		a, b, c  geom.Point
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
			a:        geom.Point{1.1, 2.5},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.RIGHT,
		},
		{
			a:        geom.Point{0.9, 2.5},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.LEFT,
		},
		{
			a:        geom.Point{1, 1},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.BEHIND,
		},
		{
			a:        geom.Point{1, 4},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.BEYOND,
		},
		{
			a:        geom.Point{1, 2},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.ORIGIN,
		},
		{
			a:        geom.Point{1, 3},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.DESTINATION,
		},
		{
			a:        geom.Point{1, 2.5},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: subdivision.BETWEEN,
		},
	}
	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}
