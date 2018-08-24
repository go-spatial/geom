package makevalid

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/delaunay"
)

func InsideTrianglesForGeometry(ctx context.Context, segs []geom.Line, hm planar.HitMapper) ([]geom.Triangle, error) {
	if debug {
		log.Printf("Step   3 : generate triangles")
	}
	builder := delaunay.NewConstrained(delaunay.TOLERANCE, []geom.Point{}, segs)
	start := time.Now()
	allTriangles, err := builder.Triangles(false)
	fmt.Printf("triangulations of segs(%v) took %v\n", len(segs), time.Since(start))
	//fmt.Printf("Wkt alltriangles:\n%v\n", wkt.MustEncode(allTriangles))

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

	for _, triangle := range allTriangles {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if hm.LabelFor(triangle.Center()) == planar.Outside {
			continue
		}
		triangles = append(triangles, triangle)
	}
	return triangles, nil

}
