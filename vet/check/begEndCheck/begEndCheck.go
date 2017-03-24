package begEndCheck

import "github.com/LukasMa/arc/vet/check"

// Check implements the check.Check interface.
type Check struct{}

func init() {
	check.Register("beginEndCheck", &Check{})
}

// Run executes the check. It implements check.Check.
func (c *Check) Run() ([]string, error) {
	return nil, nil
}
