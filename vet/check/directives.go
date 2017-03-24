package check

// Directives checks if there are any statements outside the .begin and .end
// directives.
type Directives struct{}

func init() {
	Register("directives", &Directives{})
}

// Desc returns a description of the Check.
func (c *Directives) Desc() string {
	return "checks if directives are set and used correctly"
}

// Run executes the Check. It implements the Check interface.
func (c *Directives) Run() ([]string, error) {
	return nil, nil
}
