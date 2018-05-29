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
	"strconv"
	"testing"

	"github.com/go-spatial/geom"
)

/*
TestSegmentDummy keeps coveralls from complaining. Not really necessary tests.
*/
func TestSegmentDummy(t *testing.T) {
	type tcase struct {
		line geom.Line
	}

	fn := func(t *testing.T, tc tcase) {
		s := NewSegment(tc.line)
		if s.GetStart().Equals(tc.line[0]) == false {
			t.Errorf("error, expected %v got %v", tc.line[0], s.GetStart())
		}
		if s.GetEnd().Equals(tc.line[1]) == false {
			t.Errorf("error, expected %v got %v", tc.line[1], s.GetEnd())
		}
		if s.GetLineSegment() != tc.line {
			t.Errorf("error, expected %v got %v", tc.line, s.GetLineSegment())
		}
	}
	testcases := []tcase{
		{
			line: geom.Line{{1, 2}, {3, 4}},
		},
	}

	for i, tc := range testcases {
		tc := tc
		t.Run(strconv.FormatInt(int64(i), 10), func(t *testing.T) { fn(t, tc) })
	}
}
