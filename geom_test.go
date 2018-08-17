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

	fn := func(t *testing.T, tc tcase) {
		r, err := GetCoordinates(tc.geom)
		if err != tc.err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !reflect.DeepEqual(r, tc.expected) {
			t.Errorf("error, expected %v got %v", tc.expected, r)
		}
	}
	testcases := []tcase{
		{
			geom:     Extent{},
			expected: nil,
			err:      ErrUnknownGeometry{Extent{}},
		},
		{
			geom:     Point{10, 20},
			expected: []Point{{10, 20}},
			err:      nil,
		},
		{
			geom: MultiPoint{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}},
			err:      nil,
		},
		{
			geom: LineString{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Point{{10, 20}, {30, 40}, {-10, -5}},
			err:      nil,
		},
		{
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
		{
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
		{
			geom: MultiPolygon{
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
		{
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
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}

func TestExtractLines(t *testing.T) {

	type tcase struct {
		geom     Geometry
		expected []Line
		err      error
	}

	fn := func(t *testing.T, tc tcase) {
		r, err := ExtractLines(tc.geom)
		if err != tc.err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if !(len(r) == 0 && len(tc.expected) == 0) && !reflect.DeepEqual(r, tc.expected) {
			t.Errorf("error, expected %v got %v", tc.expected, r)
		}
	}
	testcases := []tcase{
		{
			geom:     Extent{},
			expected: nil,
			err:      ErrUnknownGeometry{Extent{}},
		},
		{
			geom:     Point{10, 20},
			expected: []Line{},
			err:      nil,
		},
		{
			geom: MultiPoint{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Line{},
			err:      nil,
		},
		{
			geom: LineString{
				{10, 20},
				{30, 40},
				{-10, -5},
			},
			expected: []Line{{{10, 20}, {30, 40}}, {{30, 40}, {-10, -5}}},
			err:      nil,
		},
		{
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
		{
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
		{
			geom: MultiPolygon{
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
		{
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
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
