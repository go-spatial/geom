package wkt

import (
	"bufio"
	"fmt"
	"io"
	"strconv"

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
		return b, d.syntaxErr("unexpected eof")
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

func (d *Decoder) readWhitespace() error {
	isSpace := func(b byte) bool {
		return b == ' ' ||
			b == '\t' ||
			b == '\n' ||
			b == '\f' ||
			b == '\r' ||
			b == '\v'
	}

	var b byte
	var err error
	read := false
	for b, err = d.readByte(); isSpace(b) && err == nil; b, err = d.readByte() {
		read = true
	}

	if err != nil {
		return err
	}

	d.unreadByte()
	if !read {
		return d.syntaxErr("expected whitespace got %q", b)
	}

	return nil
}

func (d *Decoder) expected(chars string) error {
	d.unreadByte()
	b, err := d.readByte()
	if err != nil {
		// this shouldn't happen
		return err
	}

	return fmt.Errorf("syntax error (%d:%d): expected one of %q got %q",
		d.row+1,
		d.col+1,
		chars,
		b)
}

func (d *Decoder) syntaxErr(format string, v ...interface{}) error {
	return fmt.Errorf("syntax error (%d:%d): "+format,
		append([]interface{}{d.row + 1, d.col + 1}, v...)...)
}

func (d *Decoder) readFloat() (float64, error) {
	isNumeric := func(b byte) bool {
		return (b >= '0' && b <= '9') ||
			b == '-' ||
			b == '.' ||
			// b == ',' || // technically part of the spec,
			// but even postgis does not support it
			b == 'E'
	}

	token := []byte{}

	var err error
	var b byte

	for b, err = d.readByte(); isNumeric(b) && err == nil; b, err = d.readByte() {
		token = append(token, b)
	}

	if err != nil {
		return 0, err
	}

	d.unreadByte()

	ret, err := strconv.ParseFloat(string(token), 64)
	if err != nil {
		return 0, d.syntaxErr("cannot parse float %q", token)
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
	err = d.readWhitespace()
	if err != nil {
		return pt, err
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
	d.readWhitespace()

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

		d.readWhitespace()

		b, err := d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			d.readWhitespace()
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


	d.readWhitespace()

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

		d.readWhitespace()
		b, err = d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			d.readWhitespace()
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

	d.readWhitespace()

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

		d.readWhitespace()
		b, err = d.readByte()
		if err != nil {
			return nil, err
		}

		switch b {
		case ',':
			d.readWhitespace()
			continue
		case ')':
			return polys, nil
		default:
			return nil, d.expected(",)")
		}
	}
}

func (d *Decoder) readGeometry() (geom.Geometry, error) {
	tag, err := d.readTag()
	if err != nil {
		return nil, err
	}

	d.readWhitespace()

	switch tag {
	case "point":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		switch len(pts) {
		case 0:
			return nil, d.syntaxErr("POINT cannot be empty")
		case 1:
			return geom.Point(pts[0]), nil
		default:
			return nil, d.syntaxErr("too many points in POINT, %d", len(pts))
		}

	case "multipoint":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		return geom.MultiPoint(pts), nil

	case "linestring":
		pts, err := d.readPoints()
		if err != nil {
			return nil, err
		}

		if len(pts) < 2 {
			return nil, d.syntaxErr("not enough points in LINESTRING, %d", len(pts))
		}

		return geom.LineString(pts), nil

	case "multilinestring":
		lines, err := d.readLines()
		if err != nil {
			return nil, err
		}

		if len(lines) < 1 {
			return nil, d.syntaxErr("not enough lines in MULTILINESTRING, %d", len(lines))
		}

		for i, v := range lines {
			if len(v) < 2 {
				return nil, d.syntaxErr("not enough points in MULTILINESTRING[%d], %d", i, len(v))
			}
		}

		return geom.MultiLineString(lines), nil

	case "polygon":
		lines, err := d.readLines()
		if err != nil {
			return nil, err
		}

		if len(lines) < 1 {
			return nil, d.syntaxErr("not enough lines in POLYGON, %d", len(lines))
		}

		for i, v := range lines {
			if len(v) < 4 {
				return nil, d.syntaxErr("not enough points in POLYGON[%d], %d", i, len(v))
			}

			// part of the spec
			if !cmp.PointEqual(v[0], v[len(v)-1]) {
				return nil, d.syntaxErr("first and last point of POLYGON[%d] not equal", i)
			}

			// part of go-spatial/geom convention
			lines[i] = v[:len(v)-1]
		}

		return geom.Polygon(lines), nil

	case "multipolygon":
		polys, err := d.readPolys()
		if err != nil {
			return nil, err
		}

		if len(polys) < 1 {
			return nil, d.syntaxErr("not enough polys in MULTIPOLYGON, %d", len(polys))
		}

		for ii, vv := range polys {
			for i, v := range vv {
				if len(v) < 4 {
					return nil, d.syntaxErr("not enough points in MULTIPOLYGON[%d][%d], %d", ii, i, len(v))
				}

				// part of the spec
				if !cmp.PointEqual(v[0], v[len(v)-1]) {
					return nil, d.syntaxErr("first and last point of POLYGON[%d] not equal", i)
				}

				// part of go-spatial/geom convention
				polys[ii][i] = v[:len(v)-1]
			}
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
		d.readWhitespace()

		geoms := geom.Collection{}

		for b, err = d.readByte(); b != ')' && err == nil; b, err = d.readByte() {
			d.unreadByte()

			geo, err := d.readGeometry()
			if err != nil {
				return nil, err
			}
			geoms = append(geoms, geo)

			d.readWhitespace()

			b, err := d.readByte()
			if err != nil {
				return nil, err
			}

			switch b {
			case ')':
				d.unreadByte()
			case ',':
				//noop
				d.readWhitespace()
			default:
				return nil, d.expected(",)")
			}

		}

		if err != nil {
			return nil, err
		}

		if len(geoms) < 1 {
			return nil, d.syntaxErr("not enough geoms in GEOMETRYCOLLECTION, %d", len(geoms))
		}

		if b != ')' {
			panic("unreacheable")
		}

		return geoms, nil

	default:
		return nil, d.syntaxErr("unknown geometry type %q", tag)
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

