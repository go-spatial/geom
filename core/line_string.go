package core

import "github.com/go-spatial/geom"

// LineString is a basic line type which is made up of two or more points that don't interect.
type LineString []Point

func (ls LineString) SubPoints() []geom.Point {
	pts := make([]geom.Point, 0, len(ls))

	for i := range ls {
		pts = append(pts, ls[i])
	}

	return pts
}
