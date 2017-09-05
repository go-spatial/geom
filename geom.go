// Package geom describes geometry interfaces.
package geom

// Geometry describes a geometry.
type Geometry interface{}

// Point is a point with two dimensions.
type Point interface {
	Geometry
	XY() (float64, float64)
}

// Point3 is a point with three dimensions.
type Point3 interface {
	Point
	XYZ() (float64, float64, float64)
}

// MultiPoint is a geometry with multiple points.
type MultiPoint interface {
	Geometry
	Points() []Point
}

// LineString is a line of two or more points
type LineString interface {
	Geometry
	SubPoints() []Point
}

// MultiLineString is a geometry with multiple LineStrings.
type MultiLineString interface {
	Geometry
	LineStrings() []LineString
}

// Polygon is a geometry consisting of multiple closed LineStrings. There must be only one exterior LineString with a clockwise winding order. There may be one or more interior LineStrings with a counterclockwise winding orders.
type Polygon interface {
	Geometry
	SubLineStrings() []LineString
}

// MultiPolygon is a geometry of multiple polygons.
type MultiPolygon interface {
	Geometry
	Polygons() []Polygon
}

// Collection is a collections of different geometries.
type Collection interface {
	Geometry
	Geometries() []Geometry
}
