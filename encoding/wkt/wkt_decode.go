package wkt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"unicode"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/cmp"
)

type Decoder struct {
	src                        *bufio.Reader
	row, col, lastRow, lastCol int
}

func (d *Decoder) peekByte() (byte, error) {
	arr, err := d.src.Peek(1)
	return arr[0], err
}

func (d *Decoder) readByte() (byte, error) {
	b, err := d.src.ReadByte()
	if err == io.EOF {
		return b, d.syntaxErr("UNEXPECTED", "eof")
	}

	d.lastCol = d.col
	d.lastRow = d.row
	if b == '\n' || b == '\r' {
		d.row++
		d.col = 0
	} else {
		d.col++
	}

	return b, err
}

func (d *Decoder) unreadByte() error {
	d.row = d.lastRow
	d.col = d.lastCol

	return d.src.UnreadByte()
}

// readWhitespace eats up the whitespace and returns
// true iff any characters were read
func (d *Decoder) readWhitespace() (bool, error) {
	var b byte
	var err error
	read := false
	for b, err = d.readByte(); unicode.IsSpace(rune(b)) && err == nil; b, err = d.readByte() {
		read = true
	}

	if err != nil {
		return read, err
	}

	d.unreadByte()

	return read, nil
}

func (d *Decoder) expected(chars string) error {
	d.unreadByte()
	b, err := d.readByte()
	if err != nil {
		// this shouldn't happen
		return err
	}

	return d.syntaxErr(
		"expected",
		"one of `%q` got %q",
		chars,
		b,
	)
}

func (d *Decoder) syntaxErr(errType string, format string, v ...interface{}) error {
	return ErrSyntax{
		Line:  d.row,
		Char:  d.col,
		Type:  errType,
		Issue: fmt.Sprintf(format, v...),
	}
}

func (d *Decoder) readToken(inTok func(byte) bool) ([]byte, error) {
	token := []byte{}

	var err error
	var b byte

	for {
		b, err = d.readByte()
		if err != nil {
			return nil, err
		}

		if !inTok(b) {
			d.unreadByte()
			break
		}
		token = append(token, b)
	}
	return token, nil
}

func (d *Decoder) readInteger() (int64, error) {
	isNumeric := func(b byte) bool {
		return (b >= '0' && b <= '9') ||
			b == '-'
	}
	token, err := d.readToken(isNumeric)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(string(token), 10, 64)
	if err != nil {
		return 0, d.syntaxErr("float", "cannot parse %q", token)
	}
	return i, nil
}

func (d *Decoder) readFloat() (float64, error) {
	isNumeric := func(b byte) bool {
		return (b >= '0' && b <= '9') ||
			b == '-' ||
			b == '.' ||
			// b == ',' || // technically part of the spec,
			// but even postgis does not support it
			b == 'E' ||
			b == 'e'
	}
	token, err := d.readToken(isNumeric)

	if err != nil {
		return 0, err
	}

	ret, err := strconv.ParseFloat(string(token), 64)
	if err != nil {
		return 0, d.syntaxErr("float", "cannot parse %q", token)
	}
	return ret, nil
}

// readPoint reads a space separated tuple of floats, the inside
// of a wkt POINT
func (d *Decoder) readPoint() (pt [2]float64, err error) {
	pt[0], err = d.readFloat()
	if err != nil {
		return pt, err
	}

	// we need white space here
	didRead, err := d.readWhitespace()
	if err != nil {
		return pt, err
	}
	if !didRead {
		return pt, d.expected("WHITESPACE")
	}

	pt[1], err = d.readFloat()

	return pt, err
}

func (d *Decoder) readPoints() (pts [][2]float64, err error) {
	b, err := d.readByte()
	if err != nil {
		return nil, err
	}
	if b != '(' {
		return nil, d.expected("(")
	}
	_, err = d.readWhitespace()
	if err != nil {
		return nil, err
	}

	b, err = d.readByte()
	if err != nil {
		return nil, err
	}
	if b == ')' {
		return pts, nil
	}
	d.unreadByte()

	for {
		pt, err := d.readPoint()
		if err != nil {
			return nil, err
		}
		pts = append(pts, pt)

		_, err = d.readWhitespace()
		if err != nil {
			return nil, err
		}

		b, err := d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			_, err = d.readWhitespace()
			if err != nil {
				return nil, err
			}
			continue
		case ')':
			return pts, nil
		default:
			return nil, d.expected(",)")
		}
	}
}

