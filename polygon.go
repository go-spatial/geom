package geom

// Polygon is a geometry consisting of multiple closed LineStrings.
// There must be only one exterior LineString with a clockwise winding order.
// There may be one or more interior LineStrings with a counterclockwise winding orders.
type Polygon [][][2]float64

// LineStrings returns the coordinates of the lineStrings
func (p *Polygon) LineStrings() [][][2]float64 {
	return *p
}

// Points returns a slice of XY values
func (p *Polygon) Points() (points [][2]float64) {
	for _, ls := range *p {
		points = append(points, ls...)
	}
	return
}

// SetLineStrings modifies the array of 2D coordinates
func (p *Polygon) SetLineStrings(input [][][2]float64) (err error) {
	*p = append((*p)[:0], input...)
	return
}
