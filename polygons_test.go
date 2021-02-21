package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonSSetter(t *testing.T) {
	type tcase struct {
		srid     uint32
		lines    [][]geom.Line
		polygon  geom.Polygon
		setter   geom.PolygonSSetter
		expected geom.PolygonSSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.srid, tc.polygon)
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
			polys := struct {
				Srid uint32
				Pol  geom.Polygon
			}{tc.srid, tc.polygon}
			glr := tc.setter.LinearRings()
			if !reflect.DeepEqual(polys, glr) {
				t.Errorf("linear rings, expected %v got %v", polys, glr)
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
			polygon: geom.Polygon{
				{
					{10, 20},
					{30, 40},
					{-10, -5},
					{10, 20},
				},
			},
			lines: [][]geom.Line{
				{
					{
						{10, 20},
						{30, 40},
					},
					{
						{30, 40},
						{-10, -5},
					},
					{
						{-10, -5},
						{10, 20},
					},
				},
			},
			setter: &geom.PolygonS{
				Srid: 4326,
				Pol: geom.Polygon{
					{
						{15, 20},
						{35, 40},
						{-15, -5},
						{25, 20},
					},
				},
			},
			expected: &geom.PolygonS{
				Srid: 4326,
				Pol: geom.Polygon{
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
			setter: (*geom.PolygonS)(nil),
			err:    geom.ErrNilPolygonS,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
