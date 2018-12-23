package main

import (
	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/ui/gtkui"
	"github.com/d2r2/gotk3/gtk"
	"github.com/d2r2/gotk3/libnotify"
)

var lg = logger.NewPackageLogger("main",
	//logger.DebugLevel,
	logger.InfoLevel,
)

// contain version+buildnum
// initialized with option:
// -ldflags "-X main.version `head -1 version` -X main.buildnum `date -u +%Y%m%d%H%M%S`"
var (
	buildnum string
	version  string
)

func main() {

	lg.Debugf("Version=%v", version)
	lg.Debugf("Build number=%v", buildnum)
	core.SetVersion(version)
	core.SetBuildNum(buildnum)

	locale.SetLanguage("")
	err := libnotify.Init(core.GetAppTitle())
	if err != nil {
		lg.Fatal(err)
	}

	gtk.Init(nil)
	app, err := gtkui.CreateApp()
	if err != nil {
		lg.Fatal(err)
	}

	app.Run([]string{})

	libnotify.Uninit()

}
