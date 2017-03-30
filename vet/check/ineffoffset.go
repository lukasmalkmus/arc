package check

import "github.com/LukasMa/arc/ast"

// Ineffoffset checks if there are any useless "zero offsets" ([%r1 + 0]).
type Ineffoffset struct {
	name string
}

func init() {
	Register(&Ineffoffset{"ineffoffset"})
}

// Desc returns a description of the Check.
func (c Ineffoffset) Desc() string {
	return "checks for useless \"zero offsets\" ([%r1 + 0])"
}

// Name returns the name of the Check.
func (c Ineffoffset) Name() string {
	return c.name
}

// Run executes the Check. It implements the Check interface.
func (c *Ineffoffset) Run(prog *ast.Program) ([]string, error) {
	return nil, nil
}
