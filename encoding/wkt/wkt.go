package wkt

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-spatial/geom"
)

var ErrInvalidWKT = errors.New("WKT is not in valid form")

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

func isMultiLineStringerEmpty(ml geom.MultiLineStringer) bool {
	if isNil(ml) || len(ml.LineStrings()) == 0 {
		return true
	}
	lns := ml.LineStrings()
	// It's not nil, and there are several lines.
	// We need to go through all the lines and make sure that at least one of them has a non-zero length.
	for i := range lns {
		if len(lns[i]) != 0 {
			return false
		}
	}
	return true
}

func isPolygonerEmpty(p geom.Polygoner) bool {
	if isNil(p) || len(p.LinearRings()) == 0 {
		return true
	}
	lns := p.LinearRings()
	// It's not nil, and there are several lines.
	// We need to go through all the lines and make sure that at least one of them has a non-zero length.
	for i := range lns {
		if len(lns[i]) != 0 {
			return false
		}
	}
	return true
}

func isMultiPolygonerEmpty(mp geom.MultiPolygoner) bool {
	if isNil(mp) || len(mp.Polygons()) == 0 {
		return true
	}
	plys := mp.Polygons()
	for i := range plys {
		for j := range plys[i] {
			if len(plys[i][j]) != 0 {
				return false
			}
		}
	}
	return true
}

func isCollectionerEmpty(col geom.Collectioner) bool {
	if isNil(col) || len(col.Geometries()) == 0 {
		return true
	}
	geos := col.Geometries()
	for i := range geos {
		switch g := geos[i].(type) {
		case geom.Pointer:
			if !isNil(g) {
				return false
			}
		case geom.MultiPointer:
			if !(isNil(g) || len(g.Points()) == 0) {
				return false
			}
		case geom.LineStringer:
			if !(isNil(g) || len(g.Verticies()) == 0) {
				return false
			}
		case geom.MultiLineStringer:
			if !isMultiLineStringerEmpty(g) {
				return false
			}
		case geom.Polygoner:
			if !isPolygonerEmpty(g) {
				return false
			}
		case geom.MultiPolygoner:
			if !isMultiPolygonerEmpty(g) {
				return false
			}
		case geom.Collectioner:
			if !isCollectionerEmpty(g) {
				return false
			}
		}
	}
	return true
}

/*
This purpose of this file is to house the wkt functions. These functions are
use to take a tagola.Geometry and convert it to a wkt string. It will, also,
contain functions to parse a wkt string into a wkb.Geometry.
*/

func _encode(geo geom.Geometry) string {

	switch g := geo.(type) {

	case geom.Pointer:
		xy := g.XY()
		return fmt.Sprintf("%v %v", xy[0], xy[1])

	case geom.MultiPointer:
		var points []string
		for _, p := range g.Points() {
			points = append(points, _encode(geom.Point(p)))
		}
		return "(" + strings.Join(points, ",") + ")"

	case geom.LineStringer:
		var points []string
		for _, p := range g.Verticies() {
			points = append(points, _encode(geom.Point(p)))
		}
		return "(" + strings.Join(points, ",") + ")"

	case geom.MultiLineStringer:
		var lines []string
		for _, l := range g.LineStrings() {
			if len(l) == 0 {
				continue
			}
			lines = append(lines, _encode(geom.LineString(l)))
		}
		return "(" + strings.Join(lines, ",") + ")"

	case geom.Polygoner:
		var rings []string
		for _, l := range g.LinearRings() {
			if len(l) == 0 {
				continue
			}
			rings = append(rings, _encode(geom.LineString(l)))
		}
		return "(" + strings.Join(rings, ",") + ")"

	case geom.MultiPolygoner:
		var polygons []string
		for _, p := range g.Polygons() {
			if len(p) == 0 {
				continue
			}
			polygons = append(polygons, _encode(geom.Polygon(p)))
		}
		return "(" + strings.Join(polygons, ",") + ")"

	}
	panic(fmt.Sprintf("Don't know the geometry type! %+v", geo))
}

