package geom

import "errors"

// ErrNilMultiLineStringZMS is thrown when MultiLineStringZMS is nil but shouldn't be
var ErrNilMultiLineStringZMS = errors.New("geom: nil MultiLineStringZMS")

// MultiLineStringZMS is a geometry with multiple LineStringSs.
type MultiLineStringZMS struct {
	Srid
	Mlszm MultiLineStringZM
}

// LineStrings returns the coordinates for the linestrings
func (mlszms MultiLineStringZMS) MultiLineStringZMs() struct {
	Srid
	Mlszm MultiLineStringZM
} {
	return mlszms
}

// SetSRID modifies the struct containing the SRID int and the array of 3D+1 coordinates
func (mlszms *MultiLineStringZMS) SetSRID(srid uint32, mlszm MultiLineStringZM) (err error) {
	if mlszms == nil {
		return ErrNilMultiLineStringZMS
	}

	mlszms.Srid = Srid(srid)
	mlszms.Mlszm = mlszm
	return
}

// Get the simple 3D+1 multiline string
func (mlszms MultiLineStringZMS) MultiLineStringZM() MultiLineStringZM {
	return mlszms.Mlszm
}
