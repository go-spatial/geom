package qetriangulate

import (
	"context"
	"sort"

	"github.com/go-spatial/geom/planar/triangulate/geometry"
	"github.com/go-spatial/geom/planar/triangulate/subdivision"
	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

type Triangulator struct {
	points [][2]float64
}

func New(pts ...[2]float64) *Triangulator {

	return &Triangulator{
		points: pts,
	}
}

/*
func (t *Triangulator) initSubdivision() {
	sort.Sort(cmp.ByXY(t.points))
	tri := geometry.TriangleContaining(t.points...)
	t.sd = subdivision.New(tri[0], tri[1], tri[2])
	var oldPt geometry.Point
	for i, pt := range t.points {
		bfpt := geometry.NewPoint(pt[0], pt[1])
		if i != 0 && geometry.ArePointsEqual(oldPt, bfpt) {
			continue
		}
		oldPt = bfpt
		if !t.sd.InsertSite(bfpt) {
			log.Printf("Failed to insert point %v", bfpt)
		}
	}
}
*/

func (t *Triangulator) Triangles(ctx context.Context, includeFrame bool) (triangles [][3]geometry.Point, err error) {
	sd := subdivision.NewForPoints(ctx, t.points)
	return sd.Triangles(includeFrame)
}

type Constrained struct {
	Points      [][2]float64
	Constraints [][2][2]float64
}

func (ct *Constrained) Triangles(ctx context.Context, includeFrame bool) (triangles [][3]geom.Point, err error) {
	pts := ct.Points
	for _, ct := range ct.Constraints {
		pts = append(pts, ct[0], ct[1])
	}
	sd := subdivision.NewForPoints(ctx, pts)
	vxidx := sd.VertexIndex()
	for _, ct := range ct.Constraints {
		err = sd.InsertConstraint(ctx, vxidx, geometry.NewPoint(ct[0][0], ct[0][1]), geometry.NewPoint(ct[1][0], ct[1][1]))
		if err != nil {
			return nil, err
		}

	}
	return sd.Triangles(includeFrame)
}

type byLength []geom.Line

func (l byLength) Len() int      { return len(l) }
func (l byLength) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l byLength) Less(i, j int) bool {
	lilen := l[i].LenghtSquared()
	ljlen := l[j].LenghtSquared()
	if lilen == ljlen {
		if cmp.PointEqual(l[i][0], l[j][0]) {
			return cmp.PointLess(l[i][1], l[j][1])
		}
		return cmp.PointLess(l[i][0], l[j][0])
	}
	return lilen < ljlen
}

type GeomConstrained struct {
	Points      []geom.Point
	Constraints []geom.Line
}

func (ct *GeomConstrained) Triangles(ctx context.Context, includeFrame bool) ([]geom.Triangle, error) {
	var pts [][2]float64
	for _, pt := range ct.Points {
		pts = append(pts, [2]float64(pt))
	}
	for _, ct := range ct.Constraints {
		pts = append(pts, ct[0], ct[1])
	}
	sd := subdivision.NewForPoints(ctx, pts)
	constraints := ct.Constraints
	sort.Sort(byLength(constraints))

	vxidx := sd.VertexIndex()
	for _, ct := range constraints {
		err := sd.InsertConstraint(ctx, vxidx, geometry.NewPoint(ct[0][0], ct[0][1]), geometry.NewPoint(ct[1][0], ct[1][1]))
		if err != nil {
			return nil, err
		}

	}
	var tris []geom.Triangle
	triangles, err := sd.Triangles(includeFrame)
	if err != nil {
		return nil, err
	}
	for _, tri := range triangles {
		tris = append(tris,
			geom.Triangle{
				geometry.UnwrapPoint(tri[0]),
				geometry.UnwrapPoint(tri[1]),
				geometry.UnwrapPoint(tri[2]),
			},
		)
	}
	return tris, nil
}
