package backup

import (
	"fmt"

	"github.com/d2r2/go-logger"
)

var LocalLog = logger.NewPackageLogger("backup",
	// logger.DebugLevel,
	logger.InfoLevel,
)

var e = fmt.Errorf
var f = fmt.Sprintf
