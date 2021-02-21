package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonMSSetter(t *testing.T) {
	type tcase struct {
		srid     uint32
		polygonm geom.PolygonM
		lines    [][]geom.LineM
		setter   geom.PolygonMSSetter
		expected geom.PolygonMSSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.srid, tc.polygonm)
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
			polyms := struct {
				Srid uint32
				Polm geom.PolygonM
			}{tc.srid, tc.polygonm}
			glr := tc.setter.LinearRings()
			if !reflect.DeepEqual(polyms, glr) {
				t.Errorf("linear rings, expected %v got %v", polyms, glr)
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
			polygonm: geom.PolygonM{
				{
					{10, 20, 3},
					{30, 40, 5},
					{-10, -5, 0.5},
					{10, 20, 3},
				},
			},
			lines: [][]geom.LineM{
				{
					{
						{10, 20, 3},
						{30, 40, 5},
					},
					{
						{30, 40, 5},
						{-10, -5, 0.5},
					},
					{
						{-10, -5, 0.5},
						{10, 20, 3},
					},
				},
			},
			setter: &geom.PolygonMS{
				Srid: 4326,
				Polm: geom.PolygonM{
					{
						{15, 20, 3},
						{35, 40, 5},
						{-15, -5, 0.5},
						{25, 20, 3},
					},
				},
			},
			expected: &geom.PolygonMS{
				Srid: 4326,
				Polm: geom.PolygonM{
					{
						{10, 20, 3},
						{30, 40, 5},
						{-10, -5, 0.5},
						{10, 20, 3},
					},
				},
			},
		},
		{
			setter: (*geom.PolygonMS)(nil),
			err:    geom.ErrNilPolygonMS,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
