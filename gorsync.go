//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2020 Denis Dyakov <denis.dyakov@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	"github.com/d2r2/go-rsync/ui/gtkui"
	"github.com/d2r2/gotk3/libnotify"
)

const (
	MsgMainAppSubsystemInitialized = "MainAppSubsystemInitialized"
	MsgMainAppExitedNormally       = "MainAppExitedNormally"
)

// You can manage verbosity of log output
// in the package by changing last parameter value
// (comment/uncomment corresponding lines).
var lg = logger.NewPackageLogger("main",
	// logger.DebugLevel,
	logger.InfoLevel,
)

// Contain version+buildnum initialized with option:
// -ldflags "-X main.version `head -1 version` -X main.buildnum `date -u +%Y%m%d%H%M%S`"
var (
	buildnum string
	version  string
)

// main entry
func main() {
	lg.Debugf("Version=%v", version)
	lg.Debugf("Build number=%v", buildnum)
	// Save application version provided in compilation time.
	core.SetVersion(version)
	core.SetBuildNum(buildnum)

	var cpuprofile string
	flag.StringVar(&cpuprofile, "cpuprofile", "", `Write cpu profile to "file" for debugging purpose.
Generate CPU profile for debugging. Use command "go tool pprof --pdf <path to binary exec> ./cpu.pprof > ./profile.pdf"
to create execution graph in pdf document.`)
	var memprofile string
	flag.StringVar(&memprofile, "memprofile", "", `Write memory profile to "file" for debugging purpose.
Generate memory profile for debugging. Use command "go tool pprof --pdf <path to binary exec> ./mem.pprof > ./profile.pdf"
to create memory usage graph in pdf document.`)
	var versionFlag bool
	flag.BoolVar(&versionFlag, "version", false, `Print environment and version information.`)

	flag.Parse()

	// Activate cpu profiling to trace cpu consumption for debugging purpose.
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			lg.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close()
		if err := pprof.StartCPUProfile(f); err != nil {
			lg.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Print application version information.
	if versionFlag {
		localizer := locale.CreateLocalizer("EN")
		var b bytes.Buffer
		b.WriteString(fmt.Sprintf("\t%s %s\n", core.GetAppFullTitle(), core.GetAppVersion()))
		version, protocol, err := rsync.GetRsyncVersion()
		if err != nil {
			if rsync.IsExtractVersionAndProtocolError(err) {
				version = "?"
				protocol = version
				lg.Warn(err)
			} else {
				lg.Fatal(err)
			}
		}
		b.WriteString("\t" + localizer.Translate(gtkui.MsgRsyncInfo, struct{ RSYNCDetectedVer, RSYNCDetectedProtocol string }{
			RSYNCDetectedVer: version, RSYNCDetectedProtocol: protocol}) + "\n")
		b.WriteString("\t" + localizer.Translate(gtkui.MsgGolangInfo, struct{ GolangVersion, AppArchitecture string }{
			GolangVersion:   core.GetGolangVersion(),
			AppArchitecture: core.GetAppArchitecture()}) + "\n")
		print(b.String())
		os.Exit(0)
	}

	// Initialize language by default; later it
	// might be reinitialized from application preferences.
	locale.SetLanguage("")

	// Initialize libnotify subsystem.
	err := libnotify.Init(core.GetAppTitle())
	if err != nil {
		lg.Fatal(err)
	}
	lg.Info(locale.T(MsgMainAppSubsystemInitialized,
		struct{ Subsystem string }{Subsystem: "Libnotify"}))

	for {
		// Create application.
		app, err := gtkui.CreateApp()
		if err != nil {
			lg.Fatal(err)
		}

		// Run application.
		app.Run([]string{})

		// If request was made to reload app, then we re-run app
		// without exiting (can be used for changing app UI language).
		if core.GetAppRunMode() == core.AppRegularRun {
			break
		} else if core.GetAppRunMode() == core.AppRunReload {
			core.SetAppRunMode(core.AppRegularRun)
		}
	}

	// Uninitialize libnotify subsystem on application exit.
	libnotify.Uninit()

	// Save memory profile to investigate leaked memory.
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			lg.Fatal("could not create memory profile: ", err)
		}
		defer f.Close()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			lg.Fatal("could not write memory profile: ", err)
		}
	}

	// Say goodbye.
	lg.Info(locale.T(MsgMainAppExitedNormally, nil))
}
