package simplify

import (
	"context"
	"flag"
	"reflect"
	"testing"
)

var ignoreSanityCheck bool

func init() {
	flag.BoolVar(&ignoreSanityCheck, "ignoreSanityCheck", false, "ignore sanity checks in test cases.")
}

func TestDouglasPeucker(t *testing.T) {
	type tcase struct {
		l  [][2]float64
		dp DouglasPeucker
		el [][2]float64
	}

	fn := func(t *testing.T, tc tcase) {
		ctx := context.Background()
		gl, err := tc.dp.Simplify(ctx, tc.l, false)
		// Douglas Peucker should never return an error.
		// This is more of a sanity check.
		if err != nil {
			t.Errorf("Douglas Peucker error, expected nil got %v", err)
			return
		}
		if !reflect.DeepEqual(tc.el, gl) {
			t.Errorf("simplified points, expected %v got %v", tc.el, gl)
			return
		}

		if ignoreSanityCheck {
			return
		}

		// Let's try it with true, it should not matter, as DP does not care.
		// More sanity checking.
		gl, _ = tc.dp.Simplify(ctx, tc.l, true)

		if !reflect.DeepEqual(tc.el, gl) {
			t.Errorf("simplified points (true), expected %v got %v", tc.el, gl)
			return
		}
	}

	tests := map[string]tcase{
		"simple box": {
			l: [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
			dp: DouglasPeucker{
				Tolerance: 0.001,
			},
			el: [][2]float64{{0, 0}, {0, 1}, {1, 1}, {1, 0}},
		},
	}

	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
