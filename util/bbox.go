package util

import (
	"log"

	"github.com/go-spatial/geom"
)

//	TODO: maybe setup UL, LR as geom.Point to e more explicit?
type BoundingBox [4]float64

func BBox(g geom.Geometry) BoundingBox {
	var points []geom.Point

	switch t := g.(type) {
	case geom.Point:
		points = append(points, t)
	case geom.LineString:
		ls := t.(geom.LineString)
		points = append(points, ls.SubPoints()...)
	default:
		log.Printf("geom type (%T) not supported", g)
	}

	var bbox BoundingBox
	for i := range points {
		x, y := points[i].XY()

		if i == 0 {
			bbox[0] = x
			bbox[1] = y
			bbox[2] = x
			bbox[3] = y
			continue
		}
		if x < bbox[0] {
			bbox[0] = x
		}
		if x > bbox[2] {
			bbox[2] = x
		}
		if y < bbox[1] {
			bbox[1] = y
		}
		if y > bbox[3] {
			bbox[3] = y
		}
	}

	return bbox
}
