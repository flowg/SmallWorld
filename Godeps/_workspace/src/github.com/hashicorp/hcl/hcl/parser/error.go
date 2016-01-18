package parser

import (
	"fmt"

	"github.com/flowg/go_experiences/server/Godeps/_workspace/src/github.com/hashicorp/hcl/hcl/token"
)

// PosError is a parse error that contains a position.
type PosError struct {
	Pos token.Pos
	Err error
}

func (e *PosError) Error() string {
	return fmt.Sprintf("At %s: %s", e.Pos, e.Err)
}