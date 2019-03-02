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
					{1286956.14225588319823146, 6138803.15957210958003998}, 
					{1286957.51386759686283767, 6138809.63999249972403049}, 
					{1286961.02220776537433267, 6138815.25262837484478951}, 
					{1286966.22873386205174029, 6138819.33963736146688461}, 
					{1286972.51762022217735648, 6138821.39713920280337334}, 
					{1286979.13308080332353711, 6138821.17319339886307716}, 
					{1286985.28200678480789065, 6138818.69579335208982229}, 
					{1286990.19928143476136029, 6138814.27286623604595661}, 
					{1286993.31573253916576505, 6138808.43628553673624992}, 
					{1286994.23947104020044208, 6138801.88588315155357122}, 
					{1286992.86785932653583586, 6138795.40546863991767168}, 
					{1286989.37818054482340813, 6138789.79284578375518322}, 
					{1286984.16232375334948301, 6138785.71984746307134628}, 
					{1286977.86410670098848641, 6138783.66235419642180204}, 
					{1286971.24864611984230578, 6138783.8723024670034647}, 
					{1286965.11838152236305177, 6138786.34969243872910738}, 
					{1286960.18244549166411161, 6138790.7726051164790988}, 
					{1286957.08465576800517738, 6138796.62317034229636192}, 
					{1286956.14225588319823146, 6138803.15957210958003998},
				},
			},
		},
		ClipBox: geom.NewExtent(
			[2]float64{1.2869314293799447e+06, 6.138810017620263e+06},
			[2]float64{1.286970842222649e+06, 6.138849430462967e+06},
		),
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1286957.74991473741829395, 6138810.01762026268988848}, 
					{1286961.02220776537433267, 6138815.25262837484478951}, 
					{1286966.22873386205174029, 6138819.33963736146688461}, 
					{1286970.84222264890559018, 6138820.84900819882750511}, 
					{1286970.84222264890559018, 6138810.01762026268988848}, 
					{1286957.74991473741829395, 6138810.01762026268988848},
				},
			},
		},
	},
}
