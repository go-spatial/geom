package wkt

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"

	"github.com/go-spatial/geom"
)

func isNil(a interface{}) bool {
	defer func() { recover() }()
	return a == nil || reflect.ValueOf(a).IsNil()
}

type Encoder struct {
	w io.Writer
	fbuf []byte
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
		fbuf: make([]byte, 0, 16),
	}
}

func (enc *Encoder) putc(b byte) error {
	buf := append(enc.fbuf[:0], b)
	_, err := enc.w.Write(buf)
	return err
}

func (enc *Encoder) puts(s string) error {
	_, err := enc.w.Write([]byte(s))
	return err
}


func (enc *Encoder) formatFloat(f float64) error {
	// for reference, the min value of EPSG:3857 -20026376.390 -> 13 characters
	buf := strconv.AppendFloat(enc.fbuf[:0], f, 'f', 3, 64)
	i := len(buf) - 1;
	for ; i >= 0; i-- {
		if buf[i] != '0' {
			break
		}
	}
	if buf[i] == '.' {
		i--
	}
	buf = buf[:i+1]
	_, err := enc.w.Write(buf)
	return err
}

// var EmptyPoint = [2]float64{math.NaN(), math.NaN()}

func IsEmptyPoint(pt [2]float64) bool {
	return pt != pt
}

func (enc *Encoder) encodePair(pt [2]float64) error {
	if IsEmptyPoint(pt) {
		return enc.puts("EMPTY")
	}

	err := enc.formatFloat(pt[0])
	if err != nil {
		return err
	}

	err = enc.putc(' ')
	if err != nil {
		return err
	}

	return enc.formatFloat(pt[1])
}

func (enc *Encoder) encodePoint(pt [2]float64) error {
	// empty point
	if IsEmptyPoint(pt) {
		err := enc.puts("EMPTY")
		return err
	}

	err := enc.putc('(')
	if err != nil {
		return err
	}

	err = enc.encodePair(pt)
	if err != nil {
		return err
	}

	return enc.putc(')')
}

func isPointlessPoints(mp [][2]float64) (last int, isPointless bool) {
	for i := len(mp) - 1; i >= 0; i-- {
		if !IsEmptyPoint(mp[i]) {
			return i, false
		}
	}

	return -1, true
}

func isPointlessLines(lines [][][2]float64) (last int, isPointless bool) {
	for i := len(lines) - 1; i >= 0; i-- {
		_, pl := isPointlessPoints(lines[i])
		if !pl {
			return i, false
		}
	}
	return -1, true
}

func isPointlessPolys(polys [][][][2]float64) (last int, isPointless bool) {
	for i := len(polys) - 1; i >= 0; i-- {
		_, pl := isPointlessLines(polys[i])
		if !pl {
			return i, false
		}
	}
	return -1, true
}

func isPointlessGeo(geo geom.Geometry) (isPointless bool, err error) {
	if isNil(geo) {
		return true, nil
	}

	switch g := geo.(type) {
	case [2]float64:
		return IsEmptyPoint(g), nil

	case geom.Pointer:
		return IsEmptyPoint(g.XY()), nil

	case geom.MultiPointer:
		_, pl := isPointlessPoints(g.Points())
		return pl, nil

	case geom.LineStringer:
		_, pl := isPointlessPoints(g.Verticies())
		return pl, nil

	case geom.MultiLineStringer:
		_, pl := isPointlessLines(g.LineStrings())
		return pl, nil

	case geom.Polygoner:
		_, pl := isPointlessLines(g.LinearRings())
		return pl, nil

	case geom.MultiPolygoner:
		_, pl := isPointlessPolys(g.Polygons())
		return pl, nil

	case geom.Collectioner:
		for _, v := range g.Geometries() {
			pl, err := isPointlessGeo(v)
			if err != nil {
				return false, err
			}
			if !pl {
				return false, nil
			}
		}
		return true, nil
	default:
		return false, fmt.Errorf("unknown geometry %T", geo)
	}
}

