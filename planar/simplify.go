package planar

import "github.com/go-spatial/geom"

// Simplifer is an interface for Simplifying geometries.
type Simplifer interface {
	Simplify(linestring [][2]float64, isClosed bool) ([][2]float64, error)
}

func simplifyPolygon(simplifer Simplifer, plg [][][2]float64, isClosed bool) (ret [][][2]float64, err error) {
	ret = make([][][2]float64, len(plg))
	for i := range plg {
		ls, err := simplifer.Simplify(plg[i], isClosed)
		if err != nil {
			return nil, err
		}
		ret[i] = ls
	}
	return ret, nil

}

// Simplify will simplify the provided geometry using the provided simplifer.
// If the simplifer is nil, no simplification will be attempted.
func Simplify(simplifer Simplifer, geometry geom.Geometry) (geom.Geometry, error) {

	if simplifer == nil {
		return geometry, nil
	}

	switch gg := geometry.(type) {

	case geom.Collectioner:

		geos := gg.Geometries()
		coll := make([]geom.Geometry, len(geos))
		for i := range geos {
			geo, err := Simplify(simplifer, geos[i])
			if err != nil {
				return nil, err
			}
			coll[i] = geo
		}
		return geom.Collection(coll), nil

	case geom.MultiPolygoner:

		plys := gg.Polygons()
		mply := make([][][][2]float64, len(plys))
		for i := range plys {
			ply, err := simplifyPolygon(simplifer, plys[i], true)
			if err != nil {
				return nil, err
			}
			mply[i] = ply
		}
		return geom.MultiPolygon(mply), nil

	case geom.Polygoner:

		ply, err := simplifyPolygon(simplifer, gg.LinearRings(), true)
		if err != nil {
			return nil, err
		}
		return geom.Polygon(ply), nil

	case geom.MultiLineStringer:

		mls, err := simplifyPolygon(simplifer, gg.LineStrings(), false)
		if err != nil {
			return nil, err
		}
		return geom.MultiLineString(mls), nil

	case geom.LineStringer:

		ls, err := simplifer.Simplify(gg.Verticies(), false)
		if err != nil {
			return nil, err
		}
		return geom.LineString(ls), nil

	default: // Points, MutliPoints or anything else.
		return geometry, nil

	}
}
