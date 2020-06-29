package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointZMSSetter(t *testing.T) {
	type tcase struct {
		point_srid uint32
		point_xyzm geom.PointZM
		setter     geom.PointZMSSetter
		expected   geom.PointZMSSetter
		err        error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYZMS(tc.point_srid, tc.point_xyzm)
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
			xyzms := tc.setter.XYZMS()
			tc_xyzms := struct {
				Srid uint32
				Xyzm geom.PointZM
			}{tc.point_srid, geom.PointZM{tc.point_xyzm[0], tc.point_xyzm[1], tc.point_xyzm[2], tc.point_xyzm[3]}}
			if !reflect.DeepEqual(tc_xyzms, xyzms) {
				t.Errorf("XYZS, expected %v, got %v", tc_xyzms, xyzms)
			}
		}
	}
	tests := []tcase{
		{
			point_srid: 4326,
			point_xyzm: geom.PointZM{10, 20, 30, 1000},
			setter:     &geom.PointZMS{4326, geom.PointZM{15, 20, 30, 1000}},
			expected:   &geom.PointZMS{4326, geom.PointZM{10, 20, 30, 1000}},
		},
		{
			setter: (*geom.PointZMS)(nil),
			err:    geom.ErrNilPointZMS,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointZMS(t *testing.T) {
	fn := func(pt geom.PointZMS) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xyzm", func(t *testing.T) {
					xyzm := pt.XYZM()
					exp_xyzm := pt.Xyzm
					if xyzm != exp_xyzm {
						t.Errorf("xyzm, expected %v got %v", exp_xyzm, xyzm)
					}
				})
				t.Run("s", func(t *testing.T) {
					s := pt.S()
					exp_s := pt.Srid
					if s != exp_s {
						t.Errorf("srid, expected %v got %v", exp_s, s)
					}
				})
				t.Run("xyzms", func(t *testing.T) {
					xyzms := pt.XYZMS()
					exp_xyzms := pt
					if xyzms != exp_xyzms {
						t.Errorf("xyzms, expected %v got %v", exp_xyzms, xyzms)
					}
				})
			}
	}
	tests := []geom.PointZMS{
		{4326, geom.PointZM{0, 1, 2, 1000}}, {4326, geom.PointZM{2, 2, 3, 1000}}, {4326, geom.PointZM{1, 2, 3, 1000}},
	}
	for _, pt := range tests {
		t.Run(fn(pt))
	}
}
