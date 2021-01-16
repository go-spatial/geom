// Package geojson implements encoding and decoding of GeoJSON as
// defined in [RFC 7946](https://tools.ietf.org/html/rfc7946). The
// mapping between JSON and geom Geometry values are described in
// the documentation for the Marshal and Unmarshal functions.
//
// At current this pacakge only supports 2D Geometries unless stated
// otherwise by the documentation of the Marshal and Unmarshal functions
package geojson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding"
)

var (
	ErrUnknownFeatureType = fmt.Errorf("unknown feature type")
)

type GeoJSONType string

const (
	PointType              GeoJSONType = "Point"
	MultiPointType         GeoJSONType = "MultiPoint"
	LineStringType         GeoJSONType = "LineString"
	MultiLineStringType    GeoJSONType = "MultiLineString"
	PolygonType            GeoJSONType = "Polygon"
	MultiPolygonType       GeoJSONType = "MultiPolygon"
	GeometryCollectionType GeoJSONType = "GeometryCollection"
	FeatureType            GeoJSONType = "Feature"
	FeatureCollectionType  GeoJSONType = "FeatureCollection"
)

// Marshal returns the geojson encoding of the geojson.Feature, geojson.FeatureCollection, or a geom.Geometry.
//
// If Marshal is given a geom.Geometry, this geometry will be wrapped in a geojson.Feature, with no properties
// or and ID.
// If something other then the above is passed in the system will return an geom.ErrUnknownGeometry type.
// Values in the property map are marshaled according to the type-dependent default encoding as defined
// by the go's encoding/json package.
//
func Marshal(v interface{}) ([]byte, error) {
	switch g := v.(type) {
	case Feature:
		return json.Marshal(g)
	case Geometry:
		return json.Marshal(Feature{Geometry: g})
	case FeatureCollection:
		return json.Marshal(g)

	default:
		if isGeomGeometry(v) {
			return json.Marshal(Feature{Geometry: Geometry{g}})
		}
		if s, ok := isGeomGeometrySlice(v); ok {
			fc := FeatureCollection{
				Features: make([]Feature, 0, len(s)),
			}
			for _, g := range s {
				if !isGeomGeometry(g) {
					return nil, fmt.Errorf("in geom.Geometry slice, %w", geom.ErrUnknownGeometry{g})
				}
				fc.Features = append(fc.Features, Feature{Geometry: Geometry{g}})
			}
			return json.Marshal(fc)
		}
		return nil, geom.ErrUnknownGeometry{g}
	}
}

// MarshalIndent is like Marshal but applies Indent to format the output
// Each JSON element is the output will begin on a new line beginning with prefix
// followed by one or more copies of indent according to indentation nesting.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	b, err := Marshal(v)
	if err != nil {
		return nil, err
	}
	var buff bytes.Buffer
	if err = json.Indent(&buff, b, prefix, indent); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

// Unmarshal parses the GeoJSON-encoded data and returns the result or an error.
// The result can be either a geojson.Features or geojson.FeatureCollection.
// If the encoded data is not one of the above then function will return the
// error json.InvalidUnmarshalError.
func Unmarshal(data []byte) (feature interface{}, err error) {
	var typeMessage struct {
		Type string `json:"type"`
	}
	if err = json.Unmarshal(data, &typeMessage); err != nil {
		return nil, err
	}
	switch strings.ToLower(typeMessage.Type) {
	case "feature":
		var f Feature
		if err = json.Unmarshal(data, &f); err != nil {
			return nil, err
		}
		return f, err
	case "featurecollection":
		var fc FeatureCollection
		if err = json.Unmarshal(data, &fc); err != nil {
			return nil, err
		}
		return fc, nil
	}
	return nil, ErrUnknownFeatureType
}

