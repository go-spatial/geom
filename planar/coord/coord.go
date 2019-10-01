package coord

import (
	"fmt"
	"math"
)

// Interface is an interface that wraps a ToLngLat methods
//
// ToLngLat should returns the a LngLat value in degrees that
// represents that given value as closely as possible. It is
// understood that this value may not be excate
type Interface interface {
	ToLngLat() LngLat
}

// LngLat describes Latitude,Longitude values
type LngLat struct {
	// Longitude in decimal degrees
	Lng float64
	// Latitude in decimal degrees
	Lat float64
}

// NormalizeLng will insure that the longitude value is between 0-360
func (l LngLat) NormalizeLng() LngLat {
	return LngLat{
		Lat: l.Lat,
		Lng: l.Lng - float64(int64((l.Lng+180.0)/360.0)*360.0),
	}
}

// LatInRadians returns the Latitude value in radians
func (l LngLat) LatInRadians() float64 { return ToRadian(l.Lat) }

// LngInRadians returns the Longitude value in radians
func (l LngLat) LngInRadians() float64 { return ToRadian(l.Lng) }

// LatAsDMS returns the latitude value in Degree Minute Seconds values
func (l LngLat) LatAsDMS() DMS {
	latD, latM, latS := toDMS(l.Lat)
	latH := 'N'
	if l.Lat < 0 {
		latH = 'S'
	}
	return DMS{
		Degree:     latD,
		Minute:     latM,
		Second:     latS,
		Hemisphere: latH,
	}
}

// LngAsDMS returns the longitude value in Degree Minute Seconds values
func (l LngLat) LngAsDMS() DMS {
	lngD, lngM, lngS := toDMS(l.Lng)
	lngH := 'E'
	if l.Lng < 0 {
		lngH = 'W'
	}
	return DMS{
		Degree:     lngD,
		Minute:     lngM,
		Second:     lngS,
		Hemisphere: lngH,
	}
}

// ToRadian will convert a value in degree to radians
func ToRadian(degree float64) float64 {
	return degree * math.Pi / 180.000
}

// ToDegree will convert a value in radians to degrees
func ToDegree(radian float64) float64 {
	return radian * 180.000 / math.Pi
}

// Ellipsoid describes an Ellipsoid
// this may change when we get a proper projection package
type Ellipsoid struct {
	Name           string
	Radius         float64
	Eccentricity   float64
	NATOCompatible bool
}

// Convert the given lng or lat value to the degree minute seconds values
func toDMS(v float64) (d int64, m int64, s float64) {
	var frac float64
	df, frac := math.Modf(v)
	mf, frac := math.Modf(60.0 * frac)
	s = 60.0 * frac
	return int64(math.Abs(df)), int64(math.Abs(mf)), math.Abs(s)
}

// DMS is the degree minutes and seconds
type DMS struct {
	Degree     int64
	Minute     int64
	Second     float64
	Hemisphere rune
}

// String returns the string representation.
func (dms DMS) String() string {
	return fmt.Sprintf(`%d°%d'%f"%c`, dms.Degree, dms.Minute, dms.Second, dms.Hemisphere)
}
