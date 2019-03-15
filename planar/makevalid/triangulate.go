package makevalid

import (
	"context"
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/internal/debugger"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/delaunay"
)

func InsideTrianglesForGeometry(ctx context.Context, segs []geom.Line, hm planar.HitMapper) ([]geom.Triangle, error) {

	if debug {
		ctx = debugger.AugmentContext(ctx, "")
		defer debugger.Close(ctx)

		log.Printf("Step   3 : generate triangles")
	}

	builder := delaunay.NewConstrainedWithCtx(ctx, delaunay.TOLERANCE, []geom.Point{}, segs)

	allTriangles, err := builder.Triangles(false)
	if err != nil {
		if debug {
			log.Println("Step     3a: got error", err)
		}
		return nil, err
	}

	if debug {
		log.Printf("Step   4 : label triangles and discard outside triangles")
	}

	triangles := make([]geom.Triangle, 0, len(allTriangles))

	for i, triangle := range allTriangles {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		cpt := triangle.Center()
		lbl := hm.LabelFor(cpt)

		if debug {
			category := debuggerCategoryTriangle.With(lbl)
			debugger.Record(ctx, triangle, category, "triangle %v", i)
			debugger.Record(ctx, cpt, category, "triangle %v", i)
		}

		if lbl == planar.Outside {
			continue
		}
		triangles = append(triangles, triangle)
	}
	return triangles, nil

}
