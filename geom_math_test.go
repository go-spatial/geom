package geom_test

import (
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

func TestRoundToPrec(t *testing.T) {

	type tcase struct {
		Value    float64
		Prec     int
		Expected float64
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			v := geom.RoundToPrec(tc.Value, tc.Prec)
			if !cmp.Float(tc.Expected, v) {
				t.Errorf("rounded, expected %v got %v", tc.Expected, v)
			}
		}
	}

	tests := [...]tcase{
		{
			Value:    0.001,
			Prec:     2,
			Expected: 0.0,
		},
		{
			Value:    0.005,
			Prec:     2,
			Expected: 0.01,
		},
		{
			Value:    0.001,
			Prec:     0,
			Expected: 0.0,
		},
		{
			Value:    -0.0,
			Prec:     0,
			Expected: 0.0,
		},
	}
	for _, tc := range tests {
		t.Run(strconv.FormatFloat(tc.Value, 'E', -1, 64), fn(tc))
	}

}
