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
	// ClipBox the extent to use to clip the geometry before making valid
	ClipBox *geom.Extent
}

// Segments returns the flattened segments of the MultiPolygon, on an error it will panic.
func (mvc makevalidCase) Segments() (segments []geom.Line) {
	if debug {
		log.Printf("MakeValidTestCase Polygon: %+v", mvc.MultiPolygon)
	}
	segs, err := Destructure(context.Background(), mvc.ClipBox, mvc.MultiPolygon)
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

	{ // (4) Henry Circle one
		Description: "circle one",
		MultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1286956.1422558832, 6138803.15957211},
					{1286957.5138675969, 6138809.6399925},
					{1286961.0222077654, 6138815.252628375},
					{1286966.228733862, 6138819.3396373615},
					{1286972.5176202222, 6138821.397139203},
					{1286979.1330808033, 6138821.173193399},
					{1286985.2820067848, 6138818.695793352},
					{1286990.1992814348, 6138814.272866236},
					{1286993.3157325392, 6138808.436285537},
					{1286994.2394710402, 6138801.885883152},
					{1286992.8678593265, 6138795.40546864},
					{1286989.3781805448, 6138789.792845784},
					{1286984.1623237533, 6138785.719847463},
					{1286977.864106701, 6138783.662354196},
					{1286971.2486461198, 6138783.872302467},
					{1286965.1183815224, 6138786.349692439},
					{1286960.1824454917, 6138790.7726051165},
					{1286957.084655768, 6138796.623170342},
				},
			},
		},
		ClipBox: geom.NewExtent(
			[2]float64{1286940.46060967, 6138830.2432236},
			[2]float64{1286969.19030943, 6138807.58852643},
		),
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1.28695708e+06 ,6.138807589e+06},
					{1.28696919e+06 ,6.138807589e+06},
					{1.28696919e+06 ,6.138820309e+06},
					{1.286966229e+06,6.13881934e+06 },
					{1.286961022e+06,6.138815252e+06},
					{1.286957514e+06,6.13880964e+06 },
				},
			},
		},
	},
}
