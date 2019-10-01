package coord

import (
	"testing"

	"github.com/go-spatial/geom/cmp"
)

func tolerance(tol *float64) (float64, int64) {

	t := 0.0001
	if tol != nil {
		t = *tol
	}
	return t, cmp.BitToleranceFor(t)
}

func TestToRadianDegree(t *testing.T) {
	type tcase struct {
		Desc string
		// Add Additonal Fields here
		Degree    float64
		Radian    float64
		tolerance *float64
	}

	fn := func(tc tcase) func(*testing.T) {

		tol, bitTol := tolerance(tc.tolerance)
		return func(t *testing.T) {

			t.Run("ToRadian", func(t *testing.T) {
				rad := ToRadian(tc.Degree)
				if !cmp.Float64(rad, tc.Radian, tol, bitTol) {
					t.Errorf("radian, expect %v, got %v", tc.Radian, rad)
				}
			})

			t.Run("ToDegree", func(t *testing.T) {
				deg := ToDegree(tc.Radian)
				if !cmp.Float64(deg, tc.Degree, tol, bitTol) {
					t.Errorf("degree, expect %v, got %v", tc.Degree, deg)
				}
			})

		}
	}

	tests := []tcase{
		// Subtests
		{
			Degree: 0.0,
			Radian: 0.0,
		},
		{
			Degree: 1.0,
			Radian: 0.017453,
		},
		{
			Degree: 60.0,
			Radian: 1.0472,
		},
		{
			Degree: 90.0,
			Radian: 1.5708,
		},
		{
			Degree: 180.0,
			Radian: 3.14159,
		},
		{
			Degree: 360.0,
			Radian: 6.28319,
		},
		{
			Degree: 580.0,
			Radian: 10.1229,
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func cmpDMS(a, b DMS) bool {

	return !(a.Degree != b.Degree ||
		a.Minute != b.Minute ||
		!cmp.Float(a.Second, b.Second) ||
		a.Hemisphere != b.Hemisphere)

}

func TestLngLat_ToDMS(t *testing.T) {
	type tcase struct {
		Desc string
		// Add Additonal Fields here
		LngLat LngLat
		LngDMS DMS
		LatDMS DMS
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {

			t.Run("LatDMS", func(t *testing.T) {
				dms := tc.LngLat.LatAsDMS()
				if !cmpDMS(dms, tc.LatDMS) {
					t.Errorf("dms, expected %v got %v", tc.LatDMS, dms)
				}
			})

			t.Run("LngDMS", func(t *testing.T) {
				dms := tc.LngLat.LngAsDMS()
				if !cmpDMS(dms, tc.LngDMS) {
					t.Errorf("dms, expected %v got %v", tc.LngDMS, dms)
				}
			})

		}
	}

	tests := []tcase{
		// Subtests
		{
			Desc: "noman's land",
			LngLat: LngLat{
				Lng: 0.0,
				Lat: 0.0,
			},
			LngDMS: DMS{
				Degree:     0,
				Minute:     0,
				Second:     0.0,
				Hemisphere: 'E',
			},
			LatDMS: DMS{
				Degree:     0,
				Minute:     0,
				Second:     0.0,
				Hemisphere: 'N',
			},
		},
		{
			Desc: "india",
			LngLat: LngLat{
				Lat: 21.991952,
				Lng: 78.873755,
			},
			LngDMS: DMS{
				Degree:     78,
				Minute:     52,
				Second:     25.518,
				Hemisphere: 'E',
			},
			LatDMS: DMS{
				Degree:     21,
				Minute:     59,
				Second:     31.0272,
				Hemisphere: 'N',
			},
		},
		{
			Desc: "zambia",
			LngLat: LngLat{
				Lat: -14.723885,
				Lng: 26.162606,
			},
			LatDMS: DMS{
				Degree:     14,
				Minute:     43,
				Second:     25.986,
				Hemisphere: 'S',
			},
			LngDMS: DMS{
				Degree:     26,
				Minute:     9,
				Second:     45.3816,
				Hemisphere: 'E',
			},
		},
		{
			Desc: "brasil",
			LngLat: LngLat{
				Lat: -11.126663,
				Lng: -49.038633,
			},
			LatDMS: DMS{
				Degree:     11,
				Minute:     7,
				Second:     35.9868,
				Hemisphere: 'S',
			},
			LngDMS: DMS{
				Degree:     49,
				Minute:     2,
				Second:     19.0788,
				Hemisphere: 'W',
			},
		},
		{
			Desc: "north canada",
			LngLat: LngLat{
				Lat: 66.743373,
				Lng: -102.452597,
			},
			LatDMS: DMS{
				Degree:     66,
				Minute:     44,
				Second:     36.1428,
				Hemisphere: 'N',
			},
			LngDMS: DMS{
				Degree:     102,
				Minute:     27,
				Second:     9.3492,
				Hemisphere: 'W',
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestDMS_String(t *testing.T) {
	type tcase struct {
		Desc string
		// Add Additonal Fields here
		DMS  DMS
		Form string
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			str := tc.DMS.String()
			if str == tc.Form {
				t.Errorf("str, got %v expected %v", str, tc.Form)
			}
		}
	}

	tests := []tcase{
		// Subtests
		{
			Form: `66°44'36.1428" N`,
			DMS: DMS{
				Degree:     66,
				Minute:     44,
				Second:     36.1428,
				Hemisphere: 'N',
			},
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}

func TestLngLat_InRadians(t *testing.T) {
	type tcase struct {
		Desc       string
		LngLat     LngLat
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
			LngLat: LngLat{
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

func TestLngLat_NormalizeLng(t *testing.T) {
	type tcase struct {
		Desc   string
		LngLat LngLat
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
		{
			Desc: "Brasil",
			LngLat: LngLat{
				Lng: -49.463803,
				Lat: -11.126665,
			},
			Lng: -49.463803,
		},
		{

			Desc: "Kabul",
			LngLat: LngLat{
				Lng: 69.1503666510912,
				Lat: 34.52518357633554,
			},
			Lng: 69.1503666510912,
		},
	}

	for i := range tests {
		t.Run(tests[i].Desc, fn(tests[i]))
	}
}
