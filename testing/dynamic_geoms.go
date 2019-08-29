package testing

import (
	"math"

	"github.com/go-spatial/geom"
)

// BoxPolygon returns a polygon that is a box with side lengths of
// dim, a clockwise winding order, and points at (0, 0) and (dim, dim).
func BoxPolygon(dim float64) geom.Polygon {
	return geom.Polygon{{{0, 0}, {dim, 0}, {dim, dim}, {0, dim}}}
}

func SelfIntBoxLineString(dim float64) geom.LineString {
	return geom.LineString{{0, 0}, {dim, dim}, {dim, 0}, {0, dim}}
}

// SinLineString returns a line string that is a sin wave with
// the given amplitude and domain (x values) of the set [start, end].
// points is the number of points in the line and must be >= 2, or the
// function panics.
func SinLineString(amp, start, end float64, points int) geom.LineString {
	if points < 2 {
		panic("cannot have a line with less than 2 points")
	}

	res := (end - start) / (float64(points) - 1)
	ret := make([][2]float64, points)
	x := 0.0

	// the last point should be end
	var i int
	for i = 0; i < points - 1; i++ {
		ret[i] = [2]float64{x, math.Sin(x)}
		x += res
	}

	ret[i] = [2]float64{end, math.Sin(end)}

	return ret
}


