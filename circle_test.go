package geom_test

import (
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

const tolerance = geom.TOLERANCE

func TestCircleFromPoints(t *testing.T) {
	type tcase struct {
		p      [3][2]float64
		circle geom.Circle
		err    error
	}
	fn := func(t *testing.T, tc tcase) {
		circle, err := geom.CircleFromPoints(tc.p[0], tc.p[1], tc.p[2])
		if (tc.err != nil && err == nil) || (tc.err == nil && err != nil) {
			t.Errorf("error, expected %v got %v", tc.err, err)
			return
		}
		if tc.err != nil {
			if tc.err != err {
				t.Errorf("error, expected %v got %v", tc.err, err)
			}
			return
		}

		if !cmp.Float64(circle.Radius, tc.circle.Radius, tolerance) {
			t.Errorf("circle radius, expected %v got %v", tc.circle, circle)
			return
		}

		if !cmp.Float64(circle.Center[0], tc.circle.Center[0], tolerance) {
			t.Errorf("circle x, expected %v got %v", tc.circle, circle)
			return
		}

		if !cmp.Float64(circle.Center[1], tc.circle.Center[1], tolerance) {
			t.Errorf("circle y, expected %v got %v", tc.circle, circle)
			return
		}
	}

	tests := map[string]tcase{
		"simple colinear": {
			p:   [3][2]float64{{1, 0}, {1, 1}, {1, 20}},
			err: geom.ErrPointsAreCoLinear,
		},
		"center outside of triangle": {
			p:      [3][2]float64{{1, 0}, {10, 20}, {5, 5}},
			circle: geom.Circle{Center: [2]float64{-21.642857, 22.214286}, Radius: 31.7202},
		},
		"center outside of triangle 1": {
			p:      [3][2]float64{{1, 0}, {5, 5}, {10, 20}},
			circle: geom.Circle{Center: [2]float64{-21.642857, 22.214286}, Radius: 31.7202},
		},
		"center outside of triangle 2": {
			p:      [3][2]float64{{5, 5}, {1, 0}, {10, 20}},
			circle: geom.Circle{Center: [2]float64{-21.642857, 22.214286}, Radius: 31.7202},
		},
		"center right triangle": {
			p:      [3][2]float64{{1, 0}, {10, 0}, {10, 7}},
			circle: geom.Circle{Center: [2]float64{5.5, 3.5}, Radius: 5.7009},
		},
		"center right triangle 1": {
			p:      [3][2]float64{{10, 0}, {1, 0}, {10, 7}},
			circle: geom.Circle{Center: [2]float64{5.5, 3.5}, Radius: 5.7009},
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