// isGeomGeometry will check to see if v is type that fulfills one of the
// geom Geometry Type interfaces. E.G. geom.Pointer, geom.MultiPointer,
// etc...
func isGeomGeometry(v interface{}) bool {
	switch v.(type) {
	case geom.Pointer:
		return true
	case geom.MultiPointer:
		return true
	case geom.LineStringer:
		return true
	case geom.MultiLineStringer:
		return true
	case geom.Polygoner:
		return true
	case geom.MultiPolygoner:
		return true
	case geom.Collectioner:
		return true
	default:
		return false
	}
}

// isGeomGeometrySlice will check to see if v is slice type that fulfills one of the
// geom Geometry Type interfaces. E.G. geom.Pointer, geom.MultiPointer, includeing
// geom.Geometry
// etc...
//
// This function does not do a deep check of the values provided, if the type is
// []geom.Geometry
func isGeomGeometrySlice(v interface{}) ([]geom.Geometry, bool) {
	switch g := v.(type) {
	case []geom.Geometry:
		return g, true
	case []geom.Pointer:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.MultiPointer:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.LineStringer:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.MultiLineStringer:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.Polygoner:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.MultiPolygoner:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	case []geom.Collectioner:
		gg := make([]geom.Geometry, len(g))
		for i := range g {
			gg[i] = g[i]
		}
		return gg, true
	default:
		return nil, false
	}
}

// Geometry wraps a geom Geometry so that it can be encoded as a GeoJSON
// feature
type Geometry struct {
	geom.Geometry
}

func (geo Geometry) MarshalJSON() ([]byte, error) {
	type coordinates struct {
		Type   GeoJSONType `json:"type"`
		Coords interface{} `json:"coordinates,omitempty"`
	}
	type collection struct {
		Type       GeoJSONType `json:"type"`
		Geometries []Geometry  `json:"geometries,omitempty"`
	}

	switch g := geo.Geometry.(type) {
	case geom.Pointer:
		return json.Marshal(coordinates{
			Type:   PointType,
			Coords: g.XY(),
		})

	case geom.MultiPointer:
		return json.Marshal(coordinates{
			Type:   MultiPointType,
			Coords: g.Points(),
		})

	case geom.LineStringer:
		return json.Marshal(coordinates{
			Type:   LineStringType,
			Coords: g.Vertices(),
		})

	case geom.MultiLineStringer:
		return json.Marshal(coordinates{
			Type:   MultiLineStringType,
			Coords: g.LineStrings(),
		})

	case geom.Polygoner:
		ps := g.LinearRings()
		closePolygon(ps)

		return json.Marshal(coordinates{
			Type:   PolygonType,
			Coords: ps,
		})

	case geom.MultiPolygoner:
		ps := g.Polygons()

		// iterate through the polygons making sure they're closed
		for i := range ps {
			closePolygon(geom.Polygon(ps[i]))
		}

		return json.Marshal(coordinates{
			Type:   MultiPolygonType,
			Coords: ps,
		})

	case geom.Collectioner:
		gs := g.Geometries()

		var geos = make([]Geometry, 0, len(gs))
		for _, gg := range gs {
			geos = append(geos, Geometry{gg})
		}

		return json.Marshal(collection{
			Type:       GeometryCollectionType,
			Geometries: geos,
		})

	default:
		return nil, geom.ErrUnknownGeometry{g}
	}
}

// featureType allows the GeoJSON type for Feature to be automatically set during json Marshalling
// which avoids the user from accidentally setting the incorrect GeoJSON type.
type featureType struct{}

func (_ featureType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + FeatureType + `"`), nil
}
func (fc *featureType) UnmarshalJSON(b []byte) error { return nil }

// Feature represents as geojson feature
type Feature struct {
	Type featureType `json:"type"`
	ID   *uint64     `json:"id,omitempty"`
	// Geometry can be null
	Geometry Geometry `json:"geometry"`
	// Properties can be null
	Properties map[string]interface{} `json:"properties"`
}

