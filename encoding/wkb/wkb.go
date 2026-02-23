// Package wkb is for decoding ESRI's Well Known Binary (WKB) format for OGC geometry (WKBGeometry)
// specification at http://edndoc.esri.com/arcsde/9.1/general_topics/wkb_representation.htm
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
func DecodeBytes(b []byte) (geom.Geometry, error) { return Decode(bytes.NewReader(b)) }

// Decode will attempt to decode a geometry encode as WKB (or EWKB) into a geom.Geometry.
func Decode(r io.Reader) (geo geom.Geometry, err error) {

	var (
		srid    uint32
		hasSRID bool
	)
	bom, typ, err := decode.ByteOrderType(r)
	if err != nil {
		return nil, err
	}
	if typ&consts.WKBSRID == consts.WKBSRID {
		typ -= consts.WKBSRID
		srid, err = decode.SRID(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode srid: %w", err)
		}
		hasSRID = true
	}
	switch typ {
	case Point:
		pt, err := decode.Point(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode point: %w", err)
		}
		if hasSRID {
			return geom.PointS{
				Srid: geom.Srid(srid),
				Xy:   pt,
			}, nil
		}
		return pt, nil
	case MultiPoint:
		mpt, err := decode.MultiPoint(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode multi point: %w", err)
		}
		if hasSRID {
			return geom.MultiPointS{
				Srid: geom.Srid(srid),
				Mp:   mpt,
			}, nil
		}
		return mpt, nil
	case LineString:
		ln, err := decode.LineString(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode linestring: %w", err)
		}
		if hasSRID {
			return geom.LineStringS{
				Srid: geom.Srid(srid),
				Ls:   ln,
			}, nil
		}
		return ln, nil
	case MultiLineString:
		mln, err := decode.MultiLineString(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode multilinestring: %w", err)
		}
		if hasSRID {
			return geom.MultiLineStringS{
				Srid: geom.Srid(srid),
				Mls:  mln,
			}, nil
		}
		return mln, err
	case Polygon:
		pl, err := decode.Polygon(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode polygon: %w", err)
		}
		if hasSRID {
			return geom.PolygonS{
				Srid: geom.Srid(srid),
				Pol:  pl,
			}, nil
		}
		return pl, nil
	case MultiPolygon:
		mpl, err := decode.MultiPolygon(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode multi polygon: %w", err)
		}
		if hasSRID {
			return geom.MultiPolygonS{
				Srid:         geom.Srid(srid),
				MultiPolygon: mpl,
			}, nil
		}
		return mpl, nil
	case Collection:
		col, err := decode.Collection(r, bom)
		if err != nil {
			return nil, fmt.Errorf("failed to decode collection: %w", err)
		}
		if hasSRID {
			return geom.CollectionS{
				Srid:       geom.Srid(srid),
				Collection: col,
			}, nil
		}
		return col, nil
	default:
		return nil, ErrUnknownGeometryType{typ}
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
