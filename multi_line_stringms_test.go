package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringMSSetter(t *testing.T) {
	type tcase struct {
		srid             uint32
		multilinestringm geom.MultiLineStringM
		setter           geom.MultiLineStringMSSetter
		expected         geom.MultiLineStringMSSetter
		err              error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multilinestringm)
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

		mlsms := tc.setter.MultiLineStringMs()
		tc_mlsms := struct {
			Srid uint32
			Mlsm geom.MultiLineStringM
		}{tc.srid, tc.multilinestringm}
		if !reflect.DeepEqual(tc_mlsms, mlsms) {
			t.Errorf("Referenced MultiLineStringM, expected %v got %v", tc_mlsms, mlsms)
		}
	}
	tests := []tcase{
		{
			srid: 4326,
			multilinestringm: geom.MultiLineStringM{
				{
					{10, 20, 30},
					{30, 40, 50},
				},
				{
					{50, 60, 70},
					{70, 80, 90},
				},
			},
			setter: &geom.MultiLineStringMS{
				Srid: 4326,
				Mlsm: geom.MultiLineStringM{
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
			expected: &geom.MultiLineStringMS{
				Srid: 4326,
				Mlsm: geom.MultiLineStringM{
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
			setter: (*geom.MultiLineStringMS)(nil),
			err:    geom.ErrNilMultiLineStringMS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
