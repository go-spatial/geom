package constrained

import "github.com/go-spatial/geom/cmp"

type Point struct {
	Pt            [2]float64
	IsConstrained bool
}

func (pt Point) XY() [2]float64 {
	return pt.Pt
}

func RotatePointsToPos(pts []Point, pos int) {
	lp := len(pts)
	if pos == 0 || pos >= lp {
		return
	}
	is := make([]Point, lp)
	copy(is, pts)
	copy(pts, is[pos:])
	copy(pts[lp-pos:], is[:pos])
}
func RotatePointsToLowestFirst(pts []Point) {
	if len(pts) < 2 {
		return
	}
	var fi int
	for i := 1; i < len(pts); i++ {
		if cmp.PointLess(pts[i].Pt, pts[fi].Pt) {
			fi = i
		}
	}
	RotatePointsToPos(pts, fi)
}
