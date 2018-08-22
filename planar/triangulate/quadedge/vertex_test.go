package quadedge

import (
	"encoding/json"
	"strconv"
	"testing"
)

/*
TestVertexClassify tests for basic classification values. The test is quite simplistic
in that it only tests against a vertical vector, but should be good enough for a sniff
test.
*/
func TestVertexClassify(t *testing.T) {
	type tcase struct {
		u        Vertex
		p0       Vertex
		p1       Vertex
		expected int
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.u.Classify(tc.p0, tc.p1)
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{1.1, 2.5}, Vertex{1, 2}, Vertex{1, 3}, RIGHT},
		{Vertex{0.9, 2.5}, Vertex{1, 2}, Vertex{1, 3}, LEFT},
		{Vertex{1, 1}, Vertex{1, 2}, Vertex{1, 3}, BEHIND},
		{Vertex{1, 4}, Vertex{1, 2}, Vertex{1, 3}, BEYOND},
		{Vertex{1, 2}, Vertex{1, 2}, Vertex{1, 3}, ORIGIN},
		{Vertex{1, 3}, Vertex{1, 2}, Vertex{1, 3}, DESTINATION},
		{Vertex{1, 2.5}, Vertex{1, 2}, Vertex{1, 3}, BETWEEN},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexEquals(t *testing.T) {
	type tcase struct {
		v1       Vertex
		v2       Vertex
		expected bool
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.v1.Equals(tc.v2)
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{
			v1:       Vertex{1, 2},
			v2:       Vertex{1, 2},
			expected: true,
		},
		{
			v1:       Vertex{1, 2},
			v2:       Vertex{2, 3},
			expected: false,
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexEqualsTolerance(t *testing.T) {
	type tcase struct {
		v1        Vertex
		v2        Vertex
		tolerance float64
		expected  bool
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.v1.EqualsTolerance(tc.v2, tc.tolerance)
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{1, 2}, Vertex{1, 2}, 0.1, true},
		{Vertex{1, 2}, Vertex{1.09, 2}, 0.1, true},
		{Vertex{1, 2}, Vertex{1.1, 2}, 0.1, false},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexIsInCircle(t *testing.T) {
	type tcase struct {
		v1       Vertex
		expected bool
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.v1.IsInCircle(Vertex{0, 0}, Vertex{2, 0}, Vertex{1, 1})
		if r != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{.5, .5}, true},
		{Vertex{-1, 0}, false},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexMarshalJSON(t *testing.T) {
	type tcase struct {
		v1       Vertex
		expected string
	}

	fn := func(t *testing.T, tc tcase) {
		r, err := json.Marshal(tc.v1)
		if err != nil {
			t.Errorf("error, expected nil got %v", err)
		}
		if string(r) != tc.expected {
			t.Errorf("error, expected %v got %v", tc.expected, string(r))
			return
		}
	}
	testcases := []tcase{
		{Vertex{1, 2}, `{"X":1,"Y":2}`},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexScalar(t *testing.T) {
	type tcase struct {
		v      Vertex
		scalar float64
		times  Vertex
		divide Vertex
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.v.Times(tc.scalar)
		if r.Equals(tc.times) == false {
			t.Errorf("error, expected %v got %v", tc.times, r)
			return
		}

		r = tc.v.Divide(tc.scalar)
		if r.Equals(tc.divide) == false {
			t.Errorf("error, expected %v got %v", tc.divide, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{1, 2}, 3, Vertex{3, 6}, Vertex{0.3333333333333333, 0.6666666666666666}},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexUnary(t *testing.T) {
	type tcase struct {
		v         Vertex
		cross     Vertex
		magn      float64
		normalize Vertex
	}

	fn := func(t *testing.T, tc tcase) {
		r := tc.v.Cross()
		if tc.cross.Equals(r) == false {
			t.Errorf("error, expected %v got %v", tc.cross, r)
			return
		}

		m := tc.v.Magn()
		if tc.magn != m {
			t.Errorf("error, expected %v got %v", tc.magn, m)
			return
		}

		r = tc.v.Normalize()
		if tc.normalize.Equals(r) == false {
			t.Errorf("error, expected %v got %v", tc.normalize, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{1, 2}, Vertex{2, -1}, 2.23606797749979, Vertex{0.4472135954999579, 0.8944271909999159}},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestVertexVertex(t *testing.T) {
	type tcase struct {
		v1  Vertex
		v2  Vertex
		dot float64
		sum Vertex
	}

	fn := func(t *testing.T, tc tcase) {
		s := tc.v1.Dot(tc.v2)
		if s != tc.dot {
			t.Errorf("error, expected %v got %v", tc.dot, s)
			return
		}

		r := tc.v1.Sum(tc.v2)
		if r.Equals(tc.sum) == false {
			t.Errorf("error, expected %v got %v", tc.sum, r)
			return
		}
	}
	testcases := []tcase{
		{Vertex{1, 2}, Vertex{3, 4}, 11, Vertex{4, 6}},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
