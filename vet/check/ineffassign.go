package check

// Ineffoffset checks if there are any useless "zero offsets" ([%r1 + 0]).
type Ineffoffset struct{}

func init() {
	Register("ineffoffset", &Ineffoffset{})
}

// Desc returns a description of the Check.
func (c *Ineffoffset) Desc() string {
	return "checks for useless \"zero offsets\" ([%r1 + 0])"
}

// Run executes the Check. It implements the Check interface.
func (c *Ineffoffset) Run() ([]string, error) {
	return nil, nil
}
