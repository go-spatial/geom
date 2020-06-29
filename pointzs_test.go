package geom_test

import (
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPointZSSetter(t *testing.T) {
	type tcase struct {
		point_srid uint32
		point_xyz  geom.PointZ
		setter     geom.PointZSSetter
		expected   geom.PointZSSetter
		err        error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetXYZS(tc.point_srid, tc.point_xyz)
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
			xyzs := tc.setter.XYZS()
			tc_xyzs := struct {
				Srid uint32
				Xyz  geom.PointZ
			}{tc.point_srid, geom.PointZ{tc.point_xyz[0], tc.point_xyz[1], tc.point_xyz[2]}}
			if !reflect.DeepEqual(tc_xyzs, xyzs) {
				t.Errorf("XYZS, expected %v, got %v", tc_xyzs, xyzs)
			}
		}
	}
	tests := []tcase{
		{
			point_srid: 4326,
			point_xyz:  geom.PointZ{10, 20, 30},
			setter:     &geom.PointZS{4326, geom.PointZ{15, 20, 30}},
			expected:   &geom.PointZS{4326, geom.PointZ{10, 20, 30}},
		},
		{
			setter: (*geom.PointZS)(nil),
			err:    geom.ErrNilPointZS,
		},
	}

	for i, tc := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestPointZS(t *testing.T) {
	fn := func(pt geom.PointZS) (string, func(*testing.T)) {
		return fmt.Sprintf("%v", pt),
			func(t *testing.T) {
				t.Run("xyz", func(t *testing.T) {
					xyz := pt.XYZ()
					exp_xyz := pt.Xyz
					if xyz != exp_xyz {
						t.Errorf("xyz, expected %v got %v", exp_xyz, xyz)
					}
				})
				t.Run("s", func(t *testing.T) {
					s := pt.S()
					exp_s := pt.Srid
					if s != exp_s {
						t.Errorf("srid, expected %v got %v", exp_s, s)
					}
				})
				t.Run("xyzs", func(t *testing.T) {
					xyzs := pt.XYZS()
					exp_xyzs := pt
					if xyzs != exp_xyzs {
						t.Errorf("xyzs, expected %v got %v", exp_xyzs, xyzs)
					}
				})
			}
	}
	tests := []geom.PointZS{
		{4326, geom.PointZ{0, 1, 2}}, {4326, geom.PointZ{2, 2, 3}}, {4326, geom.PointZ{1, 2, 3}},
	}
	for _, pt := range tests {
		t.Run(fn(pt))
	}
}
