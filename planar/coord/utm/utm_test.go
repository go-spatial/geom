package utm

import (
	"fmt"
	"github.com/go-spatial/geom/cmp"
	"github.com/go-spatial/geom/planar/coord"
	"strings"
	"testing"
	"unicode"
)

// datum is the ellipsoid Structure for various datums.
// this is here to reduce the dependency tree. Don't count on these
// to be valid or accuract, better to use the official values
var knownEllipsoids = []coord.Ellipsoid{
	{
		Name:         normalizeName("Airy"),
		Radius:       6377563,
		Eccentricity: 0.00667054,
	},
	{
		Name:         normalizeName("Clarke_1866"),
		Radius:       6378206,
		Eccentricity: 0.006768658,
	},
	{
		Name:           normalizeName("WGS_84"),
		Radius:         6378137,
		Eccentricity:   0.00669438,
		NATOCompatible: true,
	},
}

func tolerance(tol *float64) (float64, int64) {
	if tol != nil {
		return *tol, cmp.BitToleranceFor(*tol)
	}
	return cmp.Tolerance, int64(cmp.BitTolerance)
}

// normalizeName will modify the value a bit;  remove trailing spaces, collapsing and transform spaces to '_' and uppercase everything else
func normalizeName(s string) string {
	var str strings.Builder
	s = strings.TrimSpace(s)
	lastIsSpace := false
	for _, r := range s {
		if unicode.IsSpace(r) {
			if !lastIsSpace {
				lastIsSpace = true
				str.WriteRune('_')
			}
			continue
		} else {
			lastIsSpace = false
		}
		str.WriteRune(unicode.ToUpper(r))
	}
	return str.String()
}

func getEllipsoidByName(name string) coord.Ellipsoid {
	if name == "" {
		name = "WGS 84"
	}
	name = normalizeName(name)
	for _, ellp := range knownEllipsoids {
		if ellp.Name == name {
			return ellp
		}
	}
	panic(fmt.Sprintf("Unknown ellipsoid: %v", name))
}

