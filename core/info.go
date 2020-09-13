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

package core

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/d2r2/go-rsync/locale"
	"github.com/davecgh/go-spew/spew"
)

// AppRunMode signify what happens, when app
// will be closed. With this type standard
// behavior might be changed and app will be
// started again.
type AppRunMode int

const (
	// AppRegularRun - regular run, app will be closed on exit.
	AppRegularRun AppRunMode = iota
	// AppRunReload - app will be reinitialized and restarted again.
	// This behavior allow to automatically restart app when some
	// settings change require app to reload.
	AppRunReload
)

// contain version+buildnum
// initialized with option:
// -ldflags "-X main.version `head -1 version` -X main.buildnum `date -u +%Y%m%d%H%M%S`"
var (
	_buildnum   string
	_version    string
	_appRunMode AppRunMode
)

// SetVersion save application version provided with compile via -ldflags CLI parameter.
func SetVersion(version string) {
	_version = version
}

// SetBuildNum save application build number provided with compile via -ldflags CLI parameter.
func SetBuildNum(buildnum string) {
	_buildnum = buildnum
}

func SetAppRunMode(appRunMode AppRunMode) {
	_appRunMode = appRunMode
}

// Pass in parameter datetime
// from bash expression `date -u +%y%m%d%H%M%S`.
func generateBuildNum() string {
	if _, err := strconv.Atoi(_buildnum); err == nil && len(_buildnum) == 14 {
		year, _ := strconv.Atoi(_buildnum[0:4])
		month, _ := strconv.Atoi(_buildnum[4:6])
		day, _ := strconv.Atoi(_buildnum[6:8])
		hour, _ := strconv.Atoi(_buildnum[8:10])
		min, _ := strconv.Atoi(_buildnum[10:12])
		sec, _ := strconv.Atoi(_buildnum[12:])
		tm := time.Date(year, time.Month(month), day, hour, min, sec, 0, time.Local)
		tm2 := time.Date(2010, time.January, 1, 0, 0, 0, 0, time.Local)
		return fmt.Sprintf("%d", (tm.Unix()-tm2.Unix())/30)
	}
	return _buildnum
}

func GetAppRunMode() AppRunMode {
	return _appRunMode
}

// GetAppVersion returns string representation of application version.
func GetAppVersion() string {
	return spew.Sprintf("v%s", _version)
}

// GetAppArchitecture1 returns application architecture.
func GetAppArchitecture() string {
	return runtime.GOARCH
}

// GetGolangVersion returns golang version used to compile application.
func GetGolangVersion() string {
	return runtime.Version()
}

// GetAppTitle returns application non-translatable title.
func GetAppTitle() string {
	return "Gorsync Backup"
}

// GetAppExtraTitle returns application translatable extra title.
func GetAppExtraTitle() string {
	return locale.T(MsgAppTitleExtra, nil)
}

// GetAppFullTitle returns application full title.
func GetAppFullTitle() string {
	appTitle := GetAppTitle()
	appTitleExtra := GetAppExtraTitle()
	if appTitleExtra != "" {
		appTitle += " " + appTitleExtra
	}
	return appTitle
}
