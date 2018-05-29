package geojson

import (
	"encoding/json"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding"
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
			Coords: g.Verticies(),
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
			Type: PolygonType,
			// make sure our rings are closed
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

type Feature struct {
	Type featureType `json:"type"`
	ID   *uint64     `json:"id,omitempty"`
	// can be null
	Geometry Geometry `json:"geometry"`
	// can be null
	Properties map[string]interface{} `json:"properties"`
}

// featureCollectionType allows the GeoJSON type for Feature to be automatically set during json Marshalling
// which avoids the user from accidentally setting the incorrect GeoJSON type.
type featureCollectionType struct{}

func (_ featureCollectionType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + FeatureCollectionType + `"`), nil
}
func (fc *featureCollectionType) UnmarshalJSON(b []byte) error { return nil }

type FeatureCollection struct {
	Type     featureCollectionType `json:"type"`
	Features []Feature             `json:"features"`
}

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