func (d *Decoder) readTag() (string, error) {

	isAlpha := func(b byte) bool {
		return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
	}

	token := []byte{}

	var err error
	var b byte

	_, err = d.readWhitespace()
	if err != nil {
		return "", err
	}

	for b, err = d.readByte(); isAlpha(b) && err == nil; b, err = d.readByte() {
		// to lower
		if b < 'a' {
			b += 'a' - 'A'
		}
		token = append(token, b)
	}

	if err != nil {
		return "", err
	}

	d.unreadByte()

	return string(token), nil
}

func (d *Decoder) readLines() ([][][2]float64, error) {
	b, err := d.readByte()
	if err != nil {
		return nil, err
	}

	if b != '(' {
		return nil, d.expected("(")
	}

	_, err = d.readWhitespace()
	if err != nil {
		return nil, err
	}

	b, err = d.readByte()
	if err != nil {
		return nil, err
	}
	if b == ')' {
		return nil, nil
	}
	d.unreadByte()

	lines := [][][2]float64{}

	for {
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		lines = append(lines, pts)

		_, err = d.readWhitespace()
		if err != nil {
			return nil, err
		}

		b, err = d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			_, err = d.readWhitespace()
			if err != nil {
				return nil, err
			}

			continue
		case ')':
			return lines, nil
		default:
			return nil, d.expected(",)")
		}
	}
}

func (d *Decoder) readPolys() ([][][][2]float64, error) {
	b, err := d.readByte()
	if err != nil {
		return nil, err
	}

	if b != '(' {
		return nil, d.expected("(")
	}

	_, err = d.readWhitespace()
	if err != nil {
		return nil, err
	}

	b, err = d.readByte()
	if err != nil {
		return nil, err
	}
	if b == ')' {
		return nil, nil
	}
	d.unreadByte()

	polys := [][][][2]float64{}
	for {
		lines, err := d.readLines()
		if err != nil {
			return nil, err
		}

		polys = append(polys, lines)

		_, err = d.readWhitespace()
		if err != nil {
			return nil, err
		}

		b, err = d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			_, err = d.readWhitespace()
			if err != nil {
				return nil, err
			}
			continue
		case ')':
			return polys, nil
		default:
			return nil, d.expected(",)")
		}
	}
}

