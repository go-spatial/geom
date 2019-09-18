package qetriangulate

import (
	"context"
	"log"
	"sort"

	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/quadedge"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar/triangulate/gdey/quadedge/subdivision"
)

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
	sd, err := subdivision.NewForPoints(ctx, pts)
	if err != nil {
		if debug && err != context.Canceled {
			if err1, ok := err.(quadedge.ErrInvalid); ok {
				for i, estr := range err1 {
					log.Printf("%v Err: %v", i, estr)
				}
			} else {
				log.Printf("Err: %v", err)
			}
			log.Printf("Points: %v\n", wkt.MustEncode(geom.MultiPoint(pts)))
		}
		return nil, err
	}
	constraints := ct.Constraints
	sort.Sort(planar.LinesByLength(constraints))

	vxidx := sd.VertexIndex()
	total := len(constraints)
	for i, ct := range constraints {
		if debug {
			log.Printf("working on constraint %v of %v", i, total)
		}
		err := sd.InsertConstraint(ctx, vxidx, geom.Point(ct[0]), geom.Point(ct[1]))
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
				[2]float64(tri[0]),
				[2]float64(tri[1]),
				[2]float64(tri[2]),
			},
		)
	}
	return tris, nil

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
	sd, err := subdivision.NewForPoints(ctx, pts)
	if err != nil {
		return nil, err
	}

	vxidx := sd.VertexIndex()
	total := len(ct.Constraints)
	for i, ct := range ct.Constraints {
		log.Printf("working on constraint %v of %v", i, total)
		err = sd.InsertConstraint(ctx, vxidx, geom.Point(ct[0]), geom.Point(ct[1]))
		if err != nil {
			return nil, err
		}
	}
	return sd.Triangles(includeFrame)
}
