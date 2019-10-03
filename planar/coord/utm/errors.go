package utm

import "github.com/gdey/errors"

const (
	// ErrInvalidZone will be return if the given zone is invalid
	ErrInvalidZone = errors.String("zone is invalid")
	// ErrLatitudeOutOfRange will be returned if the latitude is not in the correct range of acceptable values
	ErrLatitudeOutOfRange = errors.String("latitude out of range")
)
