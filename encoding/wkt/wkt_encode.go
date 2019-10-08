package wkt

import (
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

type Encoder struct {
	w io.Writer
	fbuf []byte

	// Strict causes the encoder to return errors if the geometries have empty
	// sub geometries where not allowed by the wkt spec. When Strict
	// false, empty geometries are ignored.
	//		Point: can be empty
	//		MultiPoint: can have empty points
	//		LineString: cannot have empty points
	//		MultiLineString: can have empty line strings non-empty line strings cannot have empty points
	//		Polygon: cannot have empty line strings, non-empty line strings cannot have empty points
	//		MultiPolygon: can have empty linestrings, polygons cannot have empty linestrings, line strings cannot have empty points
	//		Collection: can have empty geometries
	Strict bool

	// The precision that is passed into strconv.FormatFloat
	// https://golang.org/pkg/strconv/#FormatFloat
	Precision *int

	// The format flag that is passed into strconv.FormatFloat. If
	// performance is an issue then, the 'E' or 'e' flag is recommended.
	// https://golang.org/pkg/strconv/#FormatFloat
	Fmt byte
}

var (
	DefaultEncoderStrict = true
	DefaultEncoderPrecision = 5
	DefaultEncoderFmt = byte('g')
)

func NewEncoder(w io.Writer) *Encoder {
	// our float would be one of
	// ddd.ddd with Precision+1 number of d's => Precision + 1 + 1
	// d.ddde+xx with Precision+1 number of d's => Precision + 1 + 5
	prec := DefaultEncoderPrecision
	return &Encoder{
		w: w,
		fbuf: make([]byte, 0, prec + 6),
		Strict: DefaultEncoderStrict,
		Precision: &prec,
		Fmt: DefaultEncoderFmt,
	}
}

func (enc *Encoder) byte(b byte) error {
	buf := append(enc.fbuf[:0], b)
	_, err := enc.w.Write(buf)
	return err
}

func (enc *Encoder) string(s string) error {
	_, err := enc.w.Write([]byte(s))
	return err
}


func (enc *Encoder) formatFloat(f float64) error {
	buf := strconv.AppendFloat(enc.fbuf[:0], f, enc.Fmt, *enc.Precision, 64)
	_, err := enc.w.Write(buf)
	return err
}

// var EmptyPoint = [2]float64{math.NaN(), math.NaN()}


func (enc *Encoder) encodePair(pt [2]float64) error {
	// should onlt be called for multipoints
	if cmp.IsEmptyPoint(pt) {
		return enc.string("EMPTY")
	}

	err := enc.formatFloat(pt[0])
	if err != nil {
		return err
	}

	err = enc.byte(' ')
	if err != nil {
		return err
	}

	return enc.formatFloat(pt[1])
}

func (enc *Encoder) encodePoint(pt [2]float64) error {
	// empty point
	if cmp.IsEmptyPoint(pt) {
		err := enc.string("EMPTY")
		return err
	}

	err := enc.byte('(')
	if err != nil {
		return err
	}

	err = enc.encodePair(pt)
	if err != nil {
		return err
	}

	return enc.byte(')')
}

func lastNonEmptyIdxPoints(mp [][2]float64) (last int) {
	for i := len(mp) - 1; i >= 0; i-- {
		if !cmp.IsEmptyPoint(mp[i]) {
			return i
		}
	}

	return -1
}

func lastNonEmptyIdxLines(lines [][][2]float64) (last int) {
	for i := len(lines) - 1; i >= 0; i-- {
		last := lastNonEmptyIdxPoints(lines[i])
		if last != -1 {
			return i
		}
	}
	return -1
}

func lastNonEmptyIdxPolys(polys [][][][2]float64) (last int) {
	for i := len(polys) - 1; i >= 0; i-- {
		last := lastNonEmptyIdxLines(polys[i])
		if last != -1 {
			return i
		}
	}
	return -1
}


func (enc *Encoder) encodePoints(mp [][2]float64, last int, gType byte) (err error) {

	// the last encode point
	var firstEnc *[2]float64
	var lastEnc *[2]float64
	var count int


	for i, v := range mp[:last+1] {
		// if the last point is the same as this point and
		// we aren't encoding a multipoint, then dups should get dropped
		if lastEnc != nil && *lastEnc == v && gType != mpType {
			continue
		}

		if cmp.IsEmptyPoint(v) {
			if enc.Strict {
				switch gType {
				case mpType:
					// multipoints can have empty points
					// encodePair will write EMPTY
					break
				case lsType:
					return errors.New("cannot have empty points in strict LINESTRING")
				case mlType:
					return errors.New("cannot have empty points in strict MULTILINESTRING")
				case polyType:
					return errors.New("cannot have empty points in strict POLYGON")
				case mPolyType:
					return errors.New("cannot have empty points in strict MULTIPOLYGON")
				default:
					panic("unrechable")
				}
			} else {
				// multipoints can have empty points
				// encodePair will write EMPTY
				if gType != mpType {
					continue
				}
			}
		}

		// this also the first encoded point
		// we save it in case we need to close the polygon later
		if firstEnc == nil {
			firstEnc = &mp[i]
		}

		// update what the last encoded value is
		lastEnc = &mp[i]

		switch count {
		case 0:
			err = enc.byte('(')
		default:
			err = enc.byte(',')
		}
		if err != nil {
			return err
		}

		count++
		err = enc.encodePair(v)
		if err != nil {
			return err
		}
	}

	// do size checking before encoding a closing point
	if count == 0 {
		return enc.string("EMPTY")
	}

	if (gType == polyType || gType == mPolyType) && count < 3 {
		return fmt.Errorf("not enough points for linear ring of POLYGON %v", mp)
	} else if (gType == lsType || gType == mlType) && count < 2{
		return fmt.Errorf("not enough points for LINESTRING %v", mp)
	}


	// if we need to close the polygon/multipolygon
	// and the value we encoded last isn't (already) the last
	// value to encode
	if (gType == polyType || gType == mPolyType) && *firstEnc != *lastEnc {
		err = enc.byte(',')
		if err != nil {
			return err
		}
		err = enc.encodePair(*firstEnc)
		if err != nil {
			return err
		}
	}

	return enc.byte(')')
}

const (
	mpType byte = iota
	lsType
	mlType
	polyType
	mPolyType
)

func (enc *Encoder) encodeLines(lines [][][2]float64, last int, gType byte) error {
	if gType != mlType {
		idx := lastNonEmptyIdxLines(lines)
		if idx != last && enc.Strict {
			switch gType {
			case polyType:
				return errors.New("cannot have empty linear ring in strict POLYGON")
			case mPolyType:
				return errors.New("cannot have empty linear ring in strict MULTIPOLYGON")
			case mlType:
				// empty linestrings are allowed in multilines
				break
			default:
				panic("unrechable")
			}
		} else {
			last = idx
		}
	}

	if last == -1 {
		return enc.string("EMPTY")
	}

	err := enc.byte('(')
	if err != nil {
		return err
	}

	for _, v := range lines[:last] {
		// polygons and multipolygons cannot have empty linestrings
		if lastNonEmptyIdxPoints(v) == -1 {
			if enc.Strict {
				switch gType {
				case polyType:
					return errors.New("cannot have empty linear ring in strict POLYGON")
				case mPolyType:
					return errors.New("cannot have emtpy linear ring in strict MULTIPOLYGON")
				case mlType:
					// empty linestrings are allowed in
					// encodePoints writes EMPTY
					break
				default:
					panic("unrechable")
				}
			} else {
				// empty linestrings are allowed in
				// encodePoints writes EMPTY
				if gType != mlType {
					continue
				}
			}
		}

		err := enc.encodePoints(v, len(v) - 1, gType)
		if err != nil {
			return err
		}

		err = enc.byte(',')
		if err != nil {
			return err
		}
	}

	err = enc.encodePoints(lines[last], len(lines[last]) - 1, gType)
	if err != nil {
		return err
	}

	return enc.byte(')')
}

func (enc *Encoder) encodePolys(polys [][][][2]float64, last int) error {
	if last == -1 {
		return enc.string("EMPTY")
	}

	err := enc.byte('(')
	if err != nil {
		return err
	}

	for _, v := range polys[:last] {
		err = enc.encodeLines(v, len(v) - 1, mPolyType)
		if err != nil {
			return err
		}

		err = enc.byte(',')
		if err != nil {
			return err
		}
	}

	err = enc.encodeLines(polys[last], len(polys[last]) - 1, mPolyType)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	return enc.byte(')')
}

func (enc *Encoder) encode(geo geom.Geometry, isCollectionItem bool) error {

	switch g := geo.(type) {
	case [2]float64:
		err := enc.string("POINT ")
		if err != nil {
			return err
		}
		return enc.encodePoint(g)

	case geom.Pointer:
		if enc.Strict && isCollectionItem && (cmp.IsNil(g) || cmp.IsEmptyPoint(g.XY())) {
			return nil
		}

		err := enc.string("POINT ")
		if err != nil {
			return err
		}

		if cmp.IsNil(g) {
			return enc.string("EMPTY")
		}

		return enc.encodePoint(g.XY())

	case geom.MultiPointer:
		var mp [][2]float64

		if !cmp.IsNil(g) {
			mp = g.Points()
		}

		err := enc.string("MULTIPOINT ")
		if err != nil {
			return err
		}

		err = enc.encodePoints(mp, len(mp) - 1, mpType)
		return err

	case geom.LineStringer:
		var mp [][2]float64

		if !cmp.IsNil(g) {
			mp = g.Verticies()
		}

		err := enc.string("LINESTRING ")
		if err != nil {
			return err
		}

		err = enc.encodePoints(mp, len(mp) - 1, lsType)
		if err != nil {
			return err
		}

		return nil

	case geom.MultiLineStringer:
		var lines [][][2]float64

		if !cmp.IsNil(g) {
			lines = g.LineStrings()
		}


		err := enc.string("MULTILINESTRING ")
		if err != nil {
			return err
		}

		return enc.encodeLines(lines, len(lines) - 1, mlType)

	case geom.Polygoner:
		var lines [][][2]float64

		if !cmp.IsNil(g) {
			lines = g.LinearRings()
		}

		err := enc.string("POLYGON ")
		if err != nil {
			return err
		}

		return enc.encodeLines(lines, len(lines) - 1, polyType)

	case geom.MultiPolygoner:
		var polys [][][][2]float64

		if !cmp.IsNil(g) {
			polys = g.Polygons()
		}

		err := enc.string("MULTIPOLYGON ")
		if err != nil {
			return err
		}

		return enc.encodePolys(polys, len(polys) - 1)

	case geom.Collectioner:
		var geoms []geom.Geometry

		if !cmp.IsNil(g) {
			geoms = g.Geometries()
		}

		if len(geoms) == 0 {
			return enc.string("GEOMETRYCOLLECTION EMPTY")
		}

		err := enc.string("GEOMETRYCOLLECTION ")
		if err != nil {
			return err
		}

		err = enc.byte('(')
		if err != nil {
			return err
		}

		last := len(geoms) - 1

		for _, v := range geoms[:last] {
			err = enc.encode(v, true)
			if err != nil {
				return err
			}

			err = enc.byte(',')
			if err != nil {
				return err
			}
		}

		err = enc.encode(geoms[last], true)
		if err != nil {
			return err
		}

		return enc.byte(')')

	// non basic types

	case geom.Line:
		return enc.encode(geom.LineString(g[:]), false)

	case [2][2]float64:
		return enc.encode(geom.LineString(g[:]), false)

	case [][2]float64:
		return enc.encode(geom.LineString(g), false)

	case []geom.Line:
		lines := make(geom.MultiLineString, len(g))
		for i, v := range g {
			lines[i] = [][2]float64(v[:])
		}

		return enc.encode(lines, false)

	case []geom.Point:
		points := make(geom.MultiPoint, len(g))
		for i, v := range g {
			points[i] = v
		}

		return enc.encode(points, false)

	case geom.Triangle:
		return enc.encode(geom.Polygon{g[:]}, false)

	case geom.Extent:
		return enc.encode(g.AsPolygon(), false)

	case *geom.Extent:
		if g != nil {
			return enc.encode(g.AsPolygon(), false)
		}

		return enc.encode(geom.Polygon{}, false)

	default:
		return fmt.Errorf("unknown geometry: %T", geo)
	}
}

func (enc *Encoder) Encode(geo geom.Geometry) error {

	if enc.Precision == nil {
		prec := DefaultEncoderPrecision
		enc.Precision = &prec
	}

	if enc.fbuf == nil {
		enc.fbuf = make([]byte, 0, *enc.Precision + 6)
	}

	if enc.Fmt == 0 {
		enc.Fmt = DefaultEncoderFmt
	}

	return enc.encode(geo, false)
}
