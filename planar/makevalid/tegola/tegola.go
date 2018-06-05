package tegola

import "github.com/go-spatial/geom/planar"

type Makevalid struct {
	Hitmap planar.HitMapper
	// Currently not used, but once we have the IsValid function, we can use this instead
	// Of running the MakeValid routine on a Geometry that is alreayd valid.
	// Used to clip geometries that are not Polygon and MultiPolygons
	Clipper planar.Clipper
}
