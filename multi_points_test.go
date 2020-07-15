package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPointSSetter(t *testing.T) {
	type tcase struct {
		srid       uint32
		multipoint geom.MultiPoint
		setter     geom.MultiPointSSetter
		expected   geom.MultiPointSSetter
		err        error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multipoint)
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

		pts := tc.setter.Points()
		tc_pts := struct {
			Srid uint32
			Mp   geom.MultiPoint
		}{tc.srid, tc.multipoint}
		if !reflect.DeepEqual(tc_pts, pts) {
			t.Errorf("PointSs, expected %v got %v", tc_pts, pts)
		}
	}
	tests := []tcase{
		{
			srid:       4326,
			multipoint: geom.MultiPoint{{10, 20}, {30, 40}, {-10, -5}},
			setter:     &geom.MultiPointS{Srid: 4326, Mp: geom.MultiPoint{{15, 20}, {35, 40}, {-15, -5}}},
			expected:   &geom.MultiPointS{Srid: 4326, Mp: geom.MultiPoint{{10, 20}, {30, 40}, {-10, -5}}},
		},
		{
			setter: (*geom.MultiPointS)(nil),
			err:    geom.ErrNilMultiPointS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
