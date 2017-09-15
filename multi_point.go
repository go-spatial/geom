package geom

// MultiPoint is a geometry with multiple points.
type MultiPoint [][2]float64

// Points returns the coordinates for the points
func (mp *MultiPoint) Points() [][2]float64 {
	return *mp
}
