/*
Copyright (c) 2016 Vivid Solutions.

All rights reserved. This program and the accompanying materials
are made available under the terms of the Eclipse Public License v1.0
and Eclipse Distribution License v. 1.0 which accompanies this distribution.
The Eclipse Public License is available at http://www.eclipse.org/legal/epl-v10.html
and the Eclipse Distribution License is available at

http://www.eclipse.org/org/documents/edl-v10.php.
*/

package triangulate

import (
	"reflect"
	"strconv"
	"testing"

	"github.com/go-spatial/geom/planar/triangulate/quadedge"
)

func TestUnique(t *testing.T) {
	type tcase struct {
		points   []quadedge.Vertex
		expected []quadedge.Vertex
	}

	fn := func(t *testing.T, tc tcase) {
		var uut DelaunayTriangulationBuilder
		result := uut.unique(tc.points)
		if reflect.DeepEqual(result, tc.expected) == false {
			t.Errorf("error, expected %v got %v", tc.expected, result)
		}
		// This shouldn't exist with no data
		if uut.GetSubdivision() != nil {
			t.Errorf("error, expected nil got not nil")
		}
	}
	testcases := []tcase{
		{
			points:   []quadedge.Vertex{{0, 1}, {0, 1}},
			expected: []quadedge.Vertex{{0, 1}},
		},
		{
			points:   []quadedge.Vertex{{0, 1}, {0, 1}, {1, 0}},
			expected: []quadedge.Vertex{{0, 1}, {1, 0}},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
