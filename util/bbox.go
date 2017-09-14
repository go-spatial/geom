package util

//	TODO: maybe setup UL, LR as geom.Point to e more explicit?
type BoundingBox [4]float64

func BBox(points [][2]float64) (bbox BoundingBox) {

	for i := range points {
		if i == 0 {
			bbox[0] = points[i][0]
			bbox[1] = points[i][1]
			bbox[2] = points[i][0]
			bbox[3] = points[i][1]
			continue
		}

		switch {
		case points[i][0] < bbox[0]:
			bbox[0] = points[i][0]
		case points[i][0] > bbox[2]:
			bbox[2] = points[i][0]
		}

		switch {
		case points[i][1] < bbox[1]:
			bbox[1] = points[i][1]
		case points[i][1] > bbox[3]:
			bbox[3] = points[i][1]
		}
	}

	return
}
