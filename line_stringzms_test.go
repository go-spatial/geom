package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestLineStringZMSSetter(t *testing.T) {
	type tcase struct {
		srid         uint32
		linestringzm geom.LineStringZM
		setter       geom.LineStringZMSSetter
		expected     geom.LineStringZMSSetter
		err          error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.linestringzm)
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

		lszms := tc.setter.Vertices()
		tc_lszms := struct {
			Srid uint32
			Lszm geom.LineStringZM
		}{tc.srid, tc.linestringzm}
		if !reflect.DeepEqual(tc_lszms, lszms) {
			t.Errorf("Referenced LineString, expected %v got %v", tc_lszms, lszms)
		}
	}
	tests := []tcase{
		{
			srid:         4326,
			linestringzm: geom.LineStringZM{{10, 20, 50, 0.5}, {30, 40, 90, 0.9}},
			setter:       &geom.LineStringZMS{Srid: 4326, Lszm: geom.LineStringZM{{15, 20, 50, 0.5}, {35, 40, 90, 0.9}}},
			expected:     &geom.LineStringZMS{Srid: 4326, Lszm: geom.LineStringZM{{10, 20, 50, 0.5}, {30, 40, 90, 0.9}}},
		},
		{
			setter: (*geom.LineStringZMS)(nil),
			err:    geom.ErrNilLineStringZMS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