func (d *Decoder) readGeometry() (geom.Geometry, error) {
	var srid geom.Srid

ReadTag:
	tag, err := d.readTag()
	if err != nil {
		return nil, err
	}

	_, err = d.readWhitespace()
	if err != nil {
		return nil, err
	}

	switch tag {
	case "srid":
		if srid != 0 {
			return nil, fmt.Errorf("srid already set")
		}
		// we are in the srid, so we expect: "srid=\d+;..."
		_, err = d.readWhitespace()
		if err != nil {
			return nil, err
		}
		b, err := d.readByte()
		if err != nil {
			return nil, err
		}
		if b != '=' {
			_ = d.unreadByte()
			return nil, d.expected("=")
		}
		_, _ = d.readWhitespace()
		i, err := d.readInteger()
		if err != nil {
			return nil, err
		}
		if i <= 0 {
			return nil, fmt.Errorf("srid should be greater than or equal to 0; not %v", i)
		}
		srid = geom.Srid(i)
		_, _ = d.readWhitespace()
		b, err = d.readByte()
		if err != nil {
			return nil, err
		}
		if b != ';' {
			_ = d.unreadByte()
			return nil, d.expected(";")
		}

		// Let's do the tag processing again.
		goto ReadTag

	case "point":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		switch len(pts) {
		case 0:
			return nil, d.syntaxErr("POINT", "cannot be empty")
		case 1:
			if srid != 0 {
				return geom.PointS{
					Srid: srid,
					Xy:   pts[0],
				}, nil
			}
			return geom.Point(pts[0]), nil
		default:
			return nil, d.syntaxErr("POINT", "too many points %d", len(pts))
		}

	case "multipoint":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}
		if srid != 0 {
			return geom.MultiPointS{
				Srid: srid,
				Mp:   pts,
			}, nil
		}
		return geom.MultiPoint(pts), nil

	case "linestring":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		if len(pts) < 2 {
			return nil, d.syntaxErr("LINESTRING", "not enough points %d", len(pts))
		}

		if srid != 0 {
			return geom.LineStringS{
				Srid: srid,
				Ls:   pts,
			}, nil
		}
		return geom.LineString(pts), nil

	case "multilinestring":
		lines, err := d.readLines()
		if err != nil {
			return nil, err
		}

		if len(lines) < 1 {
			return nil, d.syntaxErr("MULTILINESTRING", "not enough lines %d", len(lines))
		}

		for i, v := range lines {
			if len(v) < 2 {
				return nil, d.syntaxErr("MULTILINESTRING", "not enough points in LINESTRING[%d], %d", i, len(v))
			}
		}

		if srid != 0 {
			return geom.MultiLineStringS{
				Srid: srid,
				Mls:  lines,
			}, nil
		}
		return geom.MultiLineString(lines), nil

	case "polygon":
		lines, err := d.readLines()
		if err != nil {
			return nil, err
		}

		if len(lines) < 1 {
			return nil, d.syntaxErr("POLYGON", "not enough lines %d", len(lines))
		}

		for i, v := range lines {
			if len(v) < 4 {
				return nil, d.syntaxErr("POLYGON", "not enough points in linear-ring[%d], %d", i, len(v))
			}

			// part of the spec
			if !cmp.PointEqual(v[0], v[len(v)-1]) {
				return nil, d.syntaxErr("POLYGON", "linear-ring[%d] not closed", i)
			}

			// part of go-spatial/geom convention
			lines[i] = v[:len(v)-1]
		}

		if srid != 0 {
			return geom.PolygonS{
				Srid: srid,
				Pol:  lines,
			}, nil
		}
		return geom.Polygon(lines), nil

	case "multipolygon":
		polys, err := d.readPolys()
		if err != nil {
			return nil, err
		}

		if len(polys) < 1 {
			return nil, d.syntaxErr("MULTIPOLYGON", "not enough polygons %d", len(polys))
		}

		for ii, vv := range polys {
			for i, v := range vv {
				if len(v) < 4 {
					return nil, d.syntaxErr("MULTIPOLYGON", "not enough points in polygon[%d] linear-ring[%d], %d", ii, i, len(v))
				}

				// part of the spec
				if !cmp.PointEqual(v[0], v[len(v)-1]) {
					return nil, d.syntaxErr("MULTIPOLYGON", "polygon[%d] linear-ring[%v] not closed", i, ii)
				}

				// part of go-spatial/geom convention
				polys[ii][i] = v[:len(v)-1]
			}
		}

		if srid != 0 {
			return geom.MultiPolygonS{
				Srid:         srid,
				MultiPolygon: polys,
			}, nil
		}
		return geom.MultiPolygon(polys), err
	case "geometrycollection":
		b, err := d.readByte()
		if err != nil {
			return nil, err
		}
		if b != '(' {
			return nil, d.expected("(")
		}
		_, err = d.readWhitespace()
		if err != nil {
			return nil, err
		}

		geoms := geom.Collection{}

		for b, err = d.readByte(); b != ')' && err == nil; b, err = d.readByte() {
			d.unreadByte()

			geo, err := d.readGeometry()
			if err != nil {
				return nil, err
			}
			geoms = append(geoms, geo)

			_, err = d.readWhitespace()
			if err != nil {
				return nil, err
			}

			b, err := d.readByte()
			if err != nil {
				return nil, err
			}

			switch b {
			case ')':
				d.unreadByte()
			case ',':
				//noop
				_, err = d.readWhitespace()
				if err != nil {
					return nil, err
				}
			default:
				return nil, d.expected(",)")
			}

		}

		if err != nil {
			return nil, err
		}

		if len(geoms) < 1 {
			return nil, d.syntaxErr("GEOMETRYCOLLECTION", "not enough geometries %d", len(geoms))
		}

		if b != ')' {
			panic("unreacheable")
		}

		if srid != 0 {
			return geom.CollectionS{
				Srid:       srid,
				Collection: geoms,
			}, nil
		}
		return geoms, nil

	default:
		return nil, d.syntaxErr("GEOMETRY", "unknown type %q", tag)
	}
}

func (d *Decoder) Decode() (geom.Geometry, error) {
	return d.readGeometry()
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		src: bufio.NewReader(r),
	}
}
