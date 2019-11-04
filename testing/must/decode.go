// Package must provides helpers to decode wkt geometries to be used in tests
package must

import (
	"fmt"

	"github.com/go-spatial/geom"
)

// Decode will panic if err is not nil otherwise return the geometry
func Decode(g geom.Geometry, err error) geom.Geometry {
	if err != nil {
		panic(fmt.Sprintf("got error decoding geometry: %v", err))
	}
	return g
}

// AsPolygon will panic if g is not a geom.Polygon
func AsPolygon(g geom.Geometry) geom.Polygon {
	p, ok := g.(geom.Polygon)
	if !ok {
		panic(fmt.Sprintf("expected polygon, not %t", g))
	}
	return p
}

// AsMultiPolygon will panic if g is not a geom.MultiPolygon or geom.Polygon
// if it is a geom.Polygon, it will return a multipolygon containing just that polygon
func AsMultiPolygon(g geom.Geometry) geom.MultiPolygon {
	switch mp := g.(type) {
	case geom.Polygon:
		return geom.MultiPolygon{mp}
	case geom.MultiPolygon:
		return mp
	default:
		panic(fmt.Sprintf("expected multi-polygon, not %t", g))
	}
}

// AsLines will panic if g can not be coerced into a set of lines
func AsLines(g geom.Geometry) (segs []geom.Line) {
	var err error
	switch geo := g.(type) {
	case geom.LineString:
		segs, err = geo.AsSegments()
		if err != nil {
			panic(err)
		}
	case geom.MultiLineString:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
	case geom.Polygon:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			segs = append(segs, s[i]...)
		}
	case geom.MultiPolygon:
		s, err := geo.AsSegments()
		if err != nil {
			panic(err)
		}
		for i := range s {
			for j := range s[i] {
				segs = append(segs, s[i][j]...)
			}
		}
	default:
		panic("geometry not supported for AsLines")
	}
	return segs
}
