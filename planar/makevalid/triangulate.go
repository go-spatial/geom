package makevalid

import (
	"context"
	"log"
	"time"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
	"github.com/go-spatial/geom/planar"
	"github.com/go-spatial/geom/planar/triangulate/delaunay"
)

func InsideTrianglesForGeometry(ctx context.Context, segs []geom.Line, hm planar.HitMapper) ([]geom.Triangle, error) {
	if debug {
		log.Printf("Step   3 : generate triangles")
	}
	builder := delaunay.NewConstrained(delaunay.TOLERANCE, []geom.Point{}, segs)
	var start time.Time
	if debug {
		start = time.Now()
	}

	allTriangles, err := builder.Triangles(false)
	if err != nil {
		if debug {
			log.Println("Step     3a: got error", err)
		}
		return nil, err
	}

	if debug {
		log.Printf("triangulations took %v\n", time.Since(start))
		log.Printf("Got %v trinangles\n", len(allTriangles))
		for i, tri := range allTriangles {
			log.Printf("% 4d: %v\n", i, wkt.MustEncode(tri))
		}
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
			log.Printf("For %04v Triangle: %v\n\tCPoint: %v\n\tLabel:%v",
				i,
				wkt.MustEncode(triangle),
				wkt.MustEncode(cpt),
				lbl,
			)
		}
		if lbl == planar.Outside {
			continue
		}
		triangles = append(triangles, triangle)
	}
	return triangles, nil

}
