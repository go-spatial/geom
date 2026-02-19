// Package wkb is for decoding ESRI's Well Known Binary (WKB) format for OGC geometry (WKBGeometry)
// sepcification at http://edndoc.esri.com/arcsde/9.1/general_topics/wkb_representation.htm
// There are a few types supported by the specification. Each general type is in it's own file.
// So, to find the implementation of Point (and MultiPoint) it will be located in the point.go
// file. Each of the basic type here adhere to the geom.Geometry interface. So, a wkb point
// is, also, a geom.Point
package wkb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkb/internal/consts"
	"github.com/go-spatial/geom/encoding/wkb/internal/decode"
	"github.com/go-spatial/geom/encoding/wkb/internal/encode"
)

type ErrUnknownGeometryType struct {
	Typ uint32
}

func (e ErrUnknownGeometryType) Error() string {
	return fmt.Sprintf("Unknown Geometry Type %v", e.Typ)
}

// geometry types
// http://edndoc.esri.com/arcsde/9.1/general_topics/wkb_representation.htm
const (
	Point           = consts.Point
	LineString      = consts.LineString
	Polygon         = consts.Polygon
	MultiPoint      = consts.MultiPoint
	MultiLineString = consts.MultiLineString
	MultiPolygon    = consts.MultiPolygon
	Collection      = consts.Collection
)

// DecodeBytes will attempt to decode a geometry encoded as WKB into a geom.Geometry.
func DecodeBytes(b []byte) (geom.Geometry, error) {
	g, _, e := DecodeWSRID(bytes.NewReader(b))
	return g, e
}
func DecodeBytesWSRID(b []byte) (geom.Geometry, uint32, error) {
	return DecodeWSRID(bytes.NewReader(b))
}

// Decode will attempt to decode a geometry encoded as WKB into a geom.Geometry.
func Decode(r io.Reader) (geo geom.Geometry, err error) {
	g, _, e := DecodeWSRID(r)
	return g, e
}
func DecodeWSRID(r io.Reader) (geo geom.Geometry, srid uint32, err error) {

	bom, typ, err := decode.ByteOrderType(r)
	if err != nil {
		return nil, srid, err
	}
	if typ&consts.WKBSRID == consts.WKBSRID {
		typ -= consts.WKBSRID
		srid, err = decode.SRID(r, bom)
		if err != nil {
			return nil, srid, fmt.Errorf("failed to decode srid: %w", err)
		}
	}
	switch typ {
	case Point:
		pt, err := decode.Point(r, bom)
		return geom.Point(pt), srid, err
	case MultiPoint:
		mpt, err := decode.MultiPoint(r, bom)
		return geom.MultiPoint(mpt), srid, err
	case LineString:
		ln, err := decode.LineString(r, bom)
		return geom.LineString(ln), srid, err
	case MultiLineString:
		mln, err := decode.MultiLineString(r, bom)
		return geom.MultiLineString(mln), srid, err
	case Polygon:
		pl, err := decode.Polygon(r, bom)
		return geom.Polygon(pl), srid, err
	case MultiPolygon:
		mpl, err := decode.MultiPolygon(r, bom)
		return geom.MultiPolygon(mpl), srid, err
	case Collection:
		col, err := decode.Collection(r, bom)
		return col, srid, err
	default:
		return nil, srid, ErrUnknownGeometryType{typ}
	}
}

func EncodeBytes(g geom.Geometry) (bs []byte, err error) {
	buff := new(bytes.Buffer)
	if err = Encode(buff, g); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func Encode(w io.Writer, g geom.Geometry) error {
	return EncodeWithByteOrder(binary.LittleEndian, 0, w, g)
}

func EncodeBytesSRID(srid uint32, g geom.Geometry) (bs []byte, err error) {
	buff := new(bytes.Buffer)
	if err = EncodeWithByteOrder(binary.LittleEndian, srid, buff, g); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func EncodeWithByteOrder(byteOrder binary.ByteOrder, srid uint32, w io.Writer, g geom.Geometry) error {
	en := encode.Encoder{W: w, ByteOrder: byteOrder, SRID: srid}
	en.Geometry(g)
	return en.Err()
}
