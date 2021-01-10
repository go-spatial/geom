package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringZMSSetter(t *testing.T) {
	type tcase struct {
		srid              uint32
		multilinestringzm geom.MultiLineStringZM
		setter            geom.MultiLineStringZMSSetter
		expected          geom.MultiLineStringZMSSetter
		err               error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetSRID(tc.srid, tc.multilinestringzm)
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

			mlszms := tc.setter.MultiLineStringZMs()
			tc_mlszms := struct {
				Srid  uint32
				Mlszm geom.MultiLineStringZM
			}{tc.srid, tc.multilinestringzm}
			if !reflect.DeepEqual(tc_mlszms, mlszms) {
				t.Errorf("Referenced MultiLineStringZM, expected %v got %v", tc_mlszms, mlszms)
			}
		}
	}

	tests := []tcase{
		{
			srid: 4326,
			multilinestringzm: geom.MultiLineStringZM{
				{
					{10, 20, 30, 5},
					{30, 40, 50, 5},
				},
				{
					{50, 60, 70, 5},
					{70, 80, 90, 5},
				},
			},
			setter: &geom.MultiLineStringZMS{
				Srid: 4326,
				Mlszm: geom.MultiLineStringZM{
					{
						{15, 20, 30, 5},
						{35, 40, 50, 5},
					},
					{
						{55, 60, 70, 5},
						{75, 80, 90, 5},
					},
				},
			},
			expected: &geom.MultiLineStringZMS{
				Srid: 4326,
				Mlszm: geom.MultiLineStringZM{
					{
						{10, 20, 30, 5},
						{30, 40, 50, 5},
					},
					{
						{50, 60, 70, 5},
						{70, 80, 90, 5},
					},
				},
			},
		},
		{
			setter: (*geom.MultiLineStringZMS)(nil),
			err:    geom.ErrNilMultiLineStringZMS,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
