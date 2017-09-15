package geom

// Collection is a collection of one or more geometries.
type Collection []Geometry

// Geometries returns the slice of Geometries
func (c Collection) Geometries() []Geometry {
	return c
}
