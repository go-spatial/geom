package hitmap

import (
	"errors"
	"sort"

	"github.com/go-spatial/geom"
)

var ErrInvalidLineString = errors.New("invalid linestring")

// label is the he label for the triangle. Is in "inside" or "outside".
// TODO: gdey — would be make more sense to just have a bool here? IsInside or somthing like that?
type Label uint8

func (l Label) String() string {
	switch l {
	case Outside:
		return "outside"
	case Inside:
		return "inside"
	default:
		return "unknown"
	}
}

const (
	Unknown Label = iota
	Outside
	Inside
)

type Interface interface {
	LabelFor(pt [2]float64) Label
	Extent() [4]float64
	Area() float64
}

func asGeomExtent(e [4]float64) *geom.Extent {
	ee := geom.Extent(e)
	return &ee

}

// PolygonHMSliceByAreaDec will allow you to sort a slice of PolygonHM in decending order
type ByAreaDec []Interface

func (hm ByAreaDec) Len() int      { return len(hm) }
func (hm ByAreaDec) Swap(i, j int) { hm[i], hm[j] = hm[j], hm[i] }
func (hm ByAreaDec) Less(i, j int) bool {
	ia, ja := hm[i].Area(), hm[j].Area()
	return ia < ja
}

// OrderedHM will iterate through a set of HitMaps looking for the first one to return
// inside, if none of the hitmaps return inside it will return outside.
type OrderedHM []Interface

func (hms OrderedHM) LabelFor(pt [2]float64) Label {
	for i := range hms {
		if hms[i].LabelFor(pt) == Inside {
			return Inside
		}
	}
	return Outside
}

// Extent is the accumlative extent of all the extens in the slice.
func (hms OrderedHM) Extent() [4]float64 {
	e := new(geom.Extent)
	for i := range hms {
		e.Add(asGeomExtent(hms[i].Extent()))
	}
	return e.Extent()
}

// Area returns the area of the total extent of the hitmaps that are contain in the slice.
func (hms OrderedHM) Area() float64 {
	return asGeomExtent(hms.Extent()).Area()
}

// NewOrderdHM will add the provided hitmaps in reverse order so that the last hit map is always tried first.
func NewOrderedHM(hms ...Interface) OrderedHM {
	ohm := make(OrderedHM, len(hms))
	size := len(hms) - 1
	for i := size; i >= 0; i-- {
		ohm[size-i] = hms[i]
	}
	return ohm
}

// NewHitMap will return a Polygon Hit map, a Ordered Hit Map, or a nil Hit map based on the geomtry type.
func New(clipbox *geom.Extent, geo geom.Geometry) (Interface, error) {
	var err error
	switch g := geo.(type) {
	case geom.Polygoner:

		ghm, err := NewFromPolygons(clipbox, g.LinearRings())
		if err != nil {
			return nil, err
		}

		return ghm, nil

	case geom.MultiPolygoner:

		polygons := g.Polygons()
		ghms := make([]Interface, len(polygons))
		for i := range polygons {
			ghms[i], err = NewFromPolygons(clipbox, polygons[i])
			if err != nil {
				return nil, err
			}
		}
		sort.Sort(ByAreaDec(ghms))
		return NewOrderedHM(ghms...), nil

	case geom.Collectioner:

		geometries := g.Geometries()
		ghms := make([]Interface, 0, len(geometries))
		for i := range geometries {
			g, err := New(clipbox, geometries[i])
			if err != nil {
				return nil, err
			}
			// skip empty hitmaps.
			if g == nil {
				continue
			}
			ghms = append(ghms, g)
		}
		sort.Sort(ByAreaDec(ghms))
		return NewOrderedHM(ghms...), nil

	case geom.Pointer, geom.MultiPointer, geom.LineStringer, geom.MultiLineStringer:
		return nil, nil

	default:
		return nil, geom.ErrUnknownGeometry{geo}
	}
}
