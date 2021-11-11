package geojson

import (
	"fmt"
)

type ErrMissingField string

func (err ErrMissingField) Error() string {
	return fmt.Sprintf("missing geojson field '%v'", string(err))
}

func (err ErrMissingField) Is(target error) bool {
	mf, ok := target.(ErrMissingField)
	if !ok {
		return false
	}
	return string(mf) == string(err)
}
