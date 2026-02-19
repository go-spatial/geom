package parsing

import "fmt"

type ErrAt struct {
	Err error
	Pos Position
}

func (err ErrAt) Unwrap() error { return err.Err }
func (err ErrAt) Error() string {
	return fmt.Sprintf("error at %s: %s", err.Pos.String(), err.Err)
}
