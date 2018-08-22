package quadedge

import (
	"strconv"
	"testing"
)

func TestIsInCircleRobust(t *testing.T) {
	type tcase struct {
		circle   [3]Vertex
		point    Vertex
		expected bool
	}

	fn := func(t *testing.T, tc tcase) {
		r := TrianglePredicate.IsInCircleRobust(tc.circle[0], tc.circle[1], tc.circle[2], tc.point)
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{
			circle:   [3]Vertex{{10, 10}, {15, 120}, {10, 20}},
			point:    Vertex{20, 10},
			expected: true,
		},
		{
			circle:   [3]Vertex{{15, 120}, {10, 10}, {20, 20}},
			point:    Vertex{10, 20},
			expected: true,
		},
		{
			circle:   [3]Vertex{{10, 10}, {15, 120}, {10, 20}},
			point:    Vertex{0, 10},
			expected: false,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
