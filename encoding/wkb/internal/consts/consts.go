package consts

// geometry types
// http://edndoc.esri.com/arcsde/9.1/general_topics/wkb_representation.htm
const (
	Point           uint32 = 1
	LineString      uint32 = 2
	Polygon         uint32 = 3
	MultiPoint      uint32 = 4
	MultiLineString uint32 = 5
	MultiPolygon    uint32 = 6
	Collection      uint32 = 7
)

// Extended Types
// https://portal.ogc.org/files/?artifact_id=111658&version=1#OGCSFACA
const (
	WKBZ    = 0x80_000_000
	WKBM    = 0x40_000_000
	WKBSRID = 0x20_000_000
)

const (
	WKBPointZ           = Point | WKBZ
	WKBLineStringZ      = LineString | WKBZ
	WKBPolygonZ         = Polygon | WKBZ
	WKBMultiPointZ      = MultiPoint | WKBZ
	WKBMultiLineStringZ = MultiLineString | WKBZ
	WKBMultiPolygonZ    = MultiPolygon | WKBZ
	WKBCollectionZ      = Collection | WKBZ
)
