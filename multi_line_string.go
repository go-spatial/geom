package geom

// MultiLineString is a geometry with multiple LineStrings.
type MultiLineString [][][2]float64

// LineStrings returns the coordinates for the linestrings
func (mls *MultiLineString) LineStrings() [][][2]float64 {
	return *mls
}

// Points returns a slice of XY values
func (mls *MultiLineString) Points() (points [][2]float64) {
	for _, ls := range *mls {
		points = append(points, ls...)
	}
	return
}
