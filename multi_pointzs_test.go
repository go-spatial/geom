package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointZSSetter(t *testing.T) {
	type tcase struct {
		srid        uint32
		multipointz geom.MultiPointZ
		setter      geom.MultiPointZSSetter
		expected    geom.MultiPointZSSetter
		err         error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multipointz)
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

		ptzs := tc.setter.Points()
		tc_ptzs := struct {
			Srid uint32
			Mpz  geom.MultiPointZ
		}{tc.srid, tc.multipointz}
		if !reflect.DeepEqual(tc_ptzs, ptzs) {
			t.Errorf("PointZSs, expected %v got %v", tc_ptzs, ptzs)
		}
	}
	tests := []tcase{
		{
			srid:        4326,
			multipointz: geom.MultiPointZ{{10, 20, 30}, {30, 40, 50}, {-10, -5, 0}},
			setter:      &geom.MultiPointZS{Srid: 4326, Mpz: geom.MultiPointZ{{15, 20, 30}, {35, 40, 50}, {-15, -5, 0}}},
			expected:    &geom.MultiPointZS{Srid: 4326, Mpz: geom.MultiPointZ{{10, 20, 30}, {30, 40, 50}, {-10, -5, 0}}},
		},
		{
			setter: (*geom.MultiPointZS)(nil),
			err:    geom.ErrNilMultiPointZS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
