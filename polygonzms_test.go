package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonZMSSetter(t *testing.T) {
	type tcase struct {
		srid      uint32
		polygonzm geom.PolygonZM
		lines     [][]geom.LineZM
		setter    geom.PolygonZMSSetter
		expected  geom.PolygonZMSSetter
		err       error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.srid, tc.polygonzm)
			if tc.err == nil && err != nil {
				t.Errorf("error, expected nil got %v", err)
				return
			}
			if tc.err != nil {
				if err.Error() != tc.err.Error() {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			// compare the results
			if !reflect.DeepEqual(tc.expected, tc.setter) {
				t.Errorf("Polygon Setter, expected %v got %v", tc.expected, tc.setter)
				return
			}

			// compare the results of the Rings
			polyzms := struct {
				Srid  uint32
				Polzm geom.PolygonZM
			}{tc.srid, tc.polygonzm}
			glr := tc.setter.LinearRings()
			if !reflect.DeepEqual(polyzms, glr) {
				t.Errorf("linear rings, expected %v got %v", polyzms, glr)
			}

			// compare the extracted segments
			segs, srid, err := tc.setter.AsSegments()
			if err != nil {
				if !reflect.DeepEqual(tc.lines, segs) {
					t.Errorf("segments, expected %v got %v", tc.lines, segs)
					return
				}
				if srid != tc.srid {
					t.Errorf("srid of segments, expected %v got %v", tc.srid, srid)
				}
			}
		}
	}
	tests := []tcase{
		{
			srid: 4326,
			polygonzm: geom.PolygonZM{
				{
					{10, 20, 30, 3},
					{30, 40, 50, 5},
					{-10, -5, 0, 0.5},
					{10, 20, 30, 3},
				},
			},
			lines: [][]geom.LineZM{
				{
					{
						{10, 20, 30, 3},
						{30, 40, 50, 5},
					},
					{
						{30, 40, 50, 5},
						{-10, -5, 0, 0.5},
					},
					{
						{-10, -5, 0, 0.5},
						{10, 20, 30, 3},
					},
				},
			},
			setter: &geom.PolygonZMS{
				Srid: 4326,
				Polzm: geom.PolygonZM{
					{
						{15, 20, 30, 3},
						{35, 40, 50, 5},
						{-15, -5, 0, 0.5},
						{25, 20, 30, 3},
					},
				},
			},
			expected: &geom.PolygonZMS{
				Srid: 4326,
				Polzm: geom.PolygonZM{
					{
						{10, 20, 30, 3},
						{30, 40, 50, 5},
						{-10, -5, 0, 0.5},
						{10, 20, 30, 3},
					},
				},
			},
		},
		{
			setter: (*geom.PolygonZMS)(nil),
			err:    geom.ErrNilPolygonZMS,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
