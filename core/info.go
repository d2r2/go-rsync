package core

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/d2r2/go-rsync/locale"
	"github.com/davecgh/go-spew/spew"
)

var (
	APP_ID                = "org.d2r2.gorsync"
	SETTINGS_ID           = APP_ID + ".Settings"
	SETTINGS_PROFILE_ID   = SETTINGS_ID + ".Profile"
	SETTINGS_PROFILE_PATH = "/org/d2r2/gorsync/profiles/%s/"
	SETTINGS_SOURCE_ID    = SETTINGS_PROFILE_ID + ".Source"
	SETTINGS_SOURCE_PATH  = "/org/d2r2/gorsync/profiles/%s/sources/%s/"
)

// contain version+buildnum
// initialized with option:
// -ldflags "-X main.version `head -1 version` -X main.buildnum `date -u +%Y%m%d%H%M%S`"
var (
	_buildnum string
	_version  string
)

func SetVersion(version string) {
	_version = version
}

func SetBuildNum(buildnum string) {
	_buildnum = buildnum
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

func GetAppVersion() string {
	return spew.Sprintf("v%s", _version)
}

func GetAppArchitecture() string {
	return runtime.GOARCH
}

func GetGolangVersion() string {
	return runtime.Version()
}

func GetAppTitle() string {
	return "Gorsync Backup"
}

func GetAppExtraTitle() string {
	return locale.T(MsgAppTitleExtra, nil)
}

func GetAppFullTitle() string {
	appTitle := GetAppTitle()
	appTitleExtra := GetAppExtraTitle()
	if appTitleExtra != "" {
		appTitle += " " + appTitleExtra
	}
	return appTitle
}
