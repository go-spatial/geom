package makevalid

import (
	"context"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/makevalid/hitmap"
)

// makevalidCases encapsulates the various parts of the Tegola make valid algorithm
type makevalidCase struct {
	// Description is a simple description of the test
	Description string
	// The MultiPolyon describing the test case
	MultiPolygon *geom.MultiPolygon
	// The expected valid MultiPolygon for the Multipolygon
	ExpectedMultiPolygon *geom.MultiPolygon
}

// Segments returns the flattened segments of the MultiPolygon, on an error it will panic.
func (mvc makevalidCase) Segments() (segments geom.MultiLineString) {
	if debug {
		log.Printf("MakeValidTestCase Polygon: %+v", mvc.MultiPolygon)
	}
	segs, err := Destructure(context.Background(), nil, mvc.MultiPolygon)
	if err != nil {
		panic(err)
	}
	if debug {
		log.Printf("MakeValidTestCase Polygon Segments: %+v", segs)
	}
	return segs
}

func (mvc makevalidCase) Hitmap(clp *geom.Extent) (hm planar.HitMapper) {
	var err error
	if hm, err = hitmap.New(clp, mvc.MultiPolygon); err != nil {
		panic("Hitmap gave error!")
	}
	return hm
}

var makevalidTestCases = [...]makevalidCase{
	{ //  (0) Triangle Test case
		Description:          "Triangle",
		MultiPolygon:         &geom.MultiPolygon{{{{1, 1}, {15, 10}, {10, 20}}}},
		ExpectedMultiPolygon: &geom.MultiPolygon{{{{1, 1}, {15, 10}, {10, 20}}}},
	}, // (0) Triangle Test case
	{ //  (1) Four squire IO_OI
		Description:  "Four Square IO_OI",
		MultiPolygon: &geom.MultiPolygon{{{{1, 4}, {9, 4}, {9, 0}, {5, 0}, {5, 8}, {1, 8}}}},
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{{{1, 4}, {5, 4}, {5, 8}, {1, 8}}},
			{{{5, 0}, {9, 0}, {9, 4}, {5, 4}}},
		},
	}, // (1) four square IO_OI
	{ //  (2) four columns invalid multipolygon
		Description: "Four columns invalid multipolygon",
		MultiPolygon: &geom.MultiPolygon{
			{ // First Polygon
				{{0, 7}, {3, 7}, {3, 3}, {0, 3}}, // Main squire.
				{{1, 5}, {3, 4}, {5, 5}, {3, 7}}, // invalid cutout.
			},
			{{{3, 7}, {6, 8}, {6, 0}, {3, 0}}},
		},
		ExpectedMultiPolygon: &geom.MultiPolygon{{
			{{0, 3}, {3, 3}, {3, 0}, {6, 0}, {6, 8}, {3, 7}, {0, 7}},
			{{1, 5}, {3, 7}, {5, 5}, {3, 4}},
		}},
	}, // (2) Four columns invalid multipolygon
	{ //  (3) Square
		Description:          "Square",
		MultiPolygon:         &geom.MultiPolygon{{{{0, 0}, {4096, 0}, {4096, 4096}, {0, 4096}}}},
		ExpectedMultiPolygon: &geom.MultiPolygon{{{{0, 0}, {4096, 0}, {4096, 4096}, {0, 4096}}}},
	}, // (3) Square
}
