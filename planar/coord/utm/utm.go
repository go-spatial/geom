/*
Package utm provides the ability to work with UTM coordinates

	References:
		https://stevedutch.net/FieldMethods/UTMSystem.htm
		https://gisgeography.com/central-meridian/

*/
package utm

import (
	"fmt"
	"math"

	"github.com/go-spatial/geom/planar/coord"
)

const (
	k0 = 0.9996 // k0 - 0.9996  for UTM
	l0 = 0      // λ0 = center of the map. == 0 for us.
)

// UTM Zone Letters
const (
	ZoneC ZoneLetter = 'C'
	ZoneD ZoneLetter = 'D'
	ZoneE ZoneLetter = 'E'
	ZoneF ZoneLetter = 'F'
	ZoneG ZoneLetter = 'G'
	ZoneH ZoneLetter = 'H'
	ZoneI ZoneLetter = 'I'
	ZoneJ ZoneLetter = 'J'
	ZoneK ZoneLetter = 'K'
	ZoneL ZoneLetter = 'L'
	ZoneM ZoneLetter = 'M'
	ZoneN ZoneLetter = 'N'
	ZoneP ZoneLetter = 'P'
	ZoneQ ZoneLetter = 'Q'
	ZoneR ZoneLetter = 'R'
	ZoneS ZoneLetter = 'S'
	ZoneT ZoneLetter = 'T'
	ZoneU ZoneLetter = 'U'
	ZoneV ZoneLetter = 'V'
	ZoneW ZoneLetter = 'W'
	ZoneX ZoneLetter = 'X'
)

// ZoneLetter describes the UTM zone letter
type ZoneLetter byte

// String implements the stringer interface
func (zl ZoneLetter) String() string { return fmt.Sprintf("%c", zl) }

// IsNorthern returns if the Zone is in the northern hemisphere
func (zl ZoneLetter) IsNorthern() bool { return zl >= 'N' }

// IsValid will run validity check on the zone letter and number
func (zl ZoneLetter) IsValid() bool { return zl >= 'C' && zl <= 'X' && zl != 'O' }

// quick lookup table for central meridian for each zone
// this can be calculated using the formula:
//
//      ⎧ 30 - zone if zone <= 30
//    i ⎨ zone - 30 if zone > 30
//      ⎩
//    centralMeridianDegrees( zone ) = 3 + 6i
//
var centralMeridianDegrees = []uint{
	3, 9, 15, 21, 27, 33, 39, 45, 51, 57, 63, 69, 75, 81, 87, 93, 99, 105, 111, 117, 123, 129, 135, 141, 147, 153, 159, 165, 171, 177,
}

// CentralMeridian returns the central meridian degree for the given zone.
//
// Possible errors:
// 	ErrInvalidZone
func CentralMeridian(zone Zone) (degree int, err error) {

	if !zone.IsValid() {
		return 0, ErrInvalidZone
	}

	if zone.Number <= 30 {
		idx := 30 - zone.Number
		return -1 * int(centralMeridianDegrees[idx]), nil
	}
	idx := zone.Number - 31
	return int(centralMeridianDegrees[idx]), nil

}

// lngDigraphZones are the zone labels for the longitudinal values. The labels are split in middle
// so that one can use the central medial to figure grouping to use
var lngDigraphZones = [...][2][4]rune{
	[2][4]rune{
		[4]rune{'V', 'U', 'T', 'S'},
		[4]rune{'W', 'X', 'Y', 'X'},
	},
	[2][4]rune{
		[4]rune{'D', 'C', 'B', 'A'},
		[4]rune{'E', 'F', 'G', 'H'},
	},
	[2][4]rune{
		[4]rune{'M', 'L', 'K', 'J'},
		[4]rune{'N', 'P', 'Q', 'R'},
	},
}

