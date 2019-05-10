package locale

import (
	"fmt"

	logger "github.com/d2r2/go-logger"
)

var lg = logger.NewPackageLogger("locale",
	// logger.DebugLevel,
	logger.InfoLevel,
)

var e = fmt.Errorf
var f = fmt.Sprintf
