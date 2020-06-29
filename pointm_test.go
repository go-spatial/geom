package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointMSetter(t *testing.T) {
	type tcase struct {
		point    [3]float64
		setter   geom.PointMSetter
		expected geom.PointMSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYM(tc.point)
			if tc.err == nil && err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if tc.err != nil {
				if tc.err.Error() != err.Error() {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			// compare the results
			if !reflect.DeepEqual(tc.expected, tc.setter) {
				t.Errorf("setter, expected %v got %v", tc.expected, tc.setter)
				return
			}
			xym := tc.setter.XYM()
			if !reflect.DeepEqual(tc.point, xym) {
				t.Errorf("XYM, expected %v, got %v", tc.point, xym)
			}
		}
	}
	tests := []tcase{
		{
			point:    [3]float64{10, 20, 1000.},
			setter:   &geom.PointM{15, 20, 1000.},
			expected: &geom.PointM{10, 20, 1000.},
		},
		{
			setter: (*geom.PointM)(nil),
			err:    geom.ErrNilPointM,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointM(t *testing.T) {
	fn := func(pt geom.PointM) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xy", func(t *testing.T) {
					xy := pt.XY()
					exp_xy := geom.Point{pt[0], pt[1]}
					if xy != exp_xy {
						t.Errorf("xy, expected %v got %v", exp_xy, xy)
					}
				})
				t.Run("xym", func(t *testing.T) {
					xym := pt.XYM()
					exp_xym := pt
					if xym != exp_xym {
						t.Errorf("xym, expected %v got %v", exp_xym, xym)
					}
				})
				t.Run("m", func(t *testing.T) {
					m := pt.M()
					exp_m := pt[2]
					if m != exp_m {
						t.Errorf("m, expected %v got %v", exp_m, m)
					}
				})
			}
	}
	tests := []geom.PointM{
		{0, 1, 1000.}, {2, 2, 1000.}, {1, 2, 1000.},
	}
	for _, pt := range tests {
		t.Run(fn(pt))
	}
}
