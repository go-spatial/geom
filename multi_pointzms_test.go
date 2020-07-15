package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointZMSSetter(t *testing.T) {
	type tcase struct {
		srid         uint32
		multipointzm geom.MultiPointZM
		setter       geom.MultiPointZMSSetter
		expected     geom.MultiPointZMSSetter
		err          error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multipointzm)
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

		ptzms := tc.setter.Points()
		tc_ptzms := struct {
			Srid uint32
			Mpzm geom.MultiPointZM
		}{tc.srid, tc.multipointzm}
		if !reflect.DeepEqual(tc_ptzms, ptzms) {
			t.Errorf("PointZMSs, expected %v got %v", tc_ptzms, ptzms)
		}
	}
	tests := []tcase{
		{
			srid:         4326,
			multipointzm: geom.MultiPointZM{{10, 20, 30, 40}, {30, 40, 50, 60}, {-10, -5, 0, 5}},
			setter:       &geom.MultiPointZMS{Srid: 4326, Mpzm: geom.MultiPointZM{{15, 20, 30, 40}, {35, 40, 50, 60}, {-15, -5, 0, 5}}},
			expected:     &geom.MultiPointZMS{Srid: 4326, Mpzm: geom.MultiPointZM{{10, 20, 30, 40}, {30, 40, 50, 60}, {-10, -5, 0, 5}}},
		},
		{
			setter: (*geom.MultiPointZMS)(nil),
			err:    geom.ErrNilMultiPointZMS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
