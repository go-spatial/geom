package planar

import (
	"strconv"
	"testing"
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
