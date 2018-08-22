package walker

import (
	"reflect"
	"strconv"
	"testing"
)

//TestCutPanic test the panic conditions of the cut function.
func TestCutPanic(t *testing.T) {
	test := func(start, end int) func(*testing.T) {
		return func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Error("panic, expected panic got nil")
				}
			}()
			// len for rng is 10
			rng := [][2]float64{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 9}}
			cut(&rng, start, end)
		}
	}

	t.Run("bad start", test(-1, 9))
	t.Run("bad end", test(5, -1))
	t.Run("bad start bigger then ring", test(11, 7))
	t.Run("bad end bigger then ring", test(1, 11))

}

func TestCut(t *testing.T) {
	type tcase struct {
		ring   [][2]float64
		start  int
		end    int
		rng    [][2]float64
		sliver [][2]float64
	}

	fn := func(t *testing.T, tc tcase) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Panic'd with err %v", r)
			}
		}()
		rng := make([][2]float64, len(tc.ring))
		copy(rng, tc.ring)
		sliver := cut(&rng, tc.start, tc.end)
		if !reflect.DeepEqual(tc.rng, rng) {
			t.Errorf("ring, \n\texpected %v\n\t got     %v", tc.rng, rng)
		}
		if !reflect.DeepEqual(tc.sliver, sliver) {
			t.Errorf("sliver, expected %v got %v", tc.sliver, sliver)
		}
	}

	rings := [...][][2]float64{
		{
			{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 9},
			{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9},
		},
	}

	tests := [...]tcase{
		{
			ring:  rings[0],
			start: 3,
			end:   9,
			rng: [][2]float64{
				{0, 0}, {0, 1}, {0, 2}, {0, 9},
				{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9},
			},
			sliver: [][2]float64{
				{0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8},
			},
		},
		{
			ring:  rings[0],
			start: 9,
			end:   2,
			rng: [][2]float64{
				{0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8},
			},
			sliver: [][2]float64{
				{0, 9}, {1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9},
				{0, 0}, {0, 1},
			},
		},
		{
			ring:  rings[0],
			start: 9,
			end:   9,
			rng: [][2]float64{
				{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8},
				{1, 0}, {1, 1}, {1, 2}, {1, 3}, {1, 4}, {1, 5}, {1, 6}, {1, 7}, {1, 8}, {1, 9},
			},
			sliver: [][2]float64{
				{0, 9},
			},
		},
	}
	for i, tc := range tests {
		tc := tc
		t.Run(strconv.Itoa(i), func(t *testing.T) { fn(t, tc) })
	}

}
