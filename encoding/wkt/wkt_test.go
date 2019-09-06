package wkt

import (
	"testing"
	"errors"

	"github.com/go-spatial/geom"
	gtesting "github.com/go-spatial/geom/testing"
)

func TestEncode(t *testing.T) {
	type tcase struct {
		Geom geom.Geometry
		Rep  string
		Err  error
	}
	fn := func(tc tcase) (string, func(*testing.T)) {
		return tc.Rep, func(t *testing.T) {
			t.Parallel()

			grep, gerr := EncodeString(tc.Geom)
			t.Logf("partial: %v", grep)
			if tc.Err != nil {
				if tc.Err.Error() != gerr.Error() {
					t.Errorf("error, expected %v got %v", tc.Err.Error(), gerr.Error())
				}
				return
			}


			if tc.Err == nil && gerr != nil {
				t.Errorf("error, expected nil got %v", gerr)
				return
			}
			if tc.Rep != grep {
				t.Errorf("representation, expected ‘%v’ got ‘%v’", tc.Rep, grep)
			}

		}
	}
	tests := map[string][]tcase{
		"Point": {
			{
				Err: geom.ErrUnknownGeometry{nil},
			},
			{
				Geom: (*geom.Point)(nil),
				Rep:  "POINT EMPTY",
			},
			{
				Geom: geom.Point{0, 0},
				Rep:  "POINT (0 0)",
			},
			{
				Geom: geom.Point{10, 0},
				Rep:  "POINT (10 0)",
			},
		},
		"MultiPoint": {
			{
				Geom: (*geom.MultiPoint)(nil),
				Rep:  "MULTIPOINT EMPTY",
			},
			{
				Geom: geom.MultiPoint{},
				Rep:  "MULTIPOINT EMPTY",
			},
			{
				Geom: geom.MultiPoint{{0, 0}},
				Rep:  "MULTIPOINT (0 0)",
			},
			{
				Geom: geom.MultiPoint{{0, 0}, {10, 10}},
				Rep:  "MULTIPOINT (0 0,10 10)",
			},
			{
				Geom: geom.MultiPoint{{1, 1}, {3, 3}, {4, 5}},
				Rep:  "MULTIPOINT (1 1,3 3,4 5)",
			},
		},
		"LineString": {
			{
				Geom: (*geom.LineString)(nil),
				Rep:  "LINESTRING EMPTY",
			},
			{
				Geom: geom.LineString{},
				Rep:  "LINESTRING EMPTY",
			},
			{
				Geom: geom.LineString{{0, 0}},
				Err:  errors.New("not enough points for LINESTRING [[0 0]]"),
			},
			{
				Geom: geom.LineString{{10, 10}, {0, 0}},
				Rep:  "LINESTRING (10 10,0 0)",
			},
			{
				Geom: geom.LineString{{10, 10}, {9, 9}, {0, 0}},
				Rep:  "LINESTRING (10 10,9 9,0 0)",
			},
		},
		"MultiLineString": {
			{
				Geom: (*geom.MultiLineString)(nil),
				Rep:  "MULTILINESTRING EMPTY",
			},
			{
				Geom: geom.MultiLineString{},
				Rep:  "MULTILINESTRING EMPTY",
			},
			{
				Geom: geom.MultiLineString{{}},
				Rep:  "MULTILINESTRING EMPTY",
			},
			{
				Geom: geom.MultiLineString{{{10, 10}}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{{10, 10}, {11, 11}}},
				Rep:  "MULTILINESTRING ((10 10,11 11))",
			},
			{
				Geom: geom.MultiLineString{{}, {}},
				Rep:  "MULTILINESTRING EMPTY",
			},
			{
				Geom: geom.MultiLineString{{}, {{10, 10}}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{}, {{10, 10}, {20, 20}}},
				Rep:  "MULTILINESTRING ((10 10,20 20))",
			},
			{
				Geom: geom.MultiLineString{{{10, 10}}, {}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{{10, 10}}, {{10, 10}}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{{10, 10}}, {{10, 10}, {20, 20}}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{{10, 10}, {20, 20}}, {}},
				Rep:  "MULTILINESTRING ((10 10,20 20))",
			},
			{
				Geom: geom.MultiLineString{{{10, 10}, {20, 20}}, {{10, 10}}},
				Err: errors.New("not enough points for LINESTRING [[10 10]]"),
			},
			{
				Geom: geom.MultiLineString{{{10, 10}, {20, 20}}, {{10, 10}, {20, 20}}},
				Rep:  "MULTILINESTRING ((10 10,20 20),(10 10,20 20))",
			},
		},
		"Polygon": {
			{
				Geom: (*geom.Polygon)(nil),
				Rep:  "POLYGON EMPTY",
			},
			{
				Geom: geom.Polygon{},
				Rep:  "POLYGON EMPTY",
			},
			{
				Geom: geom.Polygon{{}},
				Rep:  "POLYGON EMPTY",
			},
			{
				Geom: geom.Polygon{{}, {}},
				Rep:  "POLYGON EMPTY",
			},
			{
				Geom: geom.Polygon{{{10, 10}, {11, 11}, {12, 12}}, {}},
				Rep:  "POLYGON ((10 10,11 11,12 12,10 10))",
			},
			{
				Geom: geom.Polygon{{{10, 10}, {11, 11}, {12, 12}}, {{20, 20}, {21, 21}, {22, 22}}},
				Rep:  "POLYGON ((10 10,11 11,12 12,10 10),(20 20,21 21,22 22,20 20))",
			},
			{
				Geom: geom.Polygon{{}, {{10, 10}, {11, 11}, {12, 12}}},
				Rep:  "POLYGON ((10 10,11 11,12 12,10 10))",
			},
		},
		"MultiPolygon": {
			{
				Geom: (*geom.MultiPolygon)(nil),
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{{}}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{}, {}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{{}}, {}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{}, {{}}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{{}}, {{}}},
				Rep:  "MULTIPOLYGON EMPTY",
			},
			{
				Geom: &geom.MultiPolygon{{{{10, 10}, {11, 11}, {12, 12}}}},
				Rep:  "MULTIPOLYGON (((10 10,11 11,12 12,10 10)))",
			},
		},
		"Collectioner": {
			{
				Geom: (*geom.Collection)(nil),
				Rep:  "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{},
				Rep:  "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.Point)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.MultiPoint)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.LineString)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.MultiLineString)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.Polygon)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					(*geom.MultiPolygon)(nil),
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					geom.MultiPoint{},
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					geom.LineString{},
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					geom.MultiLineString{},
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					geom.Polygon{},
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					&geom.MultiPolygon{},
				},
				Rep: "GEOMETRYCOLLECTION EMPTY",
			},
			{
				Geom: geom.Collection{
					geom.Point{10, 10},
				},
				Rep: "GEOMETRYCOLLECTION (POINT (10 10))",
			},
			{
				Geom: geom.Collection{
					geom.Point{10, 10},
					geom.LineString{{11, 11}, {22, 22}},
				},
				Rep: "GEOMETRYCOLLECTION (POINT (10 10),LINESTRING (11 11,22 22))",
			},
		},
	}
	for name, subtests := range tests {
		t.Run(name, func(t *testing.T) {
			for _, tc := range subtests {
				t.Run(fn(tc))
			}
		})
	}
}

func BenchmarkEncodeSin100(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeBytes(gtesting.SinLineString(1.0, 0.0, 100.0, 100))
	}
}

func BenchmarkEncodeSin1000(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeBytes(gtesting.SinLineString(1.0, 0.0, 100.0, 1000))
	}
}

func BenchmarkEncodeTile(b *testing.B) {
	for n := 0; n < b.N; n++ {
		EncodeBytes(gtesting.Tiles[0])
	}
}
