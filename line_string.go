package geom

// LineString is a basic line type which is made up of two or more points that don't interect.
type LineString [][2]float64

// Points returns a slice of XY values
func (ls LineString) Points() [][2]float64 {
	return ls
}

// Vertexes returns a slice of XY values
func (ls LineString) Vertexes() [][2]float64 {
	return ls.Points()
}

// SetVertexes modifies the array of 2D coordinates
func (ls *LineString) SetVertexes(input [][2]float64) (err error) {
	*ls = append((*ls)[:0], input...)
	return
}
