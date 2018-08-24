package planar

import (
	"sort"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

type PointsByXY []geom.Point

func (pts PointsByXY) Less(i, j int) bool { return cmp.XYLessPoint(pts[0], pts[1]) }
func (pts PointsByXY) Swap(i, j int)      { pts[i], pts[j] = pts[j], pts[i] }
func (pts PointsByXY) Len() int           { return len(pts) }

func SortUniquePoints(points []geom.Point) []geom.Point {
	sort.Sort(PointsByXY(points))

	// we can use a slice trick to avoid copying the array again. Maybe better
	// than two index variables...
	uniqued := points[:0]
	for i := 0; i < len(points); i++ {
		if i == 0 || !cmp.PointEqual(points[i], points[i-1]) {
			uniqued = append(uniqued, points[i])
		}
	}
	return uniqued
}