func (enc *Encoder) encodePoints(mp [][2]float64, removePointless bool, last int, gType byte) (count int, err error) {
	if !removePointless || gType != mpType {
		last = len(mp) - 1
	}

	// the last encode point
	_v := mp[last]
	lastEnc := &_v
	if gType == polyType || gType == mPolyType {
		last++
		lastEnc = nil
	}

	err = enc.putc('(')
	if err != nil {
		return count, err
	}

	for _, v := range mp[:last] {
		if IsEmptyPoint(v) {
			switch gType {
			case mpType:
				// multipoints can have empty points
				if removePointless {
					continue
				}
			case lsType:
				return count, errors.New("cannot have empty points in LINESTRING")
			case mlType:
				return count, errors.New("cannot have empty points in MULTILINESTRING")
			case polyType:
				return count, errors.New("cannot have empty points in POLYGON")
			case mPolyType:
				return count, errors.New("cannot have empty points in MULTIPOLYGON")
			default:
				panic("unrechable")
			}
		}

		if lastEnc == nil {
			_v = v
			lastEnc = &_v
		}

		count++
		err = enc.encodePair(v)
		if err != nil {
			return count, err
		}

		err = enc.putc(',')
		if err != nil {
			return count, err
		}
	}

	if IsEmptyPoint(*lastEnc) {
		fmt.Printf("empty point, type is  %v\n", gType)
		switch gType {
		case mpType:
			// multipoints can have empty points
			if removePointless {
				panic("the last element must always be poinfull, set by caller")
			}
		case lsType:
			return count, errors.New("cannot have empty points in LINESTRING")
		case mlType:
			return count, errors.New("cannot have empty points in MULTILINESTRING")
		case polyType:
			return count, errors.New("cannot have empty points in POLYGON")
		case mPolyType:
			return count, errors.New("cannot have empty points in MULTIPOLYGON")
		default:
			panic("unrechable")
		}
	}

	count++
	err = enc.encodePair(*lastEnc)
	if err != nil {
		return count, err
	}

	return count, enc.putc(')')
}

const (
	mpType byte = iota
	lsType
	mlType
	polyType
	mPolyType
)

func (enc *Encoder) encodeLines(lines [][][2]float64, removePointless bool, last int, gType byte) error {
	if !removePointless {
		last = len(lines) - 1
	}

	err := enc.putc('(')
	if err != nil {
		return err
	}

	for _, v := range lines[:last] {
		lastt, pointless := isPointlessPoints(v)
		if removePointless && pointless {
			continue
		}

		count, err := enc.encodePoints(v, removePointless, lastt, gType)
		if err != nil {
			return err
		}

		if (gType == polyType || gType == mPolyType) && count < 3 {
			return fmt.Errorf("not enough points for POLYGON %v", v)
		} else if count < 2 {
			return fmt.Errorf("not enough points for LINESTRING %v", v)
		}

		err = enc.putc(',')
		if err != nil {
			return err
		}
	}
	lastt, pointless := isPointlessPoints(lines[last])
	if removePointless && pointless {
		panic("the last element must always be poinfull, set by caller")
	}

	count, err := enc.encodePoints(lines[last], removePointless, lastt, gType)
	if err != nil {
		return err
	}
	if (gType == polyType || gType == mPolyType) && count < 3 {
		return fmt.Errorf("not enough points for POLYGON %v", lines[last])
	} else if count < 2 {
		return fmt.Errorf("not enough points for LINESTRING %v", lines[last])
	}

	return enc.putc(')')
}

func (enc *Encoder) encodePolys(polys [][][][2]float64, removePointless bool, last int) error {
	if !removePointless {
		last = len(polys) - 1
	}

	err := enc.putc('(')
	if err != nil {
		return err
	}

	for _, v := range polys[:last] {
		lastt, pointless := isPointlessLines(v)
		if removePointless && pointless {
			continue
		}
		err = enc.encodeLines(v, removePointless, lastt, mPolyType)
		if err != nil {
			return err
		}

		err = enc.putc(',')
		if err != nil {
			return err
		}
	}

	lastt, pointless := isPointlessLines(polys[last])
	if removePointless && pointless {
		panic("the last element must always be poinfull, set by caller")
	}

	err = enc.encodeLines(polys[last], removePointless, lastt, mPolyType)
	if err != nil {
		return err
	}

	return enc.putc(')')
}

