package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringZSetter(t *testing.T) {
	type tcase struct {
		pointzs  [][][3]float64
		setter   geom.MultiLineStringZSetter
		expected geom.MultiLineStringZSetter
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLineStringZs(tc.pointzs)
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
			mlz := tc.setter.LineStringZs()
			if !reflect.DeepEqual(tc.pointzs, mlz) {
				t.Errorf("LineStringZs, expected %v got %v", tc.pointzs, mlz)
			}
		}
	}

	tests := []tcase{
		{
			pointzs: [][][3]float64{
				{
					{15, 20, 30},
					{35, 40, 50},
				},
				{
					{-15, -5, 0},
					{20, 20, 20},
				},
			},
			setter: &geom.MultiLineStringZ{
				{
					{10, 20, 30},
					{30, 40, 50},
				},
				{
					{-10, -5, 0},
					{15, 20, 20},
				},
			},
			expected: &geom.MultiLineStringZ{
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
			setter: (*geom.MultiLineStringZ)(nil),
			err:    geom.ErrNilMultiLineStringZ,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}

}
