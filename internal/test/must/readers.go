package must

import (
	"fmt"
	"os"

	"github.com/go-spatial/geom/cmp"

	"github.com/go-spatial/geom/encoding/wkt"

	"github.com/go-spatial/geom"
)

func Read(filename string) geom.Geometry {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	geo, err := wkt.Decode(f)
	if err != nil {
		panic(fmt.Sprintf("%v:%v", filename, err))
	}
	return geo
}

func Parse(content []byte) geom.Geometry {
	geo, err := wkt.DecodeBytes(content)
	if err != nil {
		panic(err)
	}
	return geo
}

// ReadLines reads the lines out of the file
// the lines are expected to be in wkt format. It will use the AsLines to
// convert the geometry into lines
func ReadLines(filename string) []geom.Line { return AsLines(Read(filename)) }

// ParseLines decodes the lines in the wkt format. It will use the AsLines to
// convert the geometry into lines
func ParseLines(content []byte) []geom.Line { return AsLines(Parse(content)) }

// AsLines will try and interpet the geom as a set of lines
func AsLines(g geom.Geometry) []geom.Line {
	var (
		err  error
		segs []geom.Line
	)
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

// ReadPoints reads the points out of a file.
// the points are expected to
func ReadPoints(filename string) [][2]float64 { return AsPoints(Read(filename)) }

// ParsePoints decodes points
func ParsePoints(content []byte) [][2]float64 { return AsPoints(Parse(content)) }

// AsPoints
func AsPoints(g geom.Geometry) [][2]float64 {
	var (
		points [][2]float64
	)
	switch geo := g.(type) {
	case geom.Point:
		return [][2]float64{geo}
	case geom.MultiPoint:
		return [][2]float64(geo)
	case [][2]float64:
		return geo
	case geom.LineString:
		return [][2]float64(geo)
	case geom.MultiLineString:
		for i := range geo {
		NEXT:
			for j := range geo[i] {
				for k := range points {
					if cmp.HiCMP.PointEqual(geo[i][j], points[k]) {
						continue NEXT
					}
					points = append(points, geo[i][j])
				}
			}
		}
		return points
	case geom.Polygon:
		for i := range geo {
		NEXT_POLY:
			for j := range geo[i] {
				for k := range points {
					if cmp.HiCMP.PointEqual(geo[i][j], points[k]) {
						continue NEXT_POLY
					}
					points = append(points, geo[i][j])
				}
			}
		}
		return points
	case geom.MultiPolygon:
		for i := range geo {
			for j := range geo[i] {
			NEXT_MULTIPOLY:
				for k := range geo[i][j] {
					for l := range points {
						if cmp.HiCMP.PointEqual(geo[i][j][k], points[l]) {
							continue NEXT_MULTIPOLY
						}
						points = append(points, geo[i][j][k])
					}
				}
			}
		}
		return points
	default:
		panic(fmt.Sprintf("geometry not supported for AsLines: %t", g))
	}
	return nil
}

func ReadMultiPolygon(filename string) geom.MultiPolygon {
	geo := Read(filename)
	mp, ok := geo.(geom.MultiPolygon)
	if !ok {
		panic("expected multipolygon")
	}
	return mp
}

func ParseMultiPolygon(content []byte) geom.MultiPolygon {
	geo := Parse(content)
	mp, ok := geo.(geom.MultiPolygon)
	if !ok {
		panic("expected multipolygon")
	}
	return mp
}

func MPPointer(mp geom.MultiPolygon) *geom.MultiPolygon { return &mp }
