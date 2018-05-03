package spherical

import "testing"

// TODO(gdey): Need to add real tests.

func TestSemimajor(t *testing.T) {
	hs := Haversine{}
	sm := hs.semimajor()
	if sm != wgs84Semimajor {
		t.Errorf("seimimajor, expected %v got %v", wgs84Semimajor, sm)
	}
	f := 1.0
	hs.Semimajor = &f
	sm = hs.semimajor()
	if sm != 1.0 {
		t.Errorf("seimimajor, expected %v got %v", 1.0, sm)
	}
}