func (enc *Encoder) encode(geo geom.Geometry, removePointless, recursive bool) error {

	switch g := geo.(type) {
	case [2]float64:
		if recursive && IsEmptyPoint(g) && removePointless {
			return nil
		}
		err := enc.puts("POINT ")
		if err != nil {
			return err
		}
		return enc.encodePoint(g)

	case geom.Pointer:
		if removePointless && recursive && (isNil(g) || IsEmptyPoint(g.XY())) {
			return nil
		}

		err := enc.puts("POINT ")
		if err != nil {
			return err
		}

		if isNil(g) {
			return enc.puts("EMPTY")
		}

		return enc.encodePoint(g.XY())

	case geom.MultiPointer:
		var mp [][2]float64

		if !isNil(g) {
			mp = g.Points()
		}

		last, isPointless := isPointlessPoints(mp)
		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("MULTIPOINT EMPTY")
			}
		}

		err := enc.puts("MULTIPOINT ")
		if err != nil {
			return err
		}

		_, err = enc.encodePoints(mp, removePointless, last, mpType)
		return err

	case geom.LineStringer:
		var mp [][2]float64

		if !isNil(g) {
			mp = g.Verticies()
		}

		last, isPointless := isPointlessPoints(mp)
		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("LINESTRING EMPTY")
			}
		}

		err := enc.puts("LINESTRING ")
		if err != nil {
			return err
		}

		count, err := enc.encodePoints(mp, removePointless, last, lsType)
		if err != nil {
			return err
		}
		if count < 2 {
			return fmt.Errorf("not enough points for LINESTRING %v", mp)
		}

		return nil

	case geom.MultiLineStringer:
		var lines [][][2]float64

		if !isNil(g) {
			lines = g.LineStrings()
		}

		last, isPointless := isPointlessLines(lines)
		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("MULTILINESTRING EMPTY")
			}
		}

		err := enc.puts("MULTILINESTRING ")
		if err != nil {
			return err
		}

		return enc.encodeLines(lines, removePointless, last, mlType)

	case geom.Polygoner:
		var lines [][][2]float64

		if !isNil(g) {
			lines = g.LinearRings()
		}

		last, isPointless := isPointlessLines(lines)
		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("POLYGON EMPTY")
			}
		}

		err := enc.puts("POLYGON ")
		if err != nil {
			return err
		}

		return enc.encodeLines(lines, removePointless, last, polyType)

	case geom.MultiPolygoner:
		var polys [][][][2]float64

		if !isNil(g) {
			polys = g.Polygons()
		}

		last, isPointless := isPointlessPolys(polys)
		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("MULTIPOLYGON EMPTY")
			}
		}

		err := enc.puts("MULTIPOLYGON ")
		if err != nil {
			return err
		}

		return enc.encodePolys(polys, removePointless, last)

	case geom.Collectioner:
		var geoms []geom.Geometry

		if !isNil(g) {
			geoms = g.Geometries()
		}

		last := -1
		isPointless := true

		for i := len(geoms) - 1; i >= 0; i-- {
			pl, err := isPointlessGeo(geoms[i])
			if err != nil {
				return err
			}
			if !pl {
				last = i
				isPointless = false
				break
			}
		}

		if isPointless {
			if removePointless && recursive {
				return nil
			} else if removePointless {
				return enc.puts("GEOMETRYCOLLECTION EMPTY")
			}
		}

		err := enc.puts("GEOMETRYCOLLECTION ")
		if err != nil {
			return err
		}

		err = enc.putc('(')
		if err != nil {
			return err
		}

		for _, v := range geoms[:last] {
			pl, err := isPointlessGeo(v)
			if err != nil {
				return err
			}
			if removePointless && pl {
				continue
			}
			err = enc.encode(v, removePointless, true)
			if err != nil {
				return err
			}

			err = enc.putc(',')
			if err != nil {
				return err
			}
		}

		pl, err := isPointlessGeo(geoms[last])
		if err != nil {
			return err
		}
		if removePointless && pl {
			panic("the last element must always be poinfull, set by caller")
		}

		err = enc.encode(geoms[last], removePointless, true)
		if err != nil {
			return err
		}

		return enc.putc(')')

	default:
		return fmt.Errorf("unknown geometry: %T", geo)
	}
}

func (enc *Encoder) Encode(geo geom.Geometry, removePointless bool) error {
	return enc.encode(geo, removePointless, false)
}