//WKT returns a WKT representation of the Geometry if possible.
// the Error will be non-nil if geometry is unknown.
func Encode(geo geom.Geometry) (string, error) {
	switch g := geo.(type) {
	default:
		return "", geom.ErrUnknownGeometry{geo}
	case geom.Pointer:
		// POINT( 10 10)
		if isNil(g) {
			return "POINT EMPTY", nil
		}
		return "POINT (" + _encode(geo) + ")", nil

	case geom.MultiPointer:
		if isNil(g) || len(g.Points()) == 0 {
			return "MULTIPOINT EMPTY", nil
		}
		return "MULTIPOINT " + _encode(geo), nil

	case geom.LineStringer:
		if isNil(g) || len(g.Verticies()) == 0 {
			return "LINESTRING EMPTY", nil
		}
		return "LINESTRING " + _encode(geo), nil

	case geom.MultiLineStringer:
		if isMultiLineStringerEmpty(g) {
			return "MULTILINESTRING EMPTY", nil
		}
		return "MULTILINESTRING " + _encode(geo), nil

	case geom.Polygoner:
		if isPolygonerEmpty(g) {
			return "POLYGON EMPTY", nil
		}
		return "POLYGON " + _encode(geo), nil

	case geom.MultiPolygoner:
		if isMultiPolygonerEmpty(g) {
			return "MULTIPOLYGON EMPTY", nil
		}
		return "MULTIPOLYGON " + _encode(geo), nil

	case geom.Collectioner:
		if isCollectionerEmpty(g) {
			return "GEOMETRYCOLLECTION EMPTY", nil
		}
		var geometries []string
		for _, sg := range g.Geometries() {
			s, err := Encode(sg)
			if err != nil {
				return "", err
			}
			geometries = append(geometries, s)
		}
		return "GEOMETRYCOLLECTION (" + strings.Join(geometries, ",") + ")", nil
	}
}

func strCordToPoint(points string) ([]float64, error) {
	splitP := strings.Split(strings.TrimSpace(points), " ")
	floats := make([]float64, len(splitP))
	for i, p := range splitP {
		if floatP, err := strconv.ParseFloat(p, 64); err != nil {
			return []float64{}, errors.New(fmt.Sprintf("Couldn't parse coordinate: %s", p))
		} else {
			floats[i] = floatP
		}
	}

	return floats, nil
}
func Decode(wktInput string) (geo geom.Geometry, err error) {
	if wktInput == "" { // Empty input is not an error but a nil geom
		return nil, geom.ErrUnknownGeometry{nil}
	}

	wktUpper := strings.ToUpper(wktInput)

	typeRegex := regexp.MustCompile(`(^\S*)\s*(EMPTY|ZM|Z|M)*(?:\s*\(\s*(.*)\s*\))*`)
	matchedGeom := typeRegex.FindStringSubmatch(wktUpper)
	if len(matchedGeom) != 4 {
		return nil, ErrInvalidWKT
	}

	isEmpty := matchedGeom[2] == "EMPTY"
	if !isEmpty && matchedGeom[2] != "" {
		return nil, errors.New("Z and/or M is not supported")
	}

	switch matchedGeom[1] {
	case "POINT":
		if isEmpty {
			return (*geom.Point)(nil), nil
		}

		if parsedPoints, err := strCordToPoint(matchedGeom[3]); err != nil {
			return nil, err
		} else {
			return geom.Point([2]float64{parsedPoints[0], parsedPoints[1]}), nil
		}
	case "MULTIPOINT":
		if isEmpty {
			return (*geom.MultiPoint)(nil), nil
		}

		points := [][2]float64{}
		for _, p := range strings.Split(matchedGeom[3], ",") {
			if parsedPoints, err := strCordToPoint(p); err != nil {
				return nil, err
			} else {
				points = append(points, [2]float64{parsedPoints[0], parsedPoints[1]})
			}
		}

		return geom.MultiPoint(points), nil
	case "LINESTRING":
		if isEmpty {
			return (*geom.LineString)(nil), nil
		}

		points := [][2]float64{}
		for _, p := range strings.Split(matchedGeom[3], ",") {
			if parsedPoints, err := strCordToPoint(p); err != nil {
				return nil, err
			} else {
				points = append(points, [2]float64{parsedPoints[0], parsedPoints[1]})
			}
		}

		return geom.LineString(points), nil
	case "MULTILINESTRING":
		if isEmpty {
			return (*geom.MultiLineString)(nil), nil
		}

		reg := regexp.MustCompile(`\)\s*,\s*\(`)
		indexes := reg.FindAllStringIndex(matchedGeom[3], -1)
		subPoints := make([]string, len(indexes))

		lastHead, lastTail := 1, 0
		for i, index := range indexes {
			lastTail = index[0]
			subPoints[i] = matchedGeom[3][lastHead:lastTail]
			lastHead = index[1]
		}
		subPoints = append(subPoints, matchedGeom[3][lastHead:len(matchedGeom[3])-1])

		multiPoints := [][][2]float64{}
		for _, subP := range subPoints {
			points := [][2]float64{}
			for _, p := range strings.Split(subP, ",") {
				if parsedPoints, err := strCordToPoint(p); err != nil {
					return nil, err
				} else {
					points = append(points, [2]float64{parsedPoints[0], parsedPoints[1]})
				}
			}
			multiPoints = append(multiPoints, points)
		}

		return geom.MultiLineString(multiPoints), nil
	default:
		return nil, nil
	}

	return nil, nil
}
