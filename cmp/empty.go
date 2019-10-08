package cmp

import (
	"fmt"
	"reflect"

	"github.com/go-spatial/geom"
)

func IsEmptyPoint(pt [2]float64) bool {
	return pt != pt
}

func IsEmptyPoints(pts [][2]float64) bool {
	for _, v := range pts {
		if !IsEmptyPoint(v) {
			return false
		}
	}

	return true
}

func IsEmptyLines(lns [][][2]float64) bool {
	for _, v := range lns {
		if !IsEmptyPoints(v) {
			return false
		}
	}

	return true
}

func IsNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

func IsEmptyGeo(geo geom.Geometry) (isEmpty bool, err error) {
	if IsNil(geo) {
		return true, nil
	}

	switch g := geo.(type) {
	case [2]float64:
		return IsEmptyPoint(g), nil

	case geom.Pointer:
		return IsEmptyPoint(g.XY()), nil

	case [][2]float64:
		return IsEmptyPoints(g), nil

	case geom.MultiPointer:
		return IsEmptyPoints(g.Points()), nil

	case geom.LineStringer:
		return IsEmptyPoints(g.Verticies()), nil

	case geom.MultiLineStringer:
		return IsEmptyLines(g.LineStrings()), nil

	case geom.Polygoner:
		return IsEmptyLines(g.LinearRings()), nil

	case geom.MultiPolygoner:
		for _, v := range g.Polygons() {
			if !IsEmptyLines(v) {
				return false, nil
			}
		}

		return true, nil

	case geom.Collectioner:
		for _, v := range g.Geometries() {
			isEmpty, err := IsEmptyGeo(v)
			if err != nil {
				return false, err
			}
			if !isEmpty {
				return false, nil
			}
		}

		return true, nil
	default:
		return false, fmt.Errorf("unknown geometry %T", geo)
	}
}
