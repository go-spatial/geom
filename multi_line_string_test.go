package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestMultiLineStringSetter(t *testing.T) {
	type tcase struct {
		points   [][][2]float64
		setter   geom.MultiLineStringSetter
		expected geom.MultiLineStringSetter
		err      error
	}
	fn := func(t *testing.T, tc tcase) {
		err := tc.setter.SetLineStrings(tc.points)
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
		ml := tc.setter.LineStrings()
		if !reflect.DeepEqual(tc.points, ml) {
			t.Errorf("LineStrings, expected %v got %v", tc.points, ml)
		}
	}
	tests := []tcase{
		{
			points: [][][2]float64{
				{
					{15, 20},
					{35, 40},
				},
				{
					{-15, -5},
					{20, 20},
				},
			},
			setter: &geom.MultiLineString{
				{
					{10, 20},
					{30, 40},
				},
				{
					{-10, -5},
					{15, 20},
				},
			},
			expected: &geom.MultiLineString{
				{
					{15, 20},
					{35, 40},
				},
				{
					{-15, -5},
					{20, 20},
				},
			},
		},
		{
			setter: (*geom.MultiLineString)(nil),
			err:    geom.ErrNilMultiLineString,
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}

}
