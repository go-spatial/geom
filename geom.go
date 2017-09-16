// Package geom describes geometry interfaces.
package geom

// Geometry is an object with a spatial reference.
type Geometry interface {
	Points() [][2]float64
}

// Pointer is a point with two dimensions.
type Pointer interface {
	Geometry
	XY() [2]float64
}

// PointSetter is a mutable Pointer.
type PointSetter interface {
	Pointer
	SetXY([2]float64) error
}

// MultiPointer is a geometry with multiple points.
type MultiPointer interface {
	Geometry
}

// MultiPointSetter is a mutable MultiPointer.
type MultiPointSetter interface {
	MultiPointer
	SetPoints([][2]float64) error
}

// LineStringer is a line of two or more points.
type LineStringer interface {
	Geometry
}

// LineStringSetter is a mutable LineStringer.
type LineStringSetter interface {
	LineStringer
	SetPoints([][2]float64) error
}

// MultiLineStringer is a geometry with multiple LineStrings.
type MultiLineStringer interface {
	Geometry
	LineStrings() [][][2]float64
}

// MultiLineStringSetter is a mutable MultiLineStringer.
type MultiLineStringSetter interface {
	MultiLineStringer
	SetLineStrings([][][2]float64) error
}

// Polygoner is a geometry consisting of multiple closed LineStrings.
// There must be only one exterior LineString with a clockwise winding order.
// There may be one or more interior LineStrings with a counterclockwise winding orders.
type Polygoner interface {
	Geometry
	LineStrings() [][][2]float64
}

// PolygonSetter is a mutable Polygoner.
type PolygonSetter interface {
	Polygoner
	SetLineStrings([][][2]float64) error
}

// MultiPolygoner is a geometry of multiple polygons.
type MultiPolygoner interface {
	Geometry
	Polygons() [][][][2]float64
}

// MultiPolygonSetter is a mutable MultiPolygoner.
type MultiPolygonSetter interface {
	MultiPolygoner
	SetPolygons([][][][2]float64) error
}

// Collectioner is a collections of different geometries.
type Collectioner interface {
	Geometry
	Geometries() []Geometry
}

// CollectionSetter is a mutable Collectioner.
type CollectionSetter interface {
	Collectioner
	SetGeometries([]Geometry) error
}
