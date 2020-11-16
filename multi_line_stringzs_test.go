package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringZSSetter(t *testing.T) {
	type tcase struct {
		srid             uint32
		multilinestringz geom.MultiLineStringZ
		setter           geom.MultiLineStringZSSetter
		expected         geom.MultiLineStringZSSetter
		err              error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multilinestringz)
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

		mlszs := tc.setter.MultiLineStringZs()
		tc_mlszs := struct {
			Srid uint32
			Mlsz geom.MultiLineStringZ
		}{tc.srid, tc.multilinestringz}
		if !reflect.DeepEqual(tc_mlszs, mlszs) {
			t.Errorf("Referenced MultiLineStringZ, expected %v got %v", tc_mlszs, mlszs)
		}
	}
	tests := []tcase{
		{
			srid: 4326,
			multilinestringz: geom.MultiLineStringZ{
				{
					{10, 20, 30},
					{30, 40, 50},
				},
				{
					{50, 60, 70},
					{70, 80, 90},
				},
			},
			setter: &geom.MultiLineStringZS{
				Srid: 4326,
				Mlsz: geom.MultiLineStringZ{
					{
						{15, 20, 30},
						{35, 40, 50},
					},
					{
						{55, 60, 70},
						{75, 80, 90},
					},
				},
			},
			expected: &geom.MultiLineStringZS{
				Srid: 4326,
				Mlsz: geom.MultiLineStringZ{
					{
						{10, 20, 30},
						{30, 40, 50},
					},
					{
						{50, 60, 70},
						{70, 80, 90},
					},
				},
			},
		},
		{
			setter: (*geom.MultiLineStringZS)(nil),
			err:    geom.ErrNilMultiLineStringZS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
