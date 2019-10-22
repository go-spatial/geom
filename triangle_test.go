package geom_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestNewTriangleForExtent(t *testing.T) {
	type tcase struct {
		Extent   *geom.Extent
		Buff     float64
		Expected geom.Triangle
		Err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			tri, err := geom.NewTriangleForExtent(tc.Extent, tc.Buff)
			if tc.Err != nil {
				if err != tc.Err {
					t.Errorf("error, expected %v, got %v ", tc.Err, err)
				}
				return
			}
			if err != nil {
				t.Errorf("error, expected nil, got %v ", err)
				return
			}

			if tri[0] != tc.Expected[0] || tri[1] != tc.Expected[1] || tri[2] != tc.Expected[2] {
				t.Errorf("triangle, expected %v, got %v ", tc.Expected, tri)
			}
		}
	}

	tests := [...]tcase{
		{
			Extent:   geom.NewExtent([2]float64{0, 0}),
			Buff:     10,
			Expected: geom.Triangle{{-10, -10}, {0, 10}, {10, -10}},
		},
	}
	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
