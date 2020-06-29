package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestLineStringZSSetter(t *testing.T) {
	type tcase struct {
		srid        uint32
		linestringz geom.LineStringZ
		setter      geom.LineStringZSSetter
		expected    geom.LineStringZSSetter
		err         error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.linestringz)
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

		lszs := tc.setter.Vertices()
		tc_lszs := struct {
			Srid uint32
			Lsz  geom.LineStringZ
		}{tc.srid, tc.linestringz}
		if !reflect.DeepEqual(tc_lszs, lszs) {
			t.Errorf("Referenced LineString, expected %v got %v", tc_lszs, lszs)
		}
	}
	tests := []tcase{
		{
			srid:        4326,
			linestringz: geom.LineStringZ{{10, 20, 50}, {30, 40, 90}},
			setter:      &geom.LineStringZS{Srid: 4326, Lsz: geom.LineStringZ{{15, 20, 50}, {35, 40, 90}}},
			expected:    &geom.LineStringZS{Srid: 4326, Lsz: geom.LineStringZ{{10, 20, 50}, {30, 40, 90}}},
		},
		{
			setter: (*geom.LineStringZS)(nil),
			err:    geom.ErrNilLineStringZS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
