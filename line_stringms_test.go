package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestLineStringMSSetter(t *testing.T) {
	type tcase struct {
		srid        uint32
		linestringm geom.LineStringM
		setter      geom.LineStringMSSetter
		expected    geom.LineStringMSSetter
		err         error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.linestringm)
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

		lsms := tc.setter.Vertices()
		tc_lsms := struct {
			Srid uint32
			Lsm  geom.LineStringM
		}{tc.srid, tc.linestringm}
		if !reflect.DeepEqual(tc_lsms, lsms) {
			t.Errorf("Referenced LineString, expected %v got %v", tc_lsms, lsms)
		}
	}
	tests := []tcase{
		{
			srid:        4326,
			linestringm: geom.LineStringM{{10, 20, 0.5}, {30, 40, 0.9}},
			setter:      &geom.LineStringMS{Srid: 4326, Lsm: geom.LineStringM{{15, 20, 0.5}, {35, 40, 0.9}}},
			expected:    &geom.LineStringMS{Srid: 4326, Lsm: geom.LineStringM{{10, 20, 0.5}, {30, 40, 0.9}}},
		},
		{
			setter: (*geom.LineStringMS)(nil),
			err:    geom.ErrNilLineStringMS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
