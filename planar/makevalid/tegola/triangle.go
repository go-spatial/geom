package tegola

import (
	"fmt"

	"github.com/go-spatial/geom/cmp"
)

type triangle [3][2]float64

func (t triangle) Center() (pt [2]float64) {
	pt[0] = (t[0][0] + t[1][0] + t[2][0]) / 3
	pt[1] = (t[0][1] + t[1][1] + t[2][1]) / 3
	return pt
}

func (t triangle) ThirdPoint(p1, p2 [2]float64) [2]float64 {
	switch {
	case (cmp.PointEqual(t[0], p1) && cmp.PointEqual(t[1], p2)) ||
		(cmp.PointEqual(t[1], p1) && cmp.PointEqual(t[0], p2)):
		return t[2]
	case (cmp.PointEqual(t[0], p1) && cmp.PointEqual(t[2], p2)) ||
		(cmp.PointEqual(t[2], p1) && cmp.PointEqual(t[0], p2)):
		return t[1]
	default:
		return t[0]
	}
}

func (t triangle) SortedEdges() [3][2][2]float64 {
	return [3][2][2]float64{
		sortedEdge(t[0], t[1]),
		sortedEdge(t[0], t[2]),
		sortedEdge(t[1], t[2]),
	}
}

func newTriangleFromPolygon(py [][][2]float64) triangle {
	// Assume we are getting triangles from the function.
	if debug && len(py) != 1 {
		panic(fmt.Sprintf("Step   3 : assumption invalid for triangle. %v", py))
	}
	if debug && len(py[0]) < 3 {
		panic(fmt.Sprintf("Step   3 : assumption invalid for triangle. %v", py))
	}
	t := triangle{py[0][0], py[0][1], py[0][2]}
	//This is really for our tests. We don't need to to do this, but makes testing easier.
	cmp.RotateToLeftMostPoint(t[:])
	return t
}
