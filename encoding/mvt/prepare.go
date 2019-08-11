package mvt

import (
	"log"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/tegola"
)

// PrepareGeo converts the geometry's coordinates to tile coordinates
func PrepareGeo(geo tegola.Geometry, tile *tegola.Tile) geom.Geometry {
	switch g := geo.(type) {
	case geom.Point:
		return preparept(g, tile)

	case geom.MultiPoint:
		pts := g.Points()
		if len(pts) == 0 {
			return nil
		}

		mp := make(geom.MultiPoint, len(pts))
		for i, pt := range g {
			mp[i] = preparept(pt, tile)
		}

		return mp

	case geom.LineString:
		return preparelinestr(g, tile)

	case geom.MultiLineString:
		var ml geom.MultiLineString
		for _, l := range g.LineStrings() {
			nl := preparelinestr(l, tile)
			if len(nl) > 0 {
				ml = append(ml, nl)
			}
		}
		return ml

	case geom.Polygon:
		return preparePolygon(g, tile)

	case geom.MultiPolygon:
		var mp geom.MultiPolygon
		for _, p := range g.Polygons() {
			np := preparePolygon(p, tile)
			if len(np) > 0 {
				mp = append(mp, np)
			}
		}
		return mp
	}

	return nil
}

func preparept(g geom.Point, tile *tegola.Tile) geom.Point {
	pt, err := tile.ToPixel(tegola.WebMercator, g)
	if err != nil {
		panic(err)
	}
	return geom.Point(pt)
}

func preparelinestr(g geom.LineString, tile *tegola.Tile) (ls geom.LineString) {
	pts := g
	// If the linestring
	if len(pts) < 2 {
		// Not enought points to make a line.
		return nil
	}
	ls = make(geom.LineString, 0, len(pts))
	ls = append(ls, preparept(pts[0], tile))
	for i := 1; i < len(pts); i++ {
		npt := preparept(pts[i], tile)
		ls = append(ls, npt)
	}

	if len(ls) < 2 {
		// Not enough points. the zoom must be too far out for this ring.
		return nil
	}
	return ls
}

func preparePolygon(g geom.Polygon, tile *tegola.Tile) (p geom.Polygon) {
	lines := geom.MultiLineString(g.LinearRings())
	p = make(geom.Polygon, 0, len(lines))

	if len(lines) == 0 {
		return p
	}

	for _, line := range lines.LineStrings() {
		ln := preparelinestr(line, tile)
		if len(ln) < 2 {
			if debug {
				// skip lines that have been reduced to less then 2 points.
				log.Println("skipping line 2", line, len(ln))
			}
			continue
		}
		// TODO: check the last and first point to make sure
		// they are not the same, per the mvt spec
		p = append(p, ln)
	}
	return p
}
