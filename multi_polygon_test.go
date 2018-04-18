package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiPolygonSetter(t *testing.T) {
	type tcase struct {
		points   [][][][2]float64
		setter   geom.MultiPolygonSetter
		expected geom.MultiPolygonSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetPolygons(tc.points)
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
		mp := tc.setter.Polygons()
		if !reflect.DeepEqual(tc.points, mp) {
			t.Errorf("Polygons, expected %v got %v", tc.points, mp)
		}
	}
	tests := []tcase{
		{
			points: [][][][2]float64{
				{
					{
						{10, 20},
						{30, 40},
						{-10, -5},
						{10, 20},
					},
				},
			},
			setter: &geom.MultiPolygon{
				{
					{
						{15, 20},
						{30, 45},
						{-15, -5},
						{10, 25},
					},
				},
			},
			expected: &geom.MultiPolygon{
				{
					{
						{10, 20},
						{30, 40},
						{-10, -5},
						{10, 20},
					},
				},
			},
		},
		{
			setter: (*geom.MultiPolygon)(nil),
			err:    geom.ErrNilMultiPolygon,
		},
	}

	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
