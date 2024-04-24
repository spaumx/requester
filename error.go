package requester

import (
	"fmt"
)

type Error struct {
	Op  string
	Err string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Op, e.Err)
}
