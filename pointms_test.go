package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointMSSetter(t *testing.T) {
	type tcase struct {
		point_srid uint32
		point_xym  geom.PointM
		setter     geom.PointMSSetter
		expected   geom.PointMSSetter
		err        error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYMS(tc.point_srid, tc.point_xym)
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
			xyms := tc.setter.XYMS()
			tc_xyms := struct {
				Srid uint32
				Xym  geom.PointM
			}{tc.point_srid, geom.PointM{tc.point_xym[0], tc.point_xym[1], tc.point_xym[2]}}
			if !reflect.DeepEqual(tc_xyms, xyms) {
				t.Errorf("XYZS, expected %v, got %v", tc_xyms, xyms)
			}
		}
	}
	tests := []tcase{
		{
			point_srid: 4326,
			point_xym:  geom.PointM{10, 20, 1000},
			setter:     &geom.PointMS{4326, geom.PointM{15, 20, 1000}},
			expected:   &geom.PointMS{4326, geom.PointM{10, 20, 1000}},
		},
		{
			setter: (*geom.PointMS)(nil),
			err:    geom.ErrNilPointMS,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointMS(t *testing.T) {
	fn := func(pt geom.PointMS) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xym", func(t *testing.T) {
					xym := pt.XYM()
					exp_xym := pt.Xym
					if xym != exp_xym {
						t.Errorf("xym, expected %v got %v", exp_xym, xym)
					}
				})
				t.Run("s", func(t *testing.T) {
					s := pt.S()
					exp_s := pt.Srid
					if s != exp_s {
						t.Errorf("srid, expected %v got %v", exp_s, s)
					}
				})
				t.Run("xyms", func(t *testing.T) {
					xyms := pt.XYMS()
					exp_xyms := pt
					if xyms != exp_xyms {
						t.Errorf("xyms, expected %v got %v", exp_xyms, xyms)
					}
				})
			}
	}
	tests := []geom.PointMS{
		{4326, geom.PointM{0, 1, 1000}}, {4326, geom.PointM{2, 2, 300}}, {4326, geom.PointM{1, 2, 1000}},
	}
	for _, pt := range tests {
		t.Run(fn(pt))
	}
}
