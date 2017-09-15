package geom

// MultiLineString is a geometry with multiple LineStrings.
type MultiLineString [][][2]float64

// LineStrings returns the coordinates for the linestrings
func (mls *MultiLineString) LineStrings() [][][2]float64 {
	return *mls
}