// latDigraphZones are the zone labels for the latitudinal values.
var latDigraphZones = [...]rune{'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'L', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'A', 'B', 'C', 'D', 'E'}

// Digraph is the NATO diagraph
type Digraph [2]rune

// String returns the string representation of a diagraph
func (d Digraph) String() string { return string(d[:]) }

// newDigraph will calculate the digraph based on the provided zone and lnglat values.
// only error returned is ErrInvalidZone
//
// Reference: https://stevedutch.net/FieldMethods/UTMSystem.htm
func newDigraph(zone Zone, lnglat coord.LngLat) (Digraph, error) {
	if !zone.IsValid() {
		return Digraph{}, ErrInvalidZone
	}

	// How many times have we cycled through the zones.
	// There are only 3 zones total
	dZone := lngDigraphZones[zone.Number%3]

	// We already know the zone is valid
	cm, _ := CentralMeridian(zone)
	degreeDiff := float64(cm) - lnglat.Lng

	// This is the distance in km from the central meridian for the zone
	// Note by the definition of a zone, we know there are 111km per degree.
	kmDist := int(111 * degreeDiff * math.Cos(lnglat.LatInRadians()))

	// each square is 100 km wide, so we need to see how many of these do we cross.
	letterIdx := int(math.Abs(float64(kmDist / 100)))
	sideSelect := 0
	if degreeDiff < 0 {
		sideSelect = 1
	}
	lngLetter := dZone[sideSelect][letterIdx]

	kmDistLat := math.Abs(111.0 * lnglat.Lat)

	// if zone is even start at F
	// even zones are f-z
	// odd zones are a-v
	offset := -1
	if zone.Number%2 == 0 {
		offset = 4 // start at f
	}

	// there are 2000km per set of 20 100km blocks from the eq to the pole which is defined to be 40,000km
	idx := int(math.Abs(math.Ceil(
		float64(int(kmDistLat)%2000) / 100.0,
	)))
	// Southern hemisphere the values are laid out from the pole to the equator
	if !zone.IsNorthern() {
		idx = 21 - idx
	}

	letterIdx = offset + idx
	latLetter := latDigraphZones[letterIdx]

	return Digraph{lngLetter, latLetter}, nil
}

// Zone describes an UTM zone
type Zone struct {
	Number int
	Letter ZoneLetter
}

// String implements the stringer interface
func (z Zone) String() string { return fmt.Sprintf("%v%v", z.Number, z.Letter) }

// IsNorthern returns if the Zone is in the northern hemisphere
func (z Zone) IsNorthern() bool { return z.Letter.IsNorthern() }

// IsValid will run validity check on the zone letter and number
func (z Zone) IsValid() bool { return z.Letter.IsValid() && z.Number >= 1 && z.Number <= 60 }

// ZoneNumberFromLngLat will get the zone number for the given LngLat value.
//
// The returned value will be from 1-60.
// If 0 is returned it means that the lat,lng value was in the polar region
// and UPS should be used instead.
//
//	 Transcribed from:
//		 https://github.com/gdey/GDGeoCocoa/blob/master/GDGeoCoordConv.m
func ZoneNumberFromLngLat(lnglat coord.LngLat) int {
	lng, lat := lnglat.Lng, lnglat.Lat
	if (lat > 84.0 && lat < 90.0) || // North Pole
		(lat > -80.0 && lat < -90.0) { // South Pole
		return 0
	}

	// Adjust for projects.
	switch {
	case lat >= 56.0 && lat < 64.0 && lng >= 3.0 && lng < 12.0: // Exceptions around Norway
		return 32

	case lat >= 72.0 && lat < 84.0: // Exceptions around Svalbard
		switch {
		case lng >= 0.0 && lng < 9.0:
			return 31
		case lng >= 9.0 && lng < 21.0:
			return 33
		case lng >= 21.0 && lng < 33.0:
			return 35
		case lng >= 33.0 && lng < 42.0:
			return 37
		}
	}
	// Recast from [-180,180) to [0,360).
	// the w<-> is then divided into 60 zones from 1-60.
	return int((lng+180)/6) + 1
}

// ZoneLetterForLat returns the UTM zone letter for the given latitude value
//
// Possible errors:
//	 ErrLatitudeOutOfRange
func ZoneLetterForLat(lat float64) (ZoneLetter, error) {
	switch {
	case 84 >= lat && lat >= 72:
		return ZoneX, nil

	case 72 > lat && lat >= 64:
		return ZoneW, nil

	case 64 > lat && lat >= 56:
		return ZoneV, nil

	case 56 > lat && lat >= 48:
		return ZoneU, nil

	case 48 > lat && lat >= 40:
		return ZoneT, nil

	case 40 > lat && lat >= 32:
		return ZoneS, nil

	case 32 > lat && lat >= 24:
		return ZoneR, nil

	case 24 > lat && lat >= 16:
		return ZoneQ, nil

	case 16 > lat && lat >= 8:
		return ZoneP, nil

	case 8 > lat && lat >= 0:
		return ZoneN, nil

	case 0 > lat && lat >= -8:
		return ZoneM, nil

	case -8 > lat && lat >= -16:
		return ZoneL, nil

	case -16 > lat && lat >= -24:
		return ZoneK, nil

	case -24 > lat && lat >= -32:
		return ZoneJ, nil

	case -32 > lat && lat >= -40:
		return ZoneH, nil

	case -40 > lat && lat >= -48:
		return ZoneG, nil

	case -48 > lat && lat >= -56:
		return ZoneF, nil

	case -56 > lat && lat >= -64:
		return ZoneE, nil

	case -64 > lat && lat >= -72:
		return ZoneD, nil

	case -72 > lat && lat >= -80:
		return ZoneC, nil

	default:
		return 0, ErrLatitudeOutOfRange
	}
}

// NewZone returns the UTM zone for the given LngLat value.
//
// Possible errors:
//	 ErrLatitudeOutOfRange
func NewZone(lnglat coord.LngLat) (Zone, error) {
	number := ZoneNumberFromLngLat(lnglat)
	letter, err := ZoneLetterForLat(lnglat.Lat)
	return Zone{
		Number: number,
		Letter: letter,
	}, err
}

// Coord defines an UTM coordinate
type Coord struct {
	Northing float64
	Easting  float64
	Zone     Zone
	Digraph  Digraph
}

// ScalarFactor will calculate the correct k scalar value for the given lnglat and eccentricity
func ScalarFactor(lnglat coord.LngLat, e float64) float64 {
	lngRad, latRad := lnglat.LngInRadians(), lnglat.LatInRadians()

	dl := lngRad - l0
	dl2 := dl * dl
	dl4 := dl2 * dl2
	dl6 := dl4 * dl2

	e2 := (e * e) / (1 - (e * e))
	e4 := e2 * e2
	e6 := e4 * e2

	c := math.Cos(latRad)
	c2 := c * c
	c4 := c2 * c2
	c6 := c4 * c2

	t := math.Tan(latRad)
	t2 := t * t
	t4 := t2 * t2

	T26 := c2 / 2 * (1 + e2*c2)
	T27 := (c4 / 24) * (5 - (4 * t2) +
		(24 * e2 * c2) +
		(13 * e4 * c4) -
		(28 * t2 * e2 * c2) +
		(4 * e6 * c6) -
		(48 * t2 * e4 * c4) -
		(24 * t2 * e6 * c6))

	T28 := (c6 / 720) * (61 - (148 * t2) + (16 * t4))

	return k0 * (1 + (dl2 * T26) + (dl4 * T27) + (dl6 * T28))
}

// fromLngLat does the majority of the work.
//
// Valid zone and ellipsoid values assumed.
func fromLngLat(lnglat coord.LngLat, zone Zone, ellips coord.Ellipsoid) Coord {

	eccentricity, radius, nato := ellips.Eccentricity, ellips.Radius, ellips.NATOCompatible
	latRad, lngRad := lnglat.LatInRadians(), lnglat.LngInRadians()

	lngOrigin := float64((zone.Number-1)*6 - 180 + 3)
	lngOriginRad := coord.ToRadian(lngOrigin)
	eccentPrime := eccentricity / (1 - eccentricity)

	scale := k0

	sinLatRad := math.Sin(latRad)

	n := radius / math.Sqrt(1-eccentricity*sinLatRad*sinLatRad)
	t0 := 0.0
	if latRad != 0.0 {
		t0 = math.Tan(latRad)
	}
	cosLatRad := math.Cos(latRad)

	t := math.Pow(t0, 2.0)
	c := math.Pow(eccentPrime, 2.0) * math.Pow(cosLatRad, 2.0)
	a := (lngRad - lngOriginRad) * cosLatRad

	t2 := math.Pow(t, 2.0)
	t3 := math.Pow(t, 3.0)

	c2 := math.Pow(c, 2.0)

	a2 := math.Pow(a, 2.0)
	a3 := math.Pow(a, 3.0)
	a4 := math.Pow(a, 4.0)
	a5 := math.Pow(a, 5.0)
	a6 := math.Pow(a, 6.0)

	e2 := math.Pow(eccentricity, 2.0)
	e3 := math.Pow(eccentricity, 3.0)

	m0101 := (eccentricity / 4.0)
	m0102 := (3.0 / 64.0 * e2)
	m0103 := (5.0 / 256.0 * e3)

	m01 := (1 - m0101 - m0102 - m0103) * latRad
	m02 := ((3.0 / 8.0 * eccentricity) + (3.0 / 32.0 * e2) + (45.0 / 1024.0 * e3)) * math.Sin(latRad*2.0)

	m0301 := math.Sin(latRad * 4.0)
	m03 := ((15.0 / 256.0 * e2) + (45.0 / 1024.0 * e3)) * m0301

	m0401 := math.Sin(latRad * 6.0)
	m04 := (35.0 / 3072.0 * e3) * m0401

	m := radius * (m01 - m02 + m03 - m04)

	easting := scale*n*(a+(1.0-t+c)*a3/6.0+(5.0-(10.0*t3)+(72.0*c)-(58.0*eccentPrime))*a5/120.0) + 500000.0
	northing := scale * (m + n*t0*(a2/2.0+
		(5.0-t+(9.0*c)+(4.0*c2))*
			a4/24.0+
		(61.0-(58.0*t)+t2+(600.0*c)-(330.0*eccentPrime))*
			a6/720.0))

	if lnglat.Lat < 0.0 {
		northing += 10000000.0
	}
	var digraph Digraph
	// only compute digraph for nato
	if nato {
		// we know the zone is good
		digraph, _ = newDigraph(zone, lnglat)
	}
	return Coord{
		Northing: math.Round(northing),
		Easting:  math.Round(easting),
		Zone:     zone,
		Digraph:  digraph,
	}
}

// FromLngLat returns a new utm coordinate based on the provided longitude and latitude values.
func FromLngLat(lnglat coord.LngLat, ellips coord.Ellipsoid) (Coord, error) {
	zone, err := NewZone(lnglat)
	if err != nil {
		return Coord{}, err
	}
	return fromLngLat(lnglat.NormalizeLng(), zone, ellips), nil
}

// ToLngLat transforms the utm Coord to it's Lat Lng representation based on the given datum
func (c Coord) ToLngLat(ellips coord.Ellipsoid) (coord.LngLat, error) {

	if !c.Zone.IsValid() {
		return coord.LngLat{}, fmt.Errorf("invalid zone")
	}

	radius, ecc := ellips.Radius, ellips.Eccentricity
	x := c.Easting - 500000.0 // remove longitude offset
	y := c.Northing

	if !c.Zone.IsNorthern() {
		// remove Southern offset
		y -= 10000000.0
	}

	ecc2 := math.Pow(ecc, 2.0)
	ecc3 := math.Pow(ecc, 3.0)

	lngOrigin := float64((c.Zone.Number-1)*6 - 180 + 3)
	eccPrimeSqr := (ecc / (1.0 - ecc))
	m := y / k0
	mu := m / (radius * (1.0 - (ecc / 4.0) - (3.0 / 64.0 * ecc2) - (5.0 / 256.0 * ecc3)))

	e_1 := 1.0 - ecc
	e1 := (1.0 - math.Sqrt(e_1)) / (1.0 + math.Sqrt(e_1))
	e12 := math.Pow(e1, 2.0)
	e13 := math.Pow(e1, 3.0)
	e14 := math.Pow(e1, 4.0)
	p1 := 3.0 / 2.0 * e1
	p2 := 27.0 / 32.0 * e13
	p3 := math.Sin(mu * 2.0)
	p4 := 21.0 / 16.0 * e12
	p5 := 55.0 / 32.0 * e14
	p6 := math.Sin(mu * 4.0)
	p7 := 151.0 / 96.0 * e13
	p8 := math.Sin(mu * 6.0)

	phi1Rad := mu + (p1-p2)*p3 + (p4-p5)*p6 + (p7)*p8

	phi1Tan := math.Tan(phi1Rad)
	phi1Sin := math.Sin(phi1Rad)
	phi1Cos := math.Cos(phi1Rad)

	a := 1 - (ecc * phi1Sin * phi1Sin)

	n1 := radius / math.Sqrt(a)
	t1 := math.Pow(phi1Tan, 2)
	t12 := math.Pow(t1, 2)
	c1 := ecc * math.Pow(phi1Cos, 2)
	c12 := math.Pow(c1, 2)
	c12_3 := 3 * c12
	r1 := radius * e_1 / math.Pow(a, 1.5)
	d := x / (n1 * k0)

	latRad := phi1Rad -
		(n1*phi1Tan/r1)*
			((math.Pow(d, 2)/2)-
				(5+3*t1+10*c1-4*c12-9*eccPrimeSqr)*
					math.Pow(d, 4)/24+
				(61+90*t1+298*c1+45*t12-252*eccPrimeSqr-c12_3)*
					math.Pow(d, 6)*720)

	lngRad := (d -
		(1+2*t1+c1)*
			math.Pow(d, 3)/6 +
		(5-2*c1+28*t1-c12_3+8*eccPrimeSqr+24*t12)*
			math.Pow(d, 5)/120) /
		phi1Cos

	return coord.LngLat{
		Lng: lngOrigin + coord.ToDegree(lngRad),
		Lat: coord.ToDegree(latRad),
	}, nil
}

var x50 = int(math.Pow(10, 5))

// NatoEasting returns the easting value for NATO
func (c Coord) NatoEasting() int { return int(c.Easting) % x50 }

// NatoNorthing returns the northing value for NATO
func (c Coord) NatoNorthing() int { return int(c.Northing) % x50 }
