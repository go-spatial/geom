package geom_test

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func TestPolygonZSSetter(t *testing.T) {
	type tcase struct {
		srid     uint32
		polygonz geom.PolygonZ
		lines    [][]geom.LineZ
		setter   geom.PolygonZSSetter
		expected geom.PolygonZSSetter
		err      error
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			err := tc.setter.SetLinearRings(tc.srid, tc.polygonz)
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
			polyzs := struct {
				Srid uint32
				Polz geom.PolygonZ
			}{tc.srid, tc.polygonz}
			glr := tc.setter.LinearRings()
			if !reflect.DeepEqual(polyzs, glr) {
				t.Errorf("linear rings, expected %v got %v", polyzs, glr)
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
			polygonz: geom.PolygonZ{
				{
					{10, 20, 30},
					{30, 40, 50},
					{-10, -5, 0},
					{10, 20, 30},
				},
			},
			lines: [][]geom.LineZ{
				{
					{
						{10, 20, 30},
						{30, 40, 50},
					},
					{
						{30, 40, 50},
						{-10, -5, 0},
					},
					{
						{-10, -5, 0},
						{10, 20, 30},
					},
				},
			},
			setter: &geom.PolygonZS{
				Srid: 4326,
				Polz: geom.PolygonZ{
					{
						{15, 20, 30},
						{35, 40, 50},
						{-15, -5, 0},
						{25, 20, 30},
					},
				},
			},
			expected: &geom.PolygonZS{
				Srid: 4326,
				Polz: geom.PolygonZ{
					{
						{10, 20, 30},
						{30, 40, 50},
						{-10, -5, 0},
						{10, 20, 30},
					},
				},
			},
		},
		{
			setter: (*geom.PolygonZS)(nil),
			err:    geom.ErrNilPolygonZS,
		},
	}

	for i := range tests {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tests[i]))
	}
}
