package qetriangulate

import (
	"context"
	"github.com/go-spatial/geom/cmp"
	"log"
	"math"

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
	var constraints []geom.Line
	{
		var seen = make(map[[2]float64]bool)
		for _, pt := range ct.Points {

			if seen[[2]float64(pt)] {
				continue
			}

			seen[[2]float64(pt)] = true
			pts = append(pts, [2]float64(pt))
		}
		for i := range ct.Constraints {
			lnt := math.Sqrt(ct.Constraints[i].LengthSquared())
			if debug {
				log.Printf("for (%v)%v lnt: %v",i,ct.Constraints[i],lnt)
			}
			if cmp.Float(lnt, 0.0) {
				continue
			}
			if !seen[ct.Constraints[i][0]] {
				pts = append(pts, ct.Constraints[i][0])
				seen[ct.Constraints[i][0]] = true
			}
			if !seen[ct.Constraints[i][1]] {
				pts = append(pts, ct.Constraints[i][1])
				seen[ct.Constraints[i][1]] = true
			}
			constraints = append(constraints, ct.Constraints[i])
		}
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
