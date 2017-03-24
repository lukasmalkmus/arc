package check

// Check is the interface that must implemented by checks.
type Check interface {
	// Run will execute the given check and return a slice of results. An error
	// is returned if the check fails.
	Run() ([]string, error)
}

var checks = make(map[string]Check)

// Register makes a check available by the provided name. If Register is called
// twice with the same name or if check is nil, it panics.
func Register(name string, check Check) {
	if check == nil {
		panic("check: Register check is nil")
	}
	if _, dup := checks[name]; dup {
		panic("check: Register called twice for check " + name)
	}
	checks[name] = check
}

// Checks returns a map of the registered checks.
func Checks() map[string]Check {
	return checks
}
