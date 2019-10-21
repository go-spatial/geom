package makevalid

import (
	"context"
	"log"

	"github.com/go-spatial/geom/encoding/wkt"

	qetriangulate "github.com/go-spatial/geom/planar/triangulate/gdey/quadedge"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/planar"
)

func InsideTrianglesForSegments(ctx context.Context, segs []geom.Line, hm planar.HitMapper) ([]geom.Triangle, error) {
	if debug {
		log.Printf("Step   3 : generate triangles")
	}
	triangulator := qetriangulate.GeomConstrained{
		Constraints: segs,
	}
	allTriangles, err := triangulator.Triangles(ctx, false)
	if err != nil {
		if debug {
			log.Println("Step     3a: got error", err)
		}
		return nil, err
	}
	if debug {
		log.Printf("Step   4 : label triangles and discard outside triangles")
		log.Printf("Step   4a: All Triangles:\n%v", wkt.MustEncode(allTriangles))
	}
	if len(allTriangles) == 0 {
		return []geom.Triangle{}, nil
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
	if debug {
		log.Printf("Step   4b: Inside Triangles:\n%v", wkt.MustEncode(triangles))
	}
	return triangles, nil

}
