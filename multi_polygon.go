package geom

type MultiPolygon [][][][2]float64

func (mp MultiPolygon) Polygons() [][][][2]float64 {
	return mp
}