// featureCollectionType allows the GeoJSON type for Feature to be automatically set during json Marshalling
// which avoids the user from accidentally setting the incorrect GeoJSON type.
type featureCollectionType struct{}

func (_ featureCollectionType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + FeatureCollectionType + `"`), nil
}
func (fc *featureCollectionType) UnmarshalJSON(b []byte) error { return nil }

// FeatureCollection describes a geoJSON collection feature
type FeatureCollection struct {
	Type     featureCollectionType `json:"type"`
	Features []Feature             `json:"features"`
}

// closePolygon will ensure that the last point of a polygon is the same as the first
// point of the polygon. geom Polygon rings are not "closed", however geoJSON polygon
// ring are.
func closePolygon(p geom.Polygon) {
	for i := range p {
		if len(p[i]) == 0 {
			continue
		}

		// check if the first point and the last point are the same
		// if they're not, make a copy of the first point and add it as the last position
		if p[i][0] != p[i][len(p[i])-1] {
			p[i] = append(p[i], p[i][0])
		}
	}
}

// UnmarshalJSON will attempt to unmarshal the given bytes into a GeoJSON object.
// It can produce a varity of json Marshaling errors or
// encoding.InvalidGeometry if the geometry type in unsupported
func (geo *Geometry) UnmarshalJSON(b []byte) error {
	var geojsonMap map[string]*json.RawMessage
	if err := json.Unmarshal(b, &geojsonMap); err != nil {
		return err
	}

	var geomType GeoJSONType
	if err := json.Unmarshal(*geojsonMap["type"], &geomType); err != nil {
		return err
	}
	switch geomType {
	case PointType:
		var pt geom.Point
		if err := json.Unmarshal(*geojsonMap["coordinates"], &pt); err != nil {
			return err
		}
		geo.Geometry = pt
		return nil
	case PolygonType:
		var poly geom.Polygon
		if err := json.Unmarshal(*geojsonMap["coordinates"], &poly); err != nil {
			return err
		}
		geo.Geometry = poly
		return nil
	case LineStringType:
		var ls geom.LineString
		if err := json.Unmarshal(*geojsonMap["coordinates"], &ls); err != nil {
			return err
		}
		geo.Geometry = ls
		return nil
	case MultiPointType:
		var mp geom.MultiPoint
		if err := json.Unmarshal(*geojsonMap["coordinates"], &mp); err != nil {
			return err
		}
		geo.Geometry = mp
		return nil
	case MultiLineStringType:
		var ml geom.MultiLineString
		if err := json.Unmarshal(*geojsonMap["coordinates"], &ml); err != nil {
			return err
		}
		geo.Geometry = ml
		return nil
	case MultiPolygonType:
		var mp geom.MultiPolygon
		if err := json.Unmarshal(*geojsonMap["coordinates"], &mp); err != nil {
			return err
		}
		geo.Geometry = mp
		return nil
	case GeometryCollectionType:
		gc := geom.Collection{}
		var rawMessageForGeometries []*json.RawMessage
		if err := json.Unmarshal(*geojsonMap["geometries"], &rawMessageForGeometries); err != nil {
			return err
		}
		geoms := make([]geom.Geometry, len(rawMessageForGeometries))
		for i, v := range rawMessageForGeometries {
			var g Geometry
			if err := json.Unmarshal(*v, &g); err != nil {
				return err
			}
			geoms[i] = g.Geometry
		}
		gc.SetGeometries(geoms)
		geo.Geometry = gc
		return nil
	case FeatureType:
		f := Feature{}
		if err := json.Unmarshal(b, &f); err != nil {
			return err
		}
		geo.Geometry = f
		return nil
	case FeatureCollectionType:
		fc := FeatureCollection{}
		if err := json.Unmarshal(b, &fc); err != nil {
			return err
		}
		geo.Geometry = fc
		return nil
	default:
		return encoding.ErrInvalidGeoJSON{b}
	}
	return nil
}
