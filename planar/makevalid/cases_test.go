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

func (mvc makevalidCase) Hitmap(ctx context.Context, clp *geom.Extent) (hm planar.HitMapper) {
	var err error
	if hm, err = hitmap.New(ctx, clp, mvc.MultiPolygon); err != nil {
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

	{ // (4) gdey Circle one, extent not based on tile
		Description: "circle one gdey",
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
					{1286956.1422558832, 6138803.15957211},
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
					{1.28695708e+06, 6.138807589e+06},
					{1.28696919e+06, 6.138807589e+06},
					{1.28696919e+06, 6.138820309e+06},
					{1.286966229e+06, 6.13881934e+06},
					{1.286961022e+06, 6.138815252e+06},
					{1.286957514e+06, 6.13880964e+06},
				},
			},
		},
	},
	{ // (5) Henry Circle one
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
					{1.28695775e+06, 6.138810018e+06},
					{1.286970842e+06, 6.138810018e+06},
					{1.286970842e+06, 6.138820849e+06},
					{1.286966229e+06, 6.13881934e+06},
					{1.286961022e+06, 6.138815252e+06},
				},
			},
		},
	},
	{ // (6) Henry Circle one with modified bbox
		Description: "circle one right",
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
			[2]float64{1.2869696478940817e+06, 6.138810017620263e+06},
			[2]float64{1.2870090607367859e+06, 6.138849430462967e+06},
		),
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1286969.64789408165961504, 6138820.45826600957661867},
					{1286972.51762022217735648, 6138821.39713920280337334},
					{1286979.13308080332353711, 6138821.17319339886307716},
					{1286985.28200678480789065, 6138818.69579335208982229},
					{1286990.19928143476136029, 6138814.27286623604595661},
					{1286992.47137646726332605, 6138810.01762026268988848},
					{1286969.64789408165961504, 6138810.01762026268988848},
					{1286969.64789408165961504, 6138820.45826600957661867},
				},
			},
		},
	},
	{ // (7) irregular polygon middle
		Description: "irregular polygon middle",
		MultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1285733.82161285122856498, 6138599.27377647440880537},
					{1286000.50211894721724093, 6138741.57482850179076195},
					{1286141.80611756001599133, 6138795.4894480612128973},
					{1286188.3289475291967392, 6138833.67217746842652559},
					{1286338.14253718149848282, 6138994.84492336492985487},
					{1286355.31101033766753972, 6139005.97246335539966822},
					{1286378.02191449701786041, 6139012.45303734578192234},
					{1286407.85213660122826695, 6139015.33640445396304131},
					{1286447.46092385076917708, 6139014.0906777661293745},
					{1286475.49032241315580904, 6139010.32550495583564043},
					{1286502.81991908047348261, 6139098.50656714290380478},
					{1286576.41108634602278471, 6139195.95469637773931026},
					{1286588.79291453957557678, 6139188.004275968298316},
					{1286609.48838924197480083, 6139174.42699059750884771},
					{1286659.51022868789732456, 6139245.65898859594017267},
					{1286840.97352537396363914, 6139262.73571997787803411},
					{1287383.30133250961080194, 6139312.66435879096388817},
					{1287393.22918872558511794, 6139313.91012894175946712},
					{1287415.46422759653069079, 6139056.13755289185792208},
					{1287421.97705056634731591, 6138965.68943751137703657},
					{1287442.80315495887771249, 6138703.88238476030528545},
					{1287409.34329369082115591, 6138702.07684669084846973},
					{1287413.98064757976680994, 6138643.04016063269227743},
					{1287088.60075854370370507, 6138637.45563034061342478},
					{1287009.04727913532406092, 6138643.29209441691637039},
					{1286958.56823578476905823, 6138655.16098329238593578},
					{1286890.50083814444951713, 6138655.13299060706049204},
					{1286770.43349436926655471, 6138667.09987043403089046},
					{1286233.20024502673186362, 6138660.12968576420098543},
					{1286139.74403464025817811, 6138613.11607844196259975},
					{1285975.57051010569557548, 6138567.18043652083724737},
					{1285753.22012137877754867, 6138551.35075232852250338},
					{1285733.82161285122856498, 6138599.27377647440880537},
				},
			},
		},
		ClipBox: geom.NewExtent(
			[2]float64{1286578.50528845749795437, 6138801.06015601102262735},
			[2]float64{1287209.11077172216027975, 6139431.66563927568495274},
		),
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1286578.50528845749795437, 6139194.61000099405646324},
					{1286588.79291453957557678, 6139188.004275968298316},
					{1286609.48838924197480083, 6139174.42699059750884771},
					{1286659.51022868789732456, 6139245.65898859594017267},
					{1286840.97352537396363914, 6139262.73571997787803411},
					{1287209.11077172216027975, 6139296.62775237020105124},
					{1287209.11077172216027975, 6138801.06015601102262735},
					{1286578.50528845749795437, 6138801.06015601102262735},
					{1286578.50528845749795437, 6139194.61000099405646324},
				},
			},
		},
	},
	{ // (8) irregular polygon right
		Description: "irregular polygon right",
		MultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1285733.82161285122856498, 6138599.27377647440880537},
					{1286000.50211894721724093, 6138741.57482850179076195},
					{1286141.80611756001599133, 6138795.4894480612128973},
					{1286188.3289475291967392, 6138833.67217746842652559},
					{1286338.14253718149848282, 6138994.84492336492985487},
					{1286355.31101033766753972, 6139005.97246335539966822},
					{1286378.02191449701786041, 6139012.45303734578192234},
					{1286407.85213660122826695, 6139015.33640445396304131},
					{1286447.46092385076917708, 6139014.0906777661293745},
					{1286475.49032241315580904, 6139010.32550495583564043},
					{1286502.81991908047348261, 6139098.50656714290380478},
					{1286576.41108634602278471, 6139195.95469637773931026},
					{1286588.79291453957557678, 6139188.004275968298316},
					{1286609.48838924197480083, 6139174.42699059750884771},
					{1286659.51022868789732456, 6139245.65898859594017267},
					{1286840.97352537396363914, 6139262.73571997787803411},
					{1287383.30133250961080194, 6139312.66435879096388817},
					{1287393.22918872558511794, 6139313.91012894175946712},
					{1287415.46422759653069079, 6139056.13755289185792208},
					{1287421.97705056634731591, 6138965.68943751137703657},
					{1287442.80315495887771249, 6138703.88238476030528545},
					{1287409.34329369082115591, 6138702.07684669084846973},
					{1287413.98064757976680994, 6138643.04016063269227743},
					{1287088.60075854370370507, 6138637.45563034061342478},
					{1287009.04727913532406092, 6138643.29209441691637039},
					{1286958.56823578476905823, 6138655.16098329238593578},
					{1286890.50083814444951713, 6138655.13299060706049204},
					{1286770.43349436926655471, 6138667.09987043403089046},
					{1286233.20024502673186362, 6138660.12968576420098543},
					{1286139.74403464025817811, 6138613.11607844196259975},
					{1285975.57051010569557548, 6138567.18043652083724737},
					{1285753.22012137877754867, 6138551.35075232852250338},
					{1285733.82161285122856498, 6138599.27377647440880537},
				},
			},
		},
		ClipBox: geom.NewExtent(
			[2]float64{1287190.001514652, 6138801.06015601},
			[2]float64{1287820.606997917, 6139431.66563927},
		),
		ExpectedMultiPolygon: &geom.MultiPolygon{
			{ // Polygon
				{ // Ring
					{1287190.00151465274393559, 6139294.86848577577620745},
					{1287383.30133250961080194, 6139312.66435879096388817},
					{1287393.22918872558511794, 6139313.91012894175946712},
					{1287415.46422759653069079, 6139056.13755289185792208},
					{1287421.97705056634731591, 6138965.68943751137703657},
					{1287435.07290328131057322, 6138801.06015601102262735},
					{1287190.00151465274393559, 6138801.06015601102262735},
					{1287190.00151465274393559, 6139294.86848577577620745},
				},
			},
		},
	},
}
