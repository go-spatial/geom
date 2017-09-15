package geom

// MultiPolygon is a geometry of multiple polygons.
type MultiPolygon [][][][2]float64

// Polygons returns the array of polygons.
func (mp *MultiPolygon) Polygons() [][][][2]float64 {
	return *mp
}
