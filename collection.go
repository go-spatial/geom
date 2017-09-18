package geom

// Collection is a collection of one or more geometries.
type Collection []Geometry

// Geometries returns the slice of Geometries
func (c Collection) Geometries() []Geometry {
	return c
}

// Points returns a slice of XY values
func (c Collection) Points() (points [][2]float64) {
	for _, g := range c {
		points = append(points, g.Points()...)
	}
	return
}

// SetGeometries modifies the array of 2D coordinates
func (c *Collection) SetGeometries(input []Geometry) (err error) {
	*c = append((*c)[:0], input...)
	return
}
