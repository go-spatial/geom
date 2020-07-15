package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointMSSetter(t *testing.T) {
	type tcase struct {
		srid        uint32
		multipointm geom.MultiPointM
		setter      geom.MultiPointMSSetter
		expected    geom.MultiPointMSSetter
		err         error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multipointm)
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
		}

		ptms := tc.setter.Points()
		tc_ptms := struct {
			Srid uint32
			Mpm  geom.MultiPointM
		}{tc.srid, tc.multipointm}
		if !reflect.DeepEqual(tc_ptms, ptms) {
			t.Errorf("PointMSs, expected %v got %v", tc_ptms, ptms)
		}
	}
	tests := []tcase{
		{
			srid:        4326,
			multipointm: geom.MultiPointM{{10, 20, 30}, {30, 40, 50}, {-10, -5, 0}},
			setter:      &geom.MultiPointMS{Srid: 4326, Mpm: geom.MultiPointM{{15, 20, 30}, {35, 40, 50}, {-15, -5, 0}}},
			expected:    &geom.MultiPointMS{Srid: 4326, Mpm: geom.MultiPointM{{10, 20, 30}, {30, 40, 50}, {-10, -5, 0}}},
		},
		{
			setter: (*geom.MultiPointMS)(nil),
			err:    geom.ErrNilMultiPointMS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
