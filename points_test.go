package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointSSetter(t *testing.T) {
	type tcase struct {
		point_srid uint32
		point_xy   geom.Point
		setter     geom.PointSSetter
		expected   geom.PointSSetter
		err        error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYS(tc.point_srid, tc.point_xy)
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
			xys := tc.setter.XYS()
			tc_xys := struct {Srid uint32; Xy geom.Point}{tc.point_srid, geom.Point{tc.point_xy[0], tc.point_xy[1]}}
			if !reflect.DeepEqual(tc_xys, xys) {
				t.Errorf("XYZ, expected %v, got %v", tc_xys, xys)
			}
		}
	}
	tests := []tcase{
		{
			point_srid: 4326,
			point_xy:   geom.Point{10, 20},
			setter:     &geom.PointS{4326, geom.Point{15, 20}},
			expected:   &geom.PointS{4326, geom.Point{10, 20}},
		},
		{
			setter: (*geom.PointS)(nil),
			err:    geom.ErrNilPointS,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointS(t *testing.T) {
	fn := func(pt geom.PointS) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xy", func(t *testing.T) {
					xy := pt.XY()
					exp_xy := pt.Xy
					if xy != exp_xy {
						t.Errorf("xy, expected %v got %v", exp_xy, xy)
					}
				})
                                t.Run("s", func(t *testing.T) {
                                        s := pt.S()
                                        exp_s := pt.Srid
                                        if s != exp_s {
                                                t.Errorf("srid, expected %v got %v", exp_s, s)
                                        }
                                })
				t.Run("xys", func(t *testing.T) {
                                        xys := pt.XYS()
                                        exp_xys := pt
                                        if xys != exp_xys {
                                                t.Errorf("xys, expected %v got %v", exp_xys, xys)
                                        }
                                })
			}
	}
	tests := []geom.PointS{
		{4326, geom.Point{0, 1}}, {4326, geom.Point{2, 2}}, {4326, geom.Point{1, 2}},
        }
	for _, pt := range tests {
		t.Run(fn(pt))
        }
}
