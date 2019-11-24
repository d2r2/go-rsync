// +build !gorsync_rel

package data

import (
	"net/http"
)

// Assets contains project assets.
// Depending on "gorsync_rel" build constraint tag take data
// from file system folder, either from embedded code.
var Assets http.FileSystem = http.Dir("data/assets")
