package rsync

import (
	"fmt"

	"github.com/d2r2/go-logger"
)

var lg = logger.NewPackageLogger("rsync",
	//logger.DebugLevel,
	logger.InfoLevel,
)

var e = fmt.Errorf
var f = fmt.Sprintf
