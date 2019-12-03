package geom

import (
	"errors"
	"reflect"
	"testing"
)

func TestClone(t *testing.T) {
	type tcase struct {
		a   Geometry
		err string
	}

	fn := func(tc tcase) func(t *testing.T) {
		return func(t *testing.T) {
			b, err := Clone(tc.a)
			if err != nil {
				if err.Error() != tc.err {
					t.Fatalf("expected error %v got %v", tc.err, err)
				}
				return
			} else if len(tc.err) != 0 {
				t.Fatalf("expected error %v got %v", tc.err, err)
				return
			}

			if !reflect.DeepEqual(b, tc.a) {
				t.Fatalf("expected geom %v got %v", tc.a, b)
			}
		}
	}

	tcases := map[string]tcase{
		"err type": {
			a:   int(0),
			err: "unknown Geometry: int",
		},
		"point ok": {
			a: Point{},
		},
		"point ok 2": {
			a: Point{3.14, 2.7},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}

func TestApply(t *testing.T) {
	type tcase struct {
		a, b Geometry
		f    func(...float64) ([]float64, error)
		err  string
	}

	fn := func(tc tcase) func(*testing.T) {
		return func(t *testing.T) {
			b, err := ApplyToPoints(tc.a, tc.f)
			if err != nil {
				if err.Error() != tc.err {
					t.Fatalf("expected error %v got %v", tc.err, err)
				}
				return
			} else if len(tc.err) != 0 {
				t.Fatalf("expected error %v got %v", tc.err, err)
			}

			if !reflect.DeepEqual(b, tc.b) {
				t.Fatalf("expected %v got %v", tc.b, b)
			}
		}
	}

	tcases := map[string]tcase{
		"type err": {
			a:   int(0),
			err: "unknown Geometry: int",
		},
		"func err": {
			a:   Point{},
			f:   func(...float64) ([]float64, error) { return nil, errors.New("fn err") },
			err: "fn err",
		},
		"add 1": {
			a: Point{},
			b: Point{1, 1},
			f: func(p ...float64) ([]float64, error) {
				for i, v := range p {
					p[i] = v + 1
				}

				return p, nil
			},
		},
	}

	for k, v := range tcases {
		t.Run(k, fn(v))
	}
}
