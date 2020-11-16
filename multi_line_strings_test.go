package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringSSetter(t *testing.T) {
	type tcase struct {
		srid            uint32
		multilinestring geom.MultiLineString
		setter          geom.MultiLineStringSSetter
		expected        geom.MultiLineStringSSetter
		err             error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetSRID(tc.srid, tc.multilinestring)
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

		mlss := tc.setter.MultiLineStrings()
		tc_mlss := struct {
			Srid uint32
			Mls  geom.MultiLineString
		}{tc.srid, tc.multilinestring}
		if !reflect.DeepEqual(tc_mlss, mlss) {
			t.Errorf("Referenced MultiLineString, expected %v got %v", tc_mlss, mlss)
		}
	}
	tests := []tcase{
		{
			srid: 4326,
			multilinestring: geom.MultiLineString{
				{
					{10, 20},
					{30, 40},
				},
				{
					{50, 60},
					{70, 80},
				},
			},
			setter: &geom.MultiLineStringS{
				Srid: 4326,
				Mls: geom.MultiLineString{
					{
						{15, 20},
						{35, 40},
					},
					{
						{55, 60},
						{75, 80},
					},
				},
			},
			expected: &geom.MultiLineStringS{
				Srid: 4326,
				Mls: geom.MultiLineString{
					{
						{10, 20},
						{30, 40},
					},
					{
						{50, 60},
						{70, 80},
					},
				},
			},
		},
		{
			setter: (*geom.MultiLineStringS)(nil),
			err:    geom.ErrNilMultiLineStringS,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
