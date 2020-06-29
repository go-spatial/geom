package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestLineStringSSetter(t *testing.T) {
	type tcase struct {
		srid       uint32
		linestring geom.LineString
		setter     geom.LineStringSSetter
		expected   geom.LineStringSSetter
		err        error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.linestring)
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

		lss := tc.setter.Vertices()
		tc_lss := struct {
			Srid uint32
			Ls   geom.LineString
		}{tc.srid, tc.linestring}
		if !reflect.DeepEqual(tc_lss, lss) {
			t.Errorf("Referenced LineString, expected %v got %v", tc_lss, lss)
		}
	}
	tests := []tcase{
		{
			srid:       4326,
			linestring: geom.LineString{{10, 20}, {30, 40}},
			setter:     &geom.LineStringS{Srid: 4326, Ls: geom.LineString{{15, 20}, {35, 40}}},
			expected:   &geom.LineStringS{Srid: 4326, Ls: geom.LineString{{10, 20}, {30, 40}}},
		},
		{
			setter: (*geom.LineStringS)(nil),
			err:    geom.ErrNilLineStringS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
