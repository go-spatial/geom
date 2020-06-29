package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointZMSetter(t *testing.T) {
	type tcase struct {
		point    [4]float64
		setter   geom.PointZMSetter
		expected geom.PointZMSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYZM(tc.point)
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
			xyzm := tc.setter.XYZM()
			if !reflect.DeepEqual(tc.point, xyzm) {
				t.Errorf("XYZM, expected %v, got %v", tc.point, xyzm)
			}
		}
	}
	tests := []tcase{
		{
			point:    [4]float64{10, 20, 30, 1000.},
			setter:   &geom.PointZM{15, 20, 30, 1000.},
			expected: &geom.PointZM{10, 20, 30, 1000.},
		},
		{
			setter: (*geom.PointZM)(nil),
			err:    geom.ErrNilPointZM,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointZM(t *testing.T) {
	fn := func(pt geom.PointZM) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xyz", func(t *testing.T) {
					xyz := pt.XYZ()
					exp_xyz := geom.PointZ{pt[0], pt[1], pt[2]}
					if xyz != exp_xyz {
						t.Errorf("xyz, expected %v got %v", exp_xyz, xyz)
					}
				})
				t.Run("xyzm", func(t *testing.T) {
					xyzm := pt.XYZM()
					exp_xyzm := pt
					if xyzm != exp_xyzm {
						t.Errorf("xyzm, expected %v got %v", exp_xyzm, xyzm)
					}
				})
				t.Run("m", func(t *testing.T) {
					m := pt.M()
					exp_m := pt[3]
					if m != exp_m {
						t.Errorf("m, expected %v got %v", exp_m, m)
					}
				})
			}
	}
	tests := []geom.PointZM{
		{0, 1, 2, 1000.}, {2, 2, 3, 1000.}, {1, 2, 3, 1000.},
	}
	for _, pt := range tests {
		t.Run(fn(pt))
	}
}
