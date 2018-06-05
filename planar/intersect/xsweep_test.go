package intersect

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

func mustAsSegments(g geom.Geometry) []geom.Line {
	switch g := g.(type) {
	default:
		panic(fmt.Sprintf("Unsupported geo! %T", g))
	case *geom.MultiLineString:
		ssegs, err := g.AsSegments()
		if err != nil {
			panic(err)
		}
		var segs []geom.Line
		for i := range ssegs {
			segs = append(segs, ssegs[i]...)
		}
		return segs
	}
	// No opt
	return nil
}

func TestIntersect(t *testing.T) {
	type iptval struct {
		i, j int
		pt   [2]float64
	}
	type eval struct {
		edge     int
		edgeType eventType
	}
	type tcase struct {
		segments  []geom.Line
		connected bool
		// expected events.
		events []eval
		// expected intersect points and the line indexes
		ipts []iptval
	}
	fn := func(t *testing.T, tc tcase) {
		t.Parallel()

		var pts []iptval

		eqfn := func(i, j int, pt [2]float64) error {

			pts = append(pts, iptval{
				i:  i,
				j:  j,
				pt: pt,
			})
			return nil
		}

		eq := NewEventQueue(tc.segments)

		// Test the eq is what we expect:
		if !reflect.DeepEqual(tc.segments, eq.segments) {
			t.Errorf("eq segments, expected %v got %v", tc.segments, eq.segments)
			return
		}

		if len(tc.events) != len(eq.events) {
			t.Errorf("eq events (len) , expected (%v) %v got (%v) %v", len(tc.events), tc.events, len(eq.events), eq.events)
			return
		}
		for i := range tc.events {
			if tc.events[i].edge != eq.events[i].edge ||
				tc.events[i].edgeType != eq.events[i].edgeType {
				t.Errorf("eq events (event %v) , expected %v got %v", i, tc.events, eq.events)
				return
			}
		}

		eq.FindIntersects(context.Background(), tc.connected, eqfn)

		// We need to now check to see if the points are the same.
		if !reflect.DeepEqual(tc.ipts, pts) {
			t.Errorf("intersect points, expected %v got %v", tc.ipts, pts)
		}

	}

	tests := [...]tcase{
		{
			segments: mustAsSegments(&geom.MultiLineString{
				[][2]float64{{1, 0}, {10, 9}, {14, 5}},
				[][2]float64{{2, 5}, {5, 2}},
				[][2]float64{{6, 3}, {9, 6}},
				[][2]float64{{10, 2}, {10, 9}, {15, 9}},
				[][2]float64{{9, 5}, {10, 5}, {10, 2}},
				[][2]float64{{8, 7}, {11, 7}},
				[][2]float64{{6, 5}, {3, 9}},
			}),
			events: []eval{
				{0, LEFT}, {2, LEFT}, {9, LEFT}, {2, RIGHT}, {3, LEFT}, {9, RIGHT}, {8, LEFT},
				{6, LEFT}, {3, RIGHT}, {4, LEFT}, {7, LEFT}, {1, LEFT}, {5, LEFT}, {6, RIGHT},
				{7, RIGHT}, {0, RIGHT}, {4, RIGHT}, {8, RIGHT}, {1, RIGHT}, {5, RIGHT},
			},
			connected: false,
			ipts: []iptval{
				{i: 2, j: 0, pt: [2]float64{4, 3}}, {i: 9, j: 0, pt: [2]float64{6, 5}},
				{i: 6, j: 4, pt: [2]float64{10, 5}}, {i: 6, j: 7, pt: [2]float64{10, 5}},
				{i: 0, j: 1, pt: [2]float64{10, 9}}, {i: 0, j: 4, pt: [2]float64{10, 9}},
				{i: 0, j: 5, pt: [2]float64{10, 9}}, {i: 0, j: 8, pt: [2]float64{8, 7}},
				{i: 4, j: 1, pt: [2]float64{10, 9}}, {i: 4, j: 5, pt: [2]float64{10, 9}},
				{i: 4, j: 8, pt: [2]float64{10, 7}}, {i: 1, j: 5, pt: [2]float64{10, 9}},
			},
		},
		{
			segments: mustAsSegments(&geom.MultiLineString{
				[][2]float64{{1, 0}, {10, 9}, {14, 5}},
				[][2]float64{{2, 5}, {5, 2}},
				[][2]float64{{6, 3}, {9, 6}},
				[][2]float64{{10, 2}, {10, 9}, {15, 9}},
				[][2]float64{{9, 5}, {10, 5}, {10, 2}},
				[][2]float64{{8, 7}, {11, 7}},
				[][2]float64{{6, 5}, {3, 9}},
			}),
			events: []eval{
				{0, LEFT}, {2, LEFT}, {9, LEFT}, {2, RIGHT}, {3, LEFT}, {9, RIGHT}, {8, LEFT},
				{6, LEFT}, {3, RIGHT}, {4, LEFT}, {7, LEFT}, {1, LEFT}, {5, LEFT}, {6, RIGHT},
				{7, RIGHT}, {0, RIGHT}, {4, RIGHT}, {8, RIGHT}, {1, RIGHT}, {5, RIGHT},
			},
			connected: true,
			ipts: []iptval{
				{i: 2, j: 0, pt: [2]float64{4, 3}}, {i: 9, j: 0, pt: [2]float64{6, 5}},
				{i: 6, j: 4, pt: [2]float64{10, 5}}, {i: 0, j: 8, pt: [2]float64{8, 7}},
				{i: 4, j: 8, pt: [2]float64{10, 7}},
			},
		},
	}

	for i := range tests {
		tc := tests[i]
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}

}
