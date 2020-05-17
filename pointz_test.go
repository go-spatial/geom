package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"
	"math"

	"github.com/go-spatial/geom"
)

func TestPointZSetter(t *testing.T) {
	type tcase struct {
		point    [3]float64
		setter   geom.PointZSetter
		expected geom.PointZSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYZ(tc.point)
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
			xyz := tc.setter.XYZ()
			if !reflect.DeepEqual(tc.point, xyz) {
				t.Errorf("XYZ, expected %v, got %v", tc.point, xyz)
			}
		}
	}
	tests := []tcase{
		{
			point:    [3]float64{10, 20, 30},
			setter:   &geom.PointZ{15, 20, 30},
			expected: &geom.PointZ{10, 20, 30},
		},
		{
			setter: (*geom.PointZ)(nil),
			err:    geom.ErrNilPointZ,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointZ(t *testing.T) {
	fn := func(pt geom.PointZ) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xy", func(t *testing.T) {
					xy := pt.XY()
					exp_xy := geom.Point{pt[0], pt[1]}
					if xy != exp_xy {
						t.Errorf("xy, expected %v got %v", exp_xy, xy)
					}
				})
				t.Run("xyz", func(t *testing.T) {
                                        xyz := pt.XYZ()
                                        exp_xyz := pt
                                        if xyz != exp_xyz {
                                                t.Errorf("xyz, expected %v got %v", exp_xyz, xyz)
                                        }
                                })
				t.Run("magnitude", func(t *testing.T) {
                                        m := pt.Magnitude()
                                        exp_m := math.Sqrt((pt[0] * pt[0]) + (pt[1] * pt[1]) + (pt[2] * pt[2]))
                                        if m != exp_m {
                                                t.Errorf("magnitude, expected %v got %v", exp_m, m)
                                        }
                                })
			}
	}
	tests := []geom.PointZ{
                {0, 1, 2}, {2, 2, 3}, {1, 2, 3},
        }
	for _, pt := range tests {
		t.Run(fn(pt))
        }
}
