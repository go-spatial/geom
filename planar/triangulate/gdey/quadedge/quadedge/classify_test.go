package quadedge_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"
)

func init() {
	debugger.DefaultOutputDir = "output"
}

func TestClassify(t *testing.T) {

	type tcase struct {
		a, b, c  geom.Point
		expected quadedge.QType
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			got := quadedge.Classify(tc.a, tc.b, tc.c)
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
			expected: quadedge.RIGHT,
		},
		{
			a:        geom.Point{0.9, 2.5},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.LEFT,
		},
		{
			a:        geom.Point{1, 1},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.BEHIND,
		},
		{
			a:        geom.Point{1, 4},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.BEYOND,
		},
		{
			a:        geom.Point{1, 2},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.ORIGIN,
		},
		{
			a:        geom.Point{1, 3},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.DESTINATION,
		},
		{
			a:        geom.Point{1, 2.5},
			b:        geom.Point{1, 2},
			c:        geom.Point{1, 3},
			expected: quadedge.BETWEEN,
		},
	}
	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}
