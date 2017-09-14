package geom

type Collection []Geometry

func (c Collection) Geometries() []Geometry {
	return c
}
