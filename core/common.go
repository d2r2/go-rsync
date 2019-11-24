package core

import (
	"fmt"

	logger "github.com/d2r2/go-logger"
)

// You can manage verbosity of log output
// in the package by changing last parameter value
// (comment/uncomment corresponding lines).
var lg = logger.NewPackageLogger("core",
	// logger.DebugLevel,
	logger.InfoLevel,
)

var e = fmt.Errorf
var f = fmt.Sprintf
