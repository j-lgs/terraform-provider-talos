package talos

import (
	"time"
)

// Helper global literals defining constants used in test VM instantiation.
var (
	testNworkers int    = 1
	testNcontrol int    = 3
	testGlobalTimeout time.Duration = 30 * time.Minute
)
