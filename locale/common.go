package locale

import (
	"fmt"

	"github.com/d2r2/go-logger"
)

var lg = logger.NewPackageLogger("i18n",
	// logger.DebugLevel,
	logger.InfoLevel,
)

var e = fmt.Errorf
var f = fmt.Sprintf