func TestFromLngLat(t *testing.T) {
	type tcase struct {
		Desc          string
		LngLat        coord.LngLat
		EllipsoidName string
		Tolerance     *float64
		UTM           Coord
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			tol, bitTol := tolerance(tc.Tolerance)
			ellips := getEllipsoidByName(tc.EllipsoidName)
			utm, err := FromLngLat(tc.LngLat, ellips)
			// TODO(gdey): add support for tests that error
			if err != nil {
				t.Errorf("error, expected nil, got: %v", err)
				return
			}
			if !cmp.Float64(utm.Northing, tc.UTM.Northing, tol, bitTol) {
				t.Errorf("northing, expected %v, got: %v", tc.UTM.Northing, utm.Northing)
			}
			if !cmp.Float64(utm.Easting, tc.UTM.Easting, tol, bitTol) {
				t.Errorf("easting, expected %v, got: %v", tc.UTM.Easting, utm.Easting)
			}
			if utm.Zone != tc.UTM.Zone {
				t.Errorf("zone, expected %v, got: %v", tc.UTM.Zone, utm.Zone)
			}
			if utm.Digraph[0] != tc.UTM.Digraph[0] || utm.Digraph[1] != tc.UTM.Digraph[1] {
				t.Errorf("Digraph, expected %v, got: %v", tc.UTM.Digraph, utm.Digraph)
			}
		}
	}

	tests := []tcase{
		// Subtests
		{
			Desc: "Kabul",
			LngLat: coord.LngLat{
				Lng: 69.1503666510912,
				Lat: 34.52518357633554,
			},
			UTM: Coord{
				Northing: 3820400.0,
				Easting:  513800.0,
				Zone: Zone{
					Letter: ZoneS,
					Number: 42,
				},
				Digraph: Digraph{'W', 'D'},
			},
		},
		{
			Desc: "Brasil",
			LngLat: coord.LngLat{
				Lng: -49.463803,
				Lat: -11.126665,
			},
			UTM: Coord{
				Northing: 8769581,
				Easting:  667767,
				Zone: Zone{
					Letter: ZoneL,
					Number: 22,
				},
				Digraph: Digraph{'F', 'N'},
			},
		},
		{
			//https://metacpan.org/pod/Geo::Coordinates::UTM"
			Desc: "perl example",
			LngLat: coord.LngLat{
				Lng: -2.788951667,
				Lat: 57.803055556,
			},
			EllipsoidName: "Clarke_1866",
			UTM: Coord{
				Northing: 6406592,
				Easting:  512544,
				Zone: Zone{
					Letter: ZoneV,
					Number: 30,
				},
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestLngLat_NormalizeLng(t *testing.T) {
	type tcase struct {
		Desc   string
		LngLat coord.LngLat
		Lng    float64
	}

	fn := func(tc tcase) func(*testing.T) {

		return func(t *testing.T) {
			lng := tc.LngLat.NormalizeLng()
			if !cmp.Float(lng.Lng, tc.Lng) {
				t.Errorf("normalized lng, expected %v, got %v", tc.Lng, lng.Lng)
			}
		}
	}

	tests := []tcase{
		// Subtests
		{},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestLngLat_InRadians(t *testing.T) {
	type tcase struct {
		Desc       string
		LngLat     coord.LngLat
		LngRadians float64
		LatRadians float64
		Tolerance  *float64
	}

	fn := func(tc tcase) func(*testing.T) {
		tol, bitTol := tolerance(tc.Tolerance)
		return func(t *testing.T) {

			t.Run("LatInRadians", func(t *testing.T) {
				lat := tc.LngLat.LatInRadians()
				if !cmp.Float64(lat, tc.LatRadians, tol, bitTol) {
					t.Errorf("radians, expected %v, got %v", tc.LatRadians, lat)
				}
			})

			t.Run("LngInRadians", func(t *testing.T) {
				lng := tc.LngLat.LngInRadians()
				if !cmp.Float64(lng, tc.LngRadians, tol, bitTol) {
					t.Errorf("radians, expected %v, got %v", tc.LngRadians, lng)
				}
			})

		}
	}

	tests := []tcase{
		// Subtests
		{
			LngLat: coord.LngLat{
				Lng: 69.1503666510912,
				Lat: 34.5251835763355,
			},
			LatRadians: 0.602578128262526,
			LngRadians: 1.20690157702283,
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestNewDigraph(t *testing.T) {
	type tcase struct {
		Desc string
		// Add Additonal Fields here
		LngLat  coord.LngLat
		Zone    *Zone
		Digraph Digraph
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			var (
				zone Zone
				err  error
			)

			if tc.Zone == nil {
				zone, err = NewZone(tc.LngLat)
				if err != nil {
					panic("Error test LatLng not producing good zone")
				}

			} else {
				zone = *tc.Zone
			}

			digraph, _ := newDigraph(zone, tc.LngLat)
			if digraph[0] != tc.Digraph[0] || digraph[1] != tc.Digraph[1] {
				t.Errorf("digraph, expected %v got %v", tc.Digraph, digraph)
			}

		}
	}

	tests := []tcase{
		// Subtests
		{
			Desc: "Green Bay",
			LngLat: coord.LngLat{
				Lat: 44.438486,
				Lng: -88.0,
			},
			Digraph: Digraph{'D', 'Q'},
		},
		{
			Desc: "Kabul",
			LngLat: coord.LngLat{
				Lng: 69.1503666510912,
				Lat: 34.52518357633554,
			},
			Digraph: Digraph{'W', 'D'},
		},
		{
			Desc: "Brasil even zone",
			LngLat: coord.LngLat{
				Lng: -49.463803,
				Lat: -11.126665,
			},
			Digraph: Digraph{'F', 'N'},
		},
		{
			Desc: "Brasil odd zone",
			LngLat: coord.LngLat{
				Lat: -11.126665015021864,
				Lng: -43.46380056756961 ,
			},
			Digraph: Digraph{'P', 'H'},
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}
