package geom

/*
 This file describes optional Interfaces to make geometries mutable.
*/

// PointSetter is a mutable Pointer.
type PointSetter interface {
	Pointer
	SetXY([2]float64) error
}

// PointZSetter is a mutable PointZer
type PointZSetter interface {
	PointZer
	SetXYZ([3]float64) error
}

// PointMSetter is a mutable PointMer
type PointMSetter interface {
	PointMer
	SetXYM([3]float64) error
}

// PointZMSetter is a mutable PointZMer
type PointZMSetter interface {
        PointZMer
        SetXYZM([4]float64) error
}

// PointSSetter is a mutable PointSer
type PointSSetter interface {
        PointSer
	SetXYS(srid uint32, xy Point) error
}

// PointZSSetter is a mutable PointZSer
type PointZSSetter interface {
        PointZSer
	SetXYZS(srid uint32, xyz PointZ) error
}

// PointMSSetter is a mutable PointMer
type PointMSSetter interface {
        PointMSer
        SetXYMS(srid uint32, xym PointM) error
}

// PointZMSSetter is a mutable PointZMer
type PointZMSSetter interface {
        PointZMSer
        SetXYZMS(srid uint32, xyzm PointZM) error
}

// MultiPointSetter is a mutable MultiPointer.
type MultiPointSetter interface {
	MultiPointer
	SetPoints([][2]float64) error
}

// LineStringSetter is a mutable LineStringer.
type LineStringSetter interface {
	LineStringer
	SetVertices([][2]float64) error
}

// MultiLineStringSetter is a mutable MultiLineStringer.
type MultiLineStringSetter interface {
	MultiLineStringer
	SetLineStrings([][][2]float64) error
}

// PolygonSetter is a mutable Polygoner.
type PolygonSetter interface {
	Polygoner
	SetLinearRings([][][2]float64) error
}

// MultiPolygonSetter is a mutable MultiPolygoner.
type MultiPolygonSetter interface {
	MultiPolygoner
	SetPolygons([][][][2]float64) error
}

// CollectionSetter is a mutable Collectioner.
type CollectionSetter interface {
	Collectioner
	SetGeometries([]Geometry) error
}
