package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringMSetter(t *testing.T) {
	type tcase struct {
		pointms  [][][3]float64
		setter   geom.MultiLineStringMSetter
		expected geom.MultiLineStringMSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLineStringMs(tc.pointms)
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
			mlm := tc.setter.LineStringMs()
			if !reflect.DeepEqual(tc.pointms, mlm) {
				t.Errorf("LineStringMs, expected %v got %v", tc.pointms, mlm)
			}
		}
	}

	tests := []tcase{
		{
			pointms: [][][3]float64{
				{
					{15, 20, 30},
					{35, 40, 50},
				},
				{
					{-15, -5, 0},
					{20, 20, 20},
				},
			},
			setter: &geom.MultiLineStringM{
				{
					{10, 20, 30},
					{30, 40, 50},
				},
				{
					{-10, -5, 0},
					{15, 20, 20},
				},
			},
			expected: &geom.MultiLineStringM{
				{
					{15, 20, 30},
					{35, 40, 50},
				},
				{
					{-15, -5, 0},
					{20, 20, 20},
				},
			},
		},
		{
			setter: (*geom.MultiLineStringM)(nil),
			err:    geom.ErrNilMultiLineStringM,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}

}
