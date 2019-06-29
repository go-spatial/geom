package subdivision

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
	"github.com/go-spatial/geom/encoding/wkt"
)

var showanswer bool

func init() {
	showanswer, _ = strconv.ParseBool(strings.TrimSpace(os.Getenv("SHOWANSWER")))
}

func TestTriangulatePseudoPolygon(t *testing.T) {
	type tcase struct {
		points []geom.Point

		edges []geom.Line
		err   error
	}
	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {

			edges, err := triangulatePseudoPolygon(tc.points)

			if tc.err != nil {
				if tc.err != err {
					t.Errorf("error, expected %v got %v", tc.err, err)
				}
				return
			}

			if err != nil {
				t.Errorf("error, expected %v got %v", tc.err, err)
				return
			}

			if !reflect.DeepEqual(tc.edges, edges) {
				t.Errorf("edges,\n\t expected %v\n\t got      %v", wkt.MustEncode(tc.edges), wkt.MustEncode(edges))
				if showanswer {
					t.Errorf("edges:\n%#v", edges)
				}
			}

		}
	}
	tests := [...]tcase{
		{
			points: []geom.Point{},
			err:    ErrInvalidPseudoPolygonSize,
		},
		{
			points: []geom.Point{{0, 0}},
			err:    ErrInvalidPseudoPolygonSize,
		},
		{
			points: []geom.Point{{0, 0}, {1, 1}},
			edges:  []geom.Line{{{0, 0}, {1, 1}}},
		},
		{ // simple triangle
			points: []geom.Point{{10, 10}, {10, 20}, {20, 20}},
			edges: []geom.Line{
				{{10, 10}, {10, 20}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {10, 10}},
			},
		},
		{
			points: []geom.Point{{10, 20}, {20, 20}, {20, 10}, {20, 0}, {10, 0}, {0, 0}, {0, 10}, {0, 20}},
			edges: []geom.Line{
				{{10, 0}, {0, 20}},
				{{0, 20}, {10, 20}},
				{{10, 20}, {10, 0}},
				{{10, 0}, {10, 20}},
				{{20, 10}, {10, 0}},
				{{20, 10}, {10, 20}},
				{{10, 20}, {20, 20}},
				{{20, 20}, {20, 10}},
				{{20, 10}, {20, 0}},
				{{20, 0}, {10, 0}},
				{{10, 0}, {20, 10}},
				{{0, 20}, {10, 0}},
				{{0, 10}, {0, 20}},
				{{0, 10}, {10, 0}},
				{{10, 0}, {0, 0}},
				{{0, 0}, {0, 10}},
			},
		},
		{
			points: []geom.Point{
				{10, 0},
				{0, 0},
				{0, 10},
				{0, 20},
			},
			edges: []geom.Line{
				{{0, 20}, {10, 0}},
				{{0, 10}, {0, 20}},
				{{0, 10}, {10, 0}},
				{{10, 0}, {0, 0}},
				{{0, 0}, {0, 10}},
			},
		},
		{
			points: []geom.Point{
				{458, 1228}, {457, 1225}, {449, 1196}, {456, 1225}, {457, 1232},
			},
			edges: []geom.Line{
				{{457, 1232}, {456, 1225}},
				{{457, 1225}, {457, 1232}},
				{{457, 1225}, {456, 1225}},
				{{456, 1225}, {449, 1196}},
				{{449, 1196}, {457, 1225}},
				{{457, 1232}, {458, 1228}},
				{{458, 1228}, {457, 1225}},
				{{457, 1225}, {457, 1232}},
			},
		},
	}
	for i, tc := range tests {
		t.Run(strconv.Itoa(i), fn(tc))
	}
}
