package geom

import (
	"reflect"
	"strconv"
	"testing"
)

func TestGetCoordinates(t *testing.T) {

	type tcase struct {
		geom     Geometry
		expected []Point
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T){
		r, err := GetCoordinates(tc.geom)
		if err != tc.err {
			t.Errorf("error, expected %v got %v", tc.err, err)
			return
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Errorf("error, expected %v got %v", tc.expected, r)
			return
		}
		}
	}
	testcases := []tcase{
		{ // 0
			geom:     Extent{},
			expected: nil,
			err:      ErrUnknownGeometry{Extent{}},
		},
		{ // 1
			geom:     Point{10, 20},
			expected: []Point{{10, 20}},
			err:      nil,
		},
		{ // 2
			geom: MultiPoint{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}},
			err:      nil,
		},
		{ // 3
			geom: LineString{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}},
			err:      nil,
		},
		{ // 4
			geom: MultiLineString{
				{
					{10, 20},
					{30, 40},
				},
				{
					{-10, -5},
					{15, 20},
				},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}, {15, 20}},
			err:      nil,
		},
		{ // 5
			geom: Polygon{
				{
					{10, 20},
					{30, 40},
					{-10, -5},
				},
				{
					{1, 2},
					{3, 4},
				},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}, {1, 2}, {3, 4}},
			err:      nil,
		},
		{ // 6
			geom: &MultiPolygon{
				{
					{
						{10, 20},
						{30, 40},
						{-10, -5},
					},
					{
						{1, 2},
						{3, 4},
					},
				},
				{
					{
						{5, 6},
						{7, 8},
						{9, 10},
					},
				},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}, {1, 2}, {3, 4}, {5, 6}, {7, 8}, {9, 10}},
			err:      nil,
		},
		{ // 7
			geom: Collection{
				Point{10, 20},
				MultiPoint{
					{10, 20},
					{30, 40},
					{-10, -5},
				},
				LineString{
					{1, 2},
					{3, 4},
					{5, 6},
				},
			},
			expected: []Point{{10, 20}, {10, 20}, {30, 40}, {-10, -5}, {1, 2}, {3, 4}, {5, 6}},
			err:      nil,
		},
	}

	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc) )
	}
}

func TestExtractLines(t *testing.T) {

	type tcase struct {
		geom     Geometry
		expected []Line
		err      error
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			r, err := ExtractLines(tc.geom)
			if err != tc.err {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}
			if !(len(r) == 0 && len(tc.expected) == 0) && !reflect.DeepEqual(r, tc.expected) {
				t.Errorf("error, expected %v got %v", tc.expected, r)
				return
			}
		}
	}
	testcases := []tcase{
		{ // 0
			geom:     Extent{},
			expected: nil,
			err:      ErrUnknownGeometry{Extent{}},
		},
		{ // 1
			geom:     Point{10, 20},
			expected: []Line{},
			err:      nil,
		},
		{ // 2
			geom: MultiPoint{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Line{},
			err:      nil,
		},
		{ // 3
			geom: LineString{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Line{{{10, 20}, {30, 40}}, {{30, 40}, {-10, -5}}},
			err:      nil,
		},
		{ // 4
			geom: MultiLineString{
				{
					{10, 20},
					{30, 40},
				},
				{
					{-10, -5},
					{15, 20},
				},
			},
			expected: []Line{{{10, 20}, {30, 40}}, {{-10, -5}, {15, 20}}},
			err:      nil,
		},
		{ // 5
			geom: Polygon{
				{
					{10, 20},
					{30, 40},
					{-10, -5},
				},
				{
					{1, 2},
					{3, 4},
				},
			},
			expected: []Line{{{10, 20}, {30, 40}}, {{30, 40}, {-10, -5}}, {{-10, -5}, {10, 20}}, {{1, 2}, {3, 4}}},
			err:      nil,
		},
		{ // 6
			geom: &MultiPolygon{
				{
					{
						{10, 20},
						{30, 40},
						{-10, -5},
					},
					{
						{1, 2},
						{3, 4},
					},
				},
				{
					{
						{5, 6},
						{7, 8},
						{9, 10},
					},
				},
			},
			expected: []Line{{{10, 20}, {30, 40}}, {{30, 40}, {-10, -5}}, {{-10, -5}, {10, 20}}, {{1, 2}, {3, 4}}, {{5, 6}, {7, 8}}, {{7, 8}, {9, 10}}, {{9, 10}, {5, 6}}},
			err:      nil,
		},
		{ // 7
			geom: Collection{
				Point{10, 20},
				MultiPoint{
					{10, 20},
					{30, 40},
					{-10, -5},
				},
				LineString{
					{1, 2},
					{3, 4},
					{5, 6},
				},
			},
			expected: []Line{{{1, 2}, {3, 4}}, {{3, 4}, {5, 6}}},
			err:      nil,
		},
	}

	for i, tc := range testcases {
		t.Run(strconv.FormatInt(int64(i), 10), fn(tc))
	}
}

func TestGetExtent(t *testing.T) {
	type tcase struct {
		g         Geometry
		e         Extent
		err       error
		expectedE Extent
	}
	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			// make a copy of e
			e := tc.e
			err := getExtent(tc.g, &e)
			if (tc.err != nil && err == nil) || (tc.err == nil && err != nil) {
				t.Errorf("err failed, expected err %v got err %v", tc.err, err)
				return
			}
			if tc.err != nil {
				// if there is an error we expected, continue to next test case.
				return
			}
			if e[0] != tc.expectedE[0] ||
				e[1] != tc.expectedE[1] ||
				e[2] != tc.expectedE[2] ||
				e[3] != tc.expectedE[3] {
				t.Errorf("extent, expected %v got %v", tc.expectedE, e)
			}
		}
	}
	tests := map[string]tcase{
		"Henery Circle One": {
			g: &MultiPolygon{
				{ // Polygon
					{ // Ring
						{1286956.1422558832, 6138803.15957211},
						{1286957.5138675969, 6138809.6399925},
						{1286961.0222077654, 6138815.252628375},
						{1286966.228733862, 6138819.3396373615},
						{1286972.5176202222, 6138821.397139203},
						{1286979.1330808033, 6138821.173193399},
						{1286985.2820067848, 6138818.695793352},
						{1286990.1992814348, 6138814.272866236},
						{1286993.3157325392, 6138808.436285537},
						{1286994.2394710402, 6138801.885883152},
						{1286992.8678593265, 6138795.40546864},
						{1286989.3781805448, 6138789.792845784},
						{1286984.1623237533, 6138785.719847463},
						{1286977.864106701, 6138783.662354196},
						{1286971.2486461198, 6138783.872302467},
						{1286965.1183815224, 6138786.349692439},
						{1286960.1824454917, 6138790.7726051165},
						{1286957.084655768, 6138796.623170342},
						{1286956.1422558832, 6138803.15957211},
					},
				},
			},
			e:         Extent{1286956.1422558832, 6138803.15957211, 1286956.1422558832, 6138803.15957211},
			expectedE: Extent{1.2869561422558832e+06,6.138783662354196e+06,1.2869942394710402e+06,6.138821397139203e+06},
		},
	}
	for name, tc := range tests {
		t.Run(name, fn(tc))
	}
}

