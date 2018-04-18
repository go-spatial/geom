package token

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/go-spatial/geom"
)

func assertError(expErr, gotErr error) (msg, expected, got string, ok bool) {
	if expErr != gotErr {
		// could be because test.err == nil and err != nil.
		if expErr == nil && gotErr != nil {
			return "unexpected", "nil", gotErr.Error(), false
		}
		if expErr != nil && gotErr == nil {
			return "expected error", expErr.Error(), "nil", false
		}
		if expErr.Error() != gotErr.Error() {
			return "did not get correct error value", expErr.Error(), gotErr.Error(), false

		}
		return "", "", "", false
	}
	if expErr != nil {
		// No need to look at other values, expected an error.
		return "", "", "", false
	}
	return "", "", "", true
}

func TestParsePointValue(t *testing.T) {
	type tcase struct {
		input string
		exp   []float64
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		pts, err := tt.parsePointValue()
		if msg, expstr, gotstr, ok := assertError(tc.err, err); !ok {
			if msg != "" {
				t.Errorf("%v, expected %v got %v", msg, expstr, gotstr)
			}
			return
		}
		if !reflect.DeepEqual(tc.exp, pts) {
			t.Errorf("points, expected %v got %v", tc.exp, pts)
		}
	}
	tests := map[string]tcase{
		"1": {input: "123 123 12", exp: []float64{123, 123, 12}},
		"2": {input: "10.0 -34,", exp: []float64{10.0, -34}},
		"3": {input: "1 ", exp: []float64{1}},
		"4": {input: "1 .0", exp: []float64{1, 0}},
		"5": {input: "1 -.1", exp: []float64{1, -.1}},
		"6": {input: " 1 2 ", exp: []float64{1, 2}},
		"7": {input: "1 .", err: &strconv.NumError{
			Func: "ParseFloat",
			Num:  ".",
			Err:  fmt.Errorf(`invalid syntax`),
		}},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParsePointe(t *testing.T) {
	type tcase struct {
		input string
		exp   *geom.Point
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		pt, err := tt.ParsePoint()
		if msg, expstr, gotstr, ok := assertError(tc.err, err); !ok {
			if msg != "" {
				t.Errorf("%v, expected %v got %v", msg, expstr, gotstr)
			}
			return
		}
		if !reflect.DeepEqual(tc.exp, pt) {
			t.Errorf("point values, expected %v got %v", tc.exp, pt)
		}
	}
	tests := map[string]tcase{
		"POINT EMPTY": {
			input: "POINT EMPTY",
		},
		"POINT EMPTY ": {
			input: "POINT EMPTY ",
		},
		"POINT ( 1 2 )": {
			input: "POINT ( 1 2 )",
			exp:   &geom.Point{1, 2},
		},
		" POINT ( 1 2 ) ": {
			input: " POINT ( 1 2 ) ",
			exp:   &geom.Point{1, 2},
		},
		" POINT ZM ( 1 2 3 4 ) ": {
			input: " POINT ZM ( 1 2 3 4 ) ",
			exp:   &geom.Point{1, 2},
		},
		"POINT 1 2": {
			input: "POINT 1 2",
			err:   fmt.Errorf("expected to find “(” or “EMPTY”"),
		},
		"POINT ( 1 2": {
			input: "POINT ( 1 2",
			err:   fmt.Errorf("expected to find “)”"),
		},
		"POINT ( 1 )": {
			input: "POINT ( 1 )",
			err:   fmt.Errorf("expected to have at least 2 coordinates in a POINT"),
		},
		"POINT ( 1 2 3 4 5 )": {
			input: "POINT ( 1 2 3 4 5 )",
			err:   fmt.Errorf("expected to have no more then 4 coordinates in a POINT"),
		},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}

func TestParseMultiPointe(t *testing.T) {
	type tcase struct {
		input string
		exp   geom.MultiPoint
		err   error
	}

	fn := func(t *testing.T, tc tcase) {
		t.Parallel()
		tt := NewT(strings.NewReader(tc.input))
		mpt, err := tt.ParseMultiPoint()
		if msg, expstr, gotstr, ok := assertError(tc.err, err); !ok {
			if msg != "" {
				t.Errorf("%v, expected %v got %v", msg, expstr, gotstr)
			}
			return
		}
		if !reflect.DeepEqual(tc.exp, mpt) {
			t.Errorf("did not get correct multipoint values, expected %v got %v", tc.exp, mpt)
		}

	}
	tests := map[string]tcase{
		"empty": {input: "MultiPoint EMPTY"},
		"without pren": {
			input: "MULTIPOINT ( 10 10, 12 12 )",
			exp:   geom.MultiPoint{{10, 10}, {12, 12}},
		},
		"with pren": {
			input: "MULTIPOINT ( (10 10), (12 12) )",
			exp:   geom.MultiPoint{{10, 10}, {12, 12}},
		},
	}
	for name, test := range tests {
		test := test // make copy
		t.Run(name, func(t *testing.T) { fn(t, test) })
	}
}

func TestParseFloat64(t *testing.T) {
	type tcase struct {
		input string
		exp   float64
		err   error
	}
	fn := func(t *testing.T, tc tcase) {
		tt := NewT(strings.NewReader(tc.input))
		f, err := tt.ParseFloat64()
		if tc.err != err {
			t.Errorf("error, expected %v got %v", tc.err, err)
		}
		if tc.err != nil {
			return
		}
		if tc.exp != f {
			t.Errorf("prase float64 expected %v got %v", tc.exp, f)
		}
	}
	tests := map[string]tcase{
		"-12":         {input: "-12", exp: -12.0},
		"0":           {input: "0", exp: 0.0},
		"+1000.00":    {input: "+1000.00", exp: 1000.0},
		"-12000.00":   {input: "-12000.00", exp: -12000.0},
		"10.005e5":    {input: "10.005e5", exp: 10.005e5},
		"10.005e+5":   {input: "10.005e+5", exp: 10.005e5},
		"10.005e+05":  {input: "10.005e+05", exp: 10.005e5},
		"1.0005e+6":   {input: "1.0005e+6", exp: 10.005e5},
		"1.0005e+06":  {input: "1.0005e+06", exp: 10.005e5},
		"1.0005e-06":  {input: "1.0005e-06", exp: 1.0005e-06},
		"1.0005e-06a": {input: "1.0005e-06a", exp: 1.0005e-06},
	}
	for name, tc := range tests {
		tc := tc
		t.Run(name, func(t *testing.T) { fn(t, tc) })
	}
}
