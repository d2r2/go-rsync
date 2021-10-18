//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2022 Denis Dyakov <denis.dyakov@gma**.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package gtkui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
)

const (
	STOCK_WARNING_ICON = "dialog-warning-symbolic"
	//STOCK_WARNING_ICON = "dialog-warning"
	STOCK_OK_ICON            = "emblem-ok-symbolic"
	STOCK_QUESTION_ICON      = "dialog-question-symbolic"
	STOCK_SYNCHRONIZING_ICON = "emblem-synchronizing-symbolic"
	// ASSET_SYNCHRONIZING_CYAN_ICON   = "emblem-synchronizing-cyan.gif"
	// ASSET_SYNCHRONIZING_YELLOW_ICON = "emblem-synchronizing-yellow.gif"
	// ASSET_SYNCHRONIZING_ICON        = ASSET_SYNCHRONIZING_CYAN_ICON
	// ASSET_SYNCHRONIZING_ANIMATED_64x64_ICON = "loading_animated_64x64.gif"
	// ASSET_SYNCHRONIZING_FRAME8_64x64_ICON   = "loading_frame%d_64x64.gif"
	STOCK_IMPORTANT_ICON = "emblem-important-symbolic"
	// ASSET_IMPORTANT_ICON     = "emblem-important-red.gif"
	STOCK_NETWORK_ERROR_ICON = "network-error-symbolic"
	STOCK_DELETE_ICON        = "edit-delete-symbolic"
)

// Return error describing issue with conversion from one type to another.
func validatorConversionError(fromType, toType string) error {
	msg := spew.Sprintf("Can't convert %[1]v to %[2]v", fromType, toType)
	err := errors.New(msg)
	return err
}

const (
	DesignIndentCol     = 0
	DesignFirstCol      = 4
	DesignSecondCol     = 5
	DesignTotalColCount = 6
)

// GeneralPreferencesNew create preference dialog with "General" page, where controls
// being bound to GLib setting object to save/restore functionality.
func GeneralPreferencesNew(win *gtk.ApplicationWindow, appSettings *SettingsStore,
	actions *glib.ActionMap, prefRow *PreferenceRow) (*gtk.Container, error) {

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 18)

	if prefRow != nil {
		prefRow.Page = &box.Container
	}

	bh := appSettings.NewBindingHelper()

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	row := 0

	// ---------------------------------------------------------
	// Interface options block
	// ---------------------------------------------------------
	markup := NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgGeneralUserInterfaceOptionsSecion, nil), "")
	lbl, err := SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Option to show about dialog on application startup
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgDoNotShowAtAppStartupCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err := gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbAboutInfo, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbAboutInfo.SetActive(!cbAboutInfo.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbAboutInfo.SetTooltipText(locale.T(MsgPrefDlgDoNotShowAtAppStartupHint, nil))
	cbAboutInfo.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_DONT_SHOW_ABOUT_ON_STARTUP, cbAboutInfo, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbAboutInfo, DesignSecondCol, row, 1, 1)
	row++

	// Show desktop notification on backup completion
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgPerformDesktopNotificationCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbPerformBackupCompletionDesktopNotification, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbPerformBackupCompletionDesktopNotification.SetActive(!cbPerformBackupCompletionDesktopNotification.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbPerformBackupCompletionDesktopNotification.SetTooltipText(locale.T(MsgPrefDlgPerformDesktopNotificationHint, nil))
	cbPerformBackupCompletionDesktopNotification.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_PERFORM_DESKTOP_NOTIFICATION, cbPerformBackupCompletionDesktopNotification,
		"active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbPerformBackupCompletionDesktopNotification, DesignSecondCol, row, 1, 1)
	row++

	// UI Language
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgLanguageCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	values := []struct{ value, key string }{
		{locale.T(MsgPrefDlgDefaultLanguageEntry, nil), ""},
		{"English", "en"},
		{"Русский", "ru"},
	}
	cbUILanguage, err := CreateNameValueCombo(values)
	if err != nil {
		return nil, err
	}
	cbUILanguage.SetTooltipText(locale.T(MsgPrefDlgLanguageHint, nil))
	bh.Bind(CFG_UI_LANGUAGE, cbUILanguage, "active-id", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbUILanguage, DesignSecondCol, row, 1, 1)
	initialLang := cbUILanguage.GetActiveID()
	const restartServiceActivationMs = 500
	restartServiceActivationTimer := time.AfterFunc(time.Millisecond*restartServiceActivationMs, func() {
		MustIdleAdd(func() {
			activate := initialLang != cbUILanguage.GetActiveID()
			// Show "restart app" panel only when language has changed
			// from original setting. Otherwise - hide panel.
			err := prefRow.ActivateRestartService(activate)
			if err != nil {
				lg.Fatal(err)
			}
		})
	})
	_, err = cbUILanguage.Connect("changed", func(v *gtk.ComboBox, tmr *time.Timer) {
		RestartTimer(tmr, restartServiceActivationMs)
	}, restartServiceActivationTimer)
	if err != nil {
		return nil, err
	}
	row++

	// Session log font size
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgSessionLogControlFontSizeCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	values = []struct{ value, key string }{
		{"10 px", "10px"},
		{"12 px", "12px"},
		{"13 px", "13px"},
		{"14 px", "14px"},
		{"16 px", "16px"},
		{"18 px", "18px"},
		{"20 px", "20px"},
	}
	cbSessionLogFontSize, err := CreateNameValueCombo(values)
	if err != nil {
		return nil, err
	}
	cbSessionLogFontSize.SetTooltipText(locale.T(MsgPrefDlgSessionLogControlFontSizeHint, nil))
	bh.Bind(CFG_SESSION_LOG_WIDGET_FONT_SIZE, cbSessionLogFontSize, "active-id", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbSessionLogFontSize, DesignSecondCol, row, 1, 1)
	row++

	sep, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	SetAllMargins(&sep.Widget, 6)
	grid.Attach(sep, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// ---------------------------------------------------------
	// Backup settings block
	// ---------------------------------------------------------
	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgGeneralBackupSettingsSection, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Ignore file signature
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgSkipFolderBackupFileSignatureCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	edIgnoreFile, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edIgnoreFile.SetHExpand(true)
	edIgnoreFile.SetTooltipText(locale.T(MsgPrefDlgSkipFolderBackupFileSignatureHint, nil))
	bh.Bind(CFG_IGNORE_FILE_SIGNATURE, edIgnoreFile, "text", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(edIgnoreFile, DesignSecondCol, row, 1, 1)
	row++

	if prefRow != nil {
		rsBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
		if err != nil {
			return nil, err
		}
		err = AddStyleClass(&rsBox.Widget, "info-panel")
		if err != nil {
			return nil, err
		}
		infoBox, err := createBoxWithThemedIcon(STOCK_IMPORTANT_ICON,
			[]string{"image-information", "image-shake"})
		if err != nil {
			return nil, err
		}
		rsBox.Add(infoBox)
		lblRestart, err := SetupLabelJustifyLeft("")
		if err != nil {
			return nil, err
		}
		lblRestart.SetMarkup(locale.T(MsgPrefDlgRestartPanelCaptionWithLink, nil))
		_, err = lblRestart.Connect("activate-link", func(v *gtk.Label, href string) {
			if href == "restart_uri" {
				MustIdleAdd(func() {
					core.SetAppRunMode(core.AppRunReload)
					actionName := "QuitAction"
					action := actions.LookupAction(actionName)
					if action == nil {
						err := errors.New(locale.T(MsgActionDoesNotFound,
							struct{ ActionName string }{ActionName: actionName}))
						lg.Fatal(err)
					}
					// close preference dialog window
					win.Close()
					// activate application quit action
					action.Activate(nil)
				})
			}
		})
		if err != nil {
			return nil, err
		}
		rsBox.Add(lblRestart)

		rvl, err := gtk.RevealerNew()
		if err != nil {
			return nil, err
		}
		rvl.Add(rsBox)
		prefRow.RestartService = &RestartService{Revealer: rvl}
		// box.PackStart(rvl, false, false, 0)
		box.Add(rvl)
	}

	box.Add(grid)

	_, err = box.Connect("destroy", func(b *gtk.Box) {
		bh.Unbind()
	})
	if err != nil {
		return nil, err
	}

	return &box.Container, nil
}

// GetSubpathNotAllowedCharsNotFoundRegexp implement path expression primitive validation on the level
// of lexical parcing. Understand path separator for different OS, taking path separator setting from runtime.
//
// Use Microsoft Windows restriction character list taken from here:
// https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
//
// Linux/Unix forbidden chars for folder names:
// / (forward slash)
//
// Windows forbidden chars for folder names:
// < (less than)
// > (greater than)
// : (colon - sometimes works, but is actually NTFS Alternate Data Streams)
// " (double quote)
// / (forward slash)
// \ (backslash)
// | (vertical bar or pipe)
// ? (question mark)
// * (asterisk)
//
func GetSubpathNotAllowedCharsNotFoundRegexp() (*regexp.Regexp, error) {
	template := spew.Sprintf(`^\%[1]c?([^\<\>\:\"\|\?\*\%[1]c]+\%[1]c?)*$`, os.PathSeparator)
	lg.Debugf("Subpath regex template: %s", template)
	rexp, err := regexp.Compile(template)
	if err != nil {
		return nil, err
	}
	return rexp, nil
}

func GetFolderNamesEmptyOrLeadingTrailingSpacesFoundRegexp() (*regexp.Regexp, error) {
	template := spew.Sprintf(`%[1]c\s+|^\s+|\s+%[1]c|\s+$|^$`, os.PathSeparator)
	lg.Debugf("Subpath regex template: %s", template)
	rexp, err := regexp.Compile(template)
	if err != nil {
		return nil, err
	}
	return rexp, nil
}

func createBackupSourceBlock(profileID, sourceID string, sourceSettings *SettingsStore,
	prefRow *PreferenceRow, validator *UIValidator) (*gtk.Container, error) {

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 12)

	bh := sourceSettings.NewBindingHelper()

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_FILL)
	row := 0

	// Source RSYNC path
	markup := NewMarkup(MARKUP_WEIGHT_NORMAL, 0, 0,
		locale.T(MsgPrefDlgSourceRsyncPathCaption, nil), "")
	lbl, err := SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	edRsyncPath, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edRsyncPath.SetHExpand(true)
	edRsyncPath.SetIconTooltipText(gtk.ENTRY_ICON_SECONDARY, locale.T(MsgPrefDlgSourceRsyncPathRetryHint, nil))

	grid.Attach(edRsyncPath, 1, row, 1, 1)
	row++

	// Destination root path
	markup = NewMarkup(MARKUP_WEIGHT_NORMAL, 0, 0,
		locale.T(MsgPrefDlgDestinationSubpathCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	edDestSubpath, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edDestSubpath.SetTooltipText(locale.T(MsgPrefDlgDestinationSubpathHint, nil))
	grid.Attach(edDestSubpath, 1, row, 1, 1)
	row++

	// Override RSYNC transfer options
	expOverrideRsyncTransferOptions, err := gtk.ExpanderNew(
		locale.T(MsgPrefDlgOverrideRsyncTransferOptionsBoxCaption, nil))
	if err != nil {
		return nil, err
	}
	expOverrideRsyncTransferOptions.SetTooltipText(
		locale.T(MsgPrefDlgOverrideRsyncTransferOptionsBoxHint, nil))
	grid.Attach(expOverrideRsyncTransferOptions, 0, row, 2, 1)
	row++

	box4, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetMargins(box4, 0, 9, 0, 9)
	expOverrideRsyncTransferOptions.Add(box4)

	frame2, err := gtk.FrameNew("")
	if err != nil {
		return nil, err
	}
	box4.PackStart(frame2, true, true, 0)

	box5, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetMargins(box5, 18, 9, 18, 9)
	frame2.Add(box5)

	grid3, err := gtk.GridNew()
	grid3.SetColumnSpacing(12)
	grid3.SetRowSpacing(6)
	grid3.SetHAlign(gtk.ALIGN_FILL)
	box5.PackStart(grid3, true, true, 0)
	row3 := 0

	// Enable/disable RSYNC transfer source owner
	cbTransferSourceOwner, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourceOwner.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourceOwnerCaption, nil))
	cbTransferSourceOwner.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourceOwnerHint, nil))
	cbTransferSourceOwner.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_OWNER_INCONSISTENT, cbTransferSourceOwner, "inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_OWNER, cbTransferSourceOwner, "active", glib.SETTINGS_BIND_DEFAULT)

	cbTransferSourceOwnerHandlerEnabled := true
	_, err = cbTransferSourceOwner.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbTransferSourceOwnerHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbTransferSourceOwnerHandlerEnabled = false
				checkBox.SetActive(false)
				cbTransferSourceOwnerHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbTransferSourceOwner, DesignFirstCol, row3, 1, 1)

	// Enable/disable RSYNC transfer source group
	cbTransferSourceGroup, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourceGroup.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourceGroupCaption, nil))
	cbTransferSourceGroup.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourceGroupHint, nil))
	cbTransferSourceGroup.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_GROUP_INCONSISTENT, cbTransferSourceGroup, "inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_GROUP, cbTransferSourceGroup, "active", glib.SETTINGS_BIND_DEFAULT)

	cbTransferSourceGroupHandlerEnabled := true
	_, err = cbTransferSourceGroup.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbTransferSourceGroupHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbTransferSourceGroupHandlerEnabled = false
				checkBox.SetActive(false)
				cbTransferSourceGroupHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbTransferSourceGroup, DesignSecondCol, row3, 1, 1)
	row3++

	// Enable/disable RSYNC transfer source permissions
	cbTransferSourcePermissions, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourcePermissions.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsCaption, nil))
	cbTransferSourcePermissions.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsHint, nil))
	cbTransferSourcePermissions.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS_INCONSISTENT, cbTransferSourcePermissions,
		"inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS, cbTransferSourcePermissions,
		"active", glib.SETTINGS_BIND_DEFAULT)

	cbTransferSourcePermissionsHandlerEnabled := true
	_, err = cbTransferSourcePermissions.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbTransferSourcePermissionsHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbTransferSourcePermissionsHandlerEnabled = false
				checkBox.SetActive(false)
				cbTransferSourcePermissionsHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbTransferSourcePermissions, DesignFirstCol, row3, 1, 1)

	// Enable/disable RSYNC symlinks recreation
	cbRecreateSymlinks, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbRecreateSymlinks.SetLabel(locale.T(MsgPrefDlgRsyncRecreateSymlinksCaption, nil))
	cbRecreateSymlinks.SetTooltipText(locale.T(MsgPrefDlgRsyncRecreateSymlinksHint, nil))
	cbRecreateSymlinks.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_RECREATE_SYMLINKS_INCONSISTENT, cbRecreateSymlinks, "inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_RECREATE_SYMLINKS, cbRecreateSymlinks, "active", glib.SETTINGS_BIND_DEFAULT)

	cbRecreateSymlinksHandlerEnabled := true
	_, err = cbRecreateSymlinks.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbRecreateSymlinksHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbRecreateSymlinksHandlerEnabled = false
				checkBox.SetActive(false)
				cbRecreateSymlinksHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbRecreateSymlinks, DesignSecondCol, row3, 1, 1)
	row3++

	// Enable/disable RSYNC transfer device files
	cbTransferDeviceFiles, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferDeviceFiles.SetLabel(locale.T(MsgPrefDlgRsyncTransferDeviceFilesCaption, nil))
	cbTransferDeviceFiles.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferDeviceFilesHint, nil))
	cbTransferDeviceFiles.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_DEVICE_FILES_INCONSISTENT, cbTransferDeviceFiles, "inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_TRANSFER_DEVICE_FILES, cbTransferDeviceFiles, "active", glib.SETTINGS_BIND_DEFAULT)

	cbTransferDeviceFilesHandlerEnabled := true
	_, err = cbTransferDeviceFiles.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbTransferDeviceFilesHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbTransferDeviceFilesHandlerEnabled = false
				checkBox.SetActive(false)
				cbTransferDeviceFilesHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbTransferDeviceFiles, DesignFirstCol, row3, 1, 1)

	// Enable/disable RSYNC transfer special files
	cbTransferSpecialFiles, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSpecialFiles.SetLabel(locale.T(MsgPrefDlgRsyncTransferSpecialFilesCaption, nil))
	cbTransferSpecialFiles.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSpecialFilesHint, nil))
	cbTransferSpecialFiles.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SPECIAL_FILES_INCONSISTENT, cbTransferSpecialFiles, "inconsistent", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_RSYNC_TRANSFER_SPECIAL_FILES, cbTransferSpecialFiles, "active", glib.SETTINGS_BIND_DEFAULT)

	cbTransferSpecialFilesHandlerEnabled := true
	_, err = cbTransferSpecialFiles.Connect("clicked", func(checkBox *gtk.CheckButton) {
		if cbTransferSpecialFilesHandlerEnabled {
			if checkBox.GetInconsistent() {
				checkBox.SetInconsistent(false)
			} else if !checkBox.GetInconsistent() && checkBox.GetActive() {
				checkBox.SetInconsistent(true)
				cbTransferSpecialFilesHandlerEnabled = false
				checkBox.SetActive(false)
				cbTransferSpecialFilesHandlerEnabled = true
			}
		}
	})
	if err != nil {
		return nil, err
	}
	grid3.Attach(cbTransferSpecialFiles, DesignSecondCol, row3, 1, 1)
	row3++

	// Extra options
	expExtraOptions, err := gtk.ExpanderNew(locale.T(MsgPrefDlgExtraOptionsBoxCaption, nil))
	if err != nil {
		return nil, err
	}
	expExtraOptions.SetTooltipText(locale.T(MsgPrefDlgExtraOptionsBoxHint, nil))
	grid.Attach(expExtraOptions, 0, row, 2, 1)
	row++

	box2, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetMargins(box2, 0, 9, 0, 9)
	expExtraOptions.Add(box2)

	frame, err := gtk.FrameNew("")
	if err != nil {
		return nil, err
	}
	box2.PackStart(frame, true, true, 0)

	box3, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetMargins(box3, 18, 9, 18, 9)
	frame.Add(box3)

	grid2, err := gtk.GridNew()
	grid2.SetColumnSpacing(12)
	grid2.SetRowSpacing(6)
	grid2.SetHAlign(gtk.ALIGN_FILL)
	box3.PackStart(grid2, true, true, 0)
	row2 := 0

	// Authenticate password
	markup = NewMarkup(MARKUP_WEIGHT_NORMAL, 0, 0,
		locale.T(MsgPrefDlgAuthPasswordCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid2.Attach(lbl, 0, row2, 1, 1)
	edAuthPasswd, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edAuthPasswd.SetTooltipText(locale.T(MsgPrefDlgAuthPasswordHint, nil))
	edAuthPasswd.SetHExpand(true)
	edAuthPasswd.SetInvisibleChar('*')
	edAuthPasswd.SetVisibility(false)
	grid2.Attach(edAuthPasswd, 1, row2, 1, 1)
	row2++

	// Change file permission
	markup = NewMarkup(MARKUP_WEIGHT_NORMAL, 0, 0,
		locale.T(MsgPrefDlgChangeFilePermissionCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid2.Attach(lbl, 0, row2, 1, 1)
	edChmod, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edChmod.SetTooltipText(locale.T(MsgPrefDlgChangeFilePermissionHint, nil))
	edChmod.SetHExpand(true)
	grid2.Attach(edChmod, 1, row2, 1, 1)
	row2++

	// Enable/disable backup block
	markup = NewMarkup(MARKUP_WEIGHT_NORMAL, 0, 0,
		locale.T(MsgPrefDlgEnableBackupBlockCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	swEnabled, err := gtk.SwitchNew()
	if err != nil {
		return nil, err
	}
	swEnabled.SetTooltipText(locale.T(MsgPrefDlgEnableBackupBlockHint, nil))
	swEnabled.SetHAlign(gtk.ALIGN_START)
	grid.Attach(swEnabled, 1, row, 1, 1)
	row++

	// UIValidator object is used to simplify and standardize communication
	// between UI and long running asynchronous processes. For instance, UIValidator
	// helps to run in background RSYNC, which may go on for minutes (in case of
	// network troubles), to verify that data source URL is valid.
	rsyncPathValidatorGroup := "RsyncPath"
	rsyncPathValidatorIndex := spew.Sprintf("%s_%s", profileID, sourceID)
	rsyncPathValidateIndex := validator.AddEntry(rsyncPathValidatorGroup, rsyncPathValidatorIndex,
		// 1st stage of UIValidator. Perform data initialization here, which will be used in next steps.
		// Synchronized call: can update GTK+ widgets from here.
		func(data *ValidatorData, group []*ValidatorData) error {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}
			row, ok := data.Items[2].(*PreferenceRow)
			if !ok {
				return validatorConversionError("ValidatorData.Items[2]", "*PreferenceRow")
			}
			if swtch.GetActive() {
				err = RemoveStyleClassesAll(&entry.Widget)
				if err != nil {
					return err
				}
				entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_SYNCHRONIZING_ICON)
				err = AddStyleClass(&entry.Widget, "entry-image-right-spin")
				if err != nil {
					return err
				}
				err = row.AddStatus(entry.Native(), ProfileStatusValidating, "")
				if err != nil {
					return err
				}
				RsyncSourcePathDescription := locale.T(MsgPrefDlgSourceRsyncPathDescriptionHint, nil)
				markup := markupTooltip(NewMarkup(0, MARKUP_COLOR_SKY_BLUE, 0,
					locale.T(MsgPrefDlgSourceRsyncValidatingHint, nil), nil), RsyncSourcePathDescription)
				entry.SetTooltipMarkup(markup.String())
			}
			return nil
		},
		// 2nd stage of UIValidator. Execute long-running validation processes here.
		// Asynchronous call: doesn't allowed to change GTK+ widgets from here (only read)!
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return nil, validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return nil, validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}

			var warning *string
			if swtch.GetActive() {
				rsyncURL, err := entry.GetText()
				if err != nil {
					return nil, err
				}
				rsyncURL = strings.TrimSpace(rsyncURL)
				lg.Debugf("Validate rsync source: %q", rsyncURL)

				if rsyncURL == "" {
					groupLock.Lock()
					msg := locale.T(MsgPrefDlgSourceRsyncPathEmptyError, nil)
					groupLock.Unlock()
					warning = &msg
				} else {
					lg.Debugf("Start rsync utility to validate rsync source")
					//					sourceSettings, err := getBackupSourceSettings(profileID, sourceID, nil)
					var authPass *string
					ap := sourceSettings.settings.GetString(CFG_MODULE_AUTH_PASSWORD)
					if ap != "" {
						authPass = &ap
					}

					// Start long-running process, where RSYNC is running to validate source path.
					// It can takes minutes.
					err = rsync.GetPathStatus(ctx, authPass, rsyncURL, false)
					// Lock global groupID context to skip race conditions.
					groupLock.Lock()
					if err != nil {
						lg.Debug(err)
						if !rsync.IsProcessTerminatedError(err) {
							msg := err.Error()
							warning = &msg
						}
					}
					groupLock.Unlock()
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage of UIValidator. Final step of data validation.
		// Asynchronous call: can't update GTK+ widgets directly, but only when code is wrapped
		// to glib.IdleAdd method.
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, data *ValidatorData, results []interface{}) error {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}
			row, ok := data.Items[2].(*PreferenceRow)
			if !ok {
				return validatorConversionError("ValidatorData.Items[2]", "*PreferenceRow")
			}
			groupLock.Lock()
			RsyncSourcePathDescription := locale.T(MsgPrefDlgSourceRsyncPathDescriptionHint, nil)
			groupLock.Unlock()
			MustIdleAdd(func() {

				if swtch.GetActive() {
					err := RemoveStyleClass(&entry.Widget, "entry-image-right-spin")
					if err != nil {
						lg.Fatal(err)
					}
					warning, ok := results[0].(*string)
					if !ok {
						lg.Fatal(validatorConversionError("interface{}[0]", "*string"))
					}
					if warning != nil {
						err = AddStyleClasses(&entry.Widget, []string{"entry-image-right-error", "entry-image-right-shake"})
						if err != nil {
							lg.Fatal(err)
						}
						entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_IMPORTANT_ICON)
						markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
							RsyncSourcePathDescription)
						entry.SetTooltipMarkup(markup.String())
						err = row.AddStatus(entry.Native(), ProfileStatusError, *warning)
						if err != nil {
							lg.Fatal(err)
						}
					} else {
						entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_OK_ICON)
						entry.SetTooltipText(RsyncSourcePathDescription)
						err := row.RemoveStatus(entry.Native())
						if err != nil {
							lg.Fatal(err)
						}
					}
				} else {
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
					markup := markupTooltip(NewMarkup(0, 0 /*MARKUP_COLOR_LIGHT_GRAY*/, 0,
						locale.T(MsgPrefDlgSourceRsyncPathNotValidatedHint, nil), nil), RsyncSourcePathDescription)
					entry.SetTooltipMarkup(markup.String())
					err := row.RemoveStatus(entry.Native())
					if err != nil {
						lg.Fatal(err)
					}
				}
			})
			return nil
		}, edRsyncPath, swEnabled, prefRow)

	rsyncPathChangeTimer := time.AfterFunc(time.Millisecond*1000, func() {
		MustIdleAdd(func() {
			err := validator.Validate(rsyncPathValidatorGroup, rsyncPathValidatorIndex)
			if err != nil {
				lg.Fatal(err)
			}
		})
	})
	rsyncPathChangeTimer.Stop()
	_, err = edRsyncPath.Connect("changed", func(v *gtk.Entry) {
		RestartTimer(rsyncPathChangeTimer, 1000)
	})
	if err != nil {
		return nil, err
	}
	_, err = edRsyncPath.Connect("icon-press", func(v *gtk.Entry) {
		RestartTimer(rsyncPathChangeTimer, 50)
	})
	if err != nil {
		return nil, err
	}
	_, err = edRsyncPath.Connect("destroy", func(entry *gtk.Entry) {
		lg.Debug("Destroy edRsyncPath")
		err := prefRow.RemoveStatus(entry.Native())
		if err != nil {
			lg.Fatal(err)
		}
		validator.RemoveEntry(rsyncPathValidateIndex)
	})
	if err != nil {
		return nil, err
	}

	bh.Bind(CFG_MODULE_RSYNC_SOURCE_PATH, edRsyncPath, "text", glib.SETTINGS_BIND_DEFAULT)
	text, err := edRsyncPath.GetText()
	if err != nil {
		return nil, err
	}
	charWidth := utf8.RuneCountInString(text)
	if charWidth > 0 {
		edRsyncPath.SetWidthChars(utf8.RuneCountInString(text))
	}

	rexpSubpathNotAllowedCharsNotFound, err := GetSubpathNotAllowedCharsNotFoundRegexp()
	if err != nil {
		return nil, err
	}
	rexpFolderNamesEmptyOrLeadingTrailingSpacesFound, err :=
		GetFolderNamesEmptyOrLeadingTrailingSpacesFoundRegexp()
	if err != nil {
		return nil, err
	}

	// UIValidator object is used to simplify and standardize communication
	// between UI and long running asynchronous processes. For instance, UIValidator
	// helps to run in background RSYNC, which may go on for minutes (in case of
	// network troubles), to verify that data source URL is valid.
	destSubPathValidatorGroup := "DestSubpath"
	destSubPathValidatorIndex := profileID
	destSubPathValidateIndex := validator.AddEntry(destSubPathValidatorGroup, destSubPathValidatorIndex,
		// 1st stage of UIValidator. Perform data initialization here, which will be used in next steps.
		// Synchronized call: can update GTK+ widgets from here.
		func(data *ValidatorData, group []*ValidatorData) error {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}
			if swtch.GetActive() {
				err := RemoveStyleClassesAll(&entry.Widget)
				if err != nil {
					return err
				}
				entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_SYNCHRONIZING_ICON)
				err = AddStyleClass(&entry.Widget, "entry-image-right-spin")
				if err != nil {
					return err
				}
			}
			return nil
		},
		// 2nd stage of UIValidator. Execute long-running validation processes here.
		// Asynchronous call: doesn't allowed to change GTK+ widgets from here (only read)!
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return nil, validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return nil, validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}
			destSubPath, err := entry.GetText()
			if err != nil {
				return nil, err
			}
			var warning *string
			if swtch.GetActive() && (!rexpSubpathNotAllowedCharsNotFound.MatchString(destSubPath) ||
				rexpFolderNamesEmptyOrLeadingTrailingSpacesFound.MatchString(destSubPath)) {
				groupLock.Lock()
				msg := locale.T(MsgPrefDlgDestinationSubpathExpressionError, nil)
				groupLock.Unlock()
				warning = &msg
			} else {
				foundCollision := false
				lg.Debugf("DestSubPath validation group count = %v", len(group))
				for _, item := range group {
					entry2, ok := item.Items[0].(*gtk.Entry)
					if !ok {
						return nil, validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
					}
					swtch2, ok := item.Items[1].(*gtk.Switch)
					if !ok {
						return nil, validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
					}
					destSubPath2, err := entry2.GetText()
					if err != nil {
						return nil, err
					}
					if entry != entry2 && swtch.GetActive() && swtch2.GetActive() &&
						normalizeSubpath(destSubPath) == normalizeSubpath(destSubPath2) {
						foundCollision = true
						break
					}
				}
				lg.Debugf("DestSubPath collision found = %v", foundCollision)
				if foundCollision {
					groupLock.Lock()
					msg := locale.T(MsgPrefDlgDestinationSubpathNotUniqueError, nil)
					groupLock.Unlock()
					warning = &msg
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage of UIValidator. Final step of data validation.
		// Asynchronous call: can't update GTK+ widgets directly, but only when code is wrapped
		// to glib.IdleAdd method.
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, data *ValidatorData, results []interface{}) error {
			groupLock.Lock()
			destSubpathHint := locale.T(MsgPrefDlgDestinationSubpathHint, nil)
			groupLock.Unlock()

			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			swtch, ok := data.Items[1].(*gtk.Switch)
			if !ok {
				return validatorConversionError("ValidatorData.Items[1]", "*gtk.Switch")
			}
			row, ok := data.Items[2].(*PreferenceRow)
			if !ok {
				return validatorConversionError("ValidatorData.Items[2]", "*PreferenceRow")
			}
			MustIdleAdd(func() {

				if swtch.GetActive() {
					err := RemoveStyleClass(&entry.Widget, "entry-image-right-spin")
					if err != nil {
						lg.Fatal(err)
					}
					warning, ok := results[0].(*string)
					if !ok {
						lg.Fatal(validatorConversionError("interface{}[0]", "*string"))
					}
					if warning != nil {
						err = AddStyleClasses(&entry.Widget, []string{"entry-image-right-error", "entry-image-right-shake"})
						if err != nil {
							lg.Fatal(err)
						}
						entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_IMPORTANT_ICON)
						markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
							destSubpathHint)
						entry.SetTooltipMarkup(markup.String())
						//entry.SetTooltipText(*warning)
						err = row.AddStatus(entry.Native(), ProfileStatusError, *warning)
						if err != nil {
							lg.Fatal(err)
						}
					} else {
						entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_OK_ICON)
						entry.SetTooltipText(destSubpathHint)
						err = row.RemoveStatus(entry.Native())
						if err != nil {
							lg.Fatal(err)
						}
					}
				} else {
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
					markup := markupTooltip(NewMarkup(0, 0 /*MARKUP_COLOR_LIGHT_GRAY*/, 0,
						locale.T(MsgPrefDlgDestinationSubpathNotValidatedHint, nil), nil), destSubpathHint)
					entry.SetTooltipMarkup(markup.String())
					err := row.RemoveStatus(entry.Native())
					if err != nil {
						lg.Fatal(err)
					}
				}
			})
			return nil
		}, edDestSubpath, swEnabled, prefRow)
	destSubpathChangeTimer := time.AfterFunc(time.Millisecond*500, func() {
		MustIdleAdd(func() {
			err := validator.Validate(destSubPathValidatorGroup, destSubPathValidatorIndex)
			if err != nil {
				lg.Fatal(err)
			}
		})
	})
	destSubpathChangeTimer.Stop()
	_, err = edDestSubpath.Connect("changed", func(v *gtk.Entry) {
		RestartTimer(destSubpathChangeTimer, 500)
	})
	if err != nil {
		return nil, err
	}
	_, err = edDestSubpath.Connect("destroy", func(entry *gtk.Entry) {
		lg.Debug("Destroy edDestSubpath")
		err := prefRow.RemoveStatus(entry.Native())
		if err != nil {
			lg.Fatal(err)
		}
		validator.RemoveEntry(destSubPathValidateIndex)
		RestartTimer(destSubpathChangeTimer, 50)
	})
	if err != nil {
		return nil, err
	}

	_, err = edAuthPasswd.Connect("changed", func(v *gtk.Entry) {
		if swEnabled.GetActive() {
			RestartTimer(rsyncPathChangeTimer, 1000)
		}
	})
	if err != nil {
		return nil, err
	}

	bh.Bind(CFG_MODULE_DEST_SUBPATH, edDestSubpath, "text", glib.SETTINGS_BIND_DEFAULT)

	bh.Bind(CFG_MODULE_CHANGE_FILE_PERMISSION, edChmod, "text", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_MODULE_AUTH_PASSWORD, edAuthPasswd, "text", glib.SETTINGS_BIND_DEFAULT)

	// Expand control's block if found that internal settings not in default state.
	expOverrideRsyncTransferOptions.SetExpanded(
		!sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_OWNER_INCONSISTENT) ||
			!sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_GROUP_INCONSISTENT) ||
			!sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS_INCONSISTENT) ||
			!sourceSettings.settings.GetBoolean(CFG_RSYNC_RECREATE_SYMLINKS_INCONSISTENT) ||
			!sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_DEVICE_FILES_INCONSISTENT) ||
			!sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SPECIAL_FILES_INCONSISTENT))

	// Expand control's block if found that internal settings not in default state.
	expExtraOptions.SetExpanded(
		sourceSettings.settings.GetString(CFG_MODULE_AUTH_PASSWORD) != "" ||
			sourceSettings.settings.GetString(CFG_MODULE_CHANGE_FILE_PERMISSION) != "")

	_, err = swEnabled.Connect("state-set", func(v *gtk.Switch) {
		RestartTimer(rsyncPathChangeTimer, 50)
		RestartTimer(destSubpathChangeTimer, 50)
	})
	if err != nil {
		return nil, err
	}
	bh.Bind(CFG_MODULE_ENABLED, swEnabled, "active", glib.SETTINGS_BIND_DEFAULT)

	box.PackStart(grid, true, true, 0)
	box.SetHExpand(true)
	box.SetVExpand(true)

	_, err = box.Connect("destroy", func(b *gtk.Box) {
		lg.Debug("Destroy box")
		bh.Unbind()
		//validator.CancelValidate(rsyncPathGroupName)
	})
	if err != nil {
		return nil, err
	}

	RestartTimer(rsyncPathChangeTimer, 50)
	RestartTimer(destSubpathChangeTimer, 50)

	return &box.Container, nil
}

// getProfileSettings create GlibSettings object with change event
// connected to specific indexed profile[profileID].
func getProfileSettings(appStore *SettingsStore, profileID string, changed func()) (*SettingsStore, error) {
	pathSuffix := fmt.Sprintf(PROFILE_SCHEMA_SUFFIX_PATH, profileID)
	store, err := appStore.GetChildSettingsStore(PROFILE_SCHEMA_SUFFIX_ID, pathSuffix, changed)
	if err != nil {
		return nil, err
	}
	return store, nil
}

// getBackupSourceSettings create GlibSettings object with change event
// connected to specific indexed source[profile[profileID], sourceID].
func getBackupSourceSettings(profileStore *SettingsStore, sourceID string, changed func()) (*SettingsStore, error) {
	path := fmt.Sprintf(SOURCE_SCHEMA_SUFFIX_PATH, sourceID)
	store, err := profileStore.GetChildSettingsStore(SOURCE_SCHEMA_SUFFIX_ID, path, changed)
	if err != nil {
		return nil, err
	}
	return store, nil
}

func createBackupSourceBlock2(win *gtk.ApplicationWindow, profileSettings *SettingsStore,
	profileID, sourceID string, prefRow *PreferenceRow, validator *UIValidator,
	profileChanged func()) (*gtk.Container, error) {

	sourceSettings, err := getBackupSourceSettings(profileSettings, sourceID, profileChanged)
	if err != nil {
		lg.Fatal(err)
	}

	box2, err := createBackupSourceBlock(profileID, sourceID, sourceSettings, prefRow, validator /*, profileChanged*/)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	box31, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	SetMargins(box31, 8, 0, 0, 0)
	box.PackStart(box31, false, false, 0)

	box.PackStart(box2, true, true, 0)

	box32, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	box.PackEnd(box32, false, false, 0)

	srclbr, err := gtk.ListBoxRowNew()
	if err != nil {
		return nil, err
	}
	SetMargins(&srclbr.Widget, 5, 5, 5, 5)
	srclbr.Add(box)

	btnDeleteSource, err := SetupButtonWithThemedImage(STOCK_DELETE_ICON)
	if err != nil {
		return nil, err
	}
	btnDeleteSource.SetVAlign(gtk.ALIGN_START)
	btnDeleteSource.SetHAlign(gtk.ALIGN_CENTER)
	btnDeleteSource.SetTooltipText(locale.T(MsgPrefDlgDeleteBackupBlockHint, nil))
	_, err = btnDeleteSource.Connect("clicked", func(btn *gtk.Button, box *gtk.ListBoxRow) {
		title := locale.T(MsgPrefDlgDeleteBackupBlockDialogTitle, nil)
		titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
			NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
		yesButtonCaption := locale.T(MsgDialogYesButton, nil)
		yesButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, removeUndescore(yesButtonCaption), nil)
		textMarkup := locale.T(MsgPrefDlgDeleteBackupBlockDialogText,
			struct{ YesButton string }{YesButton: yesButtonMarkup.String()})
		responseYes, err := questionDialog(&win.Window, titleMarkup.String(), textMarkup, true, true, false)
		if err != nil {
			lg.Fatal(err)
		}

		if responseYes {
			delete(prefRow.RsyncSources, btnDeleteSource.Native())
			box.Destroy()

			sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
			err = sarr.DeleteNode(sourceSettings, sourceID)
			if err != nil {
				lg.Fatal(err)
			}
			prefRow.EnableDisableDeleteButtonsAndRecalculateIndexes()
		}
	}, srclbr)
	if err != nil {
		return nil, err
	}
	box32.PackStart(btnDeleteSource, false, false, 0)

	lbl, err := SetupLabelMarkupJustifyCenter(nil)
	if err != nil {
		return nil, err
	}
	lbl.SetSensitive(false)
	err = AddStyleClass(&lbl.Widget, "label-index-caption")
	if err != nil {
		return nil, err
	}
	box31.PackStart(lbl, false, false, 0)

	prefRow.RsyncSources[btnDeleteSource.Native()] =
		&RsyncSource{DeleteBtn: btnDeleteSource, IndexLbl: lbl,
			Index: prefRow.GetLastRsyncModuleIndex() + 1}
	prefRow.EnableDisableDeleteButtonsAndRecalculateIndexes()

	return &srclbr.Container, nil
}

// ProfilePreferencesNew create preference dialog with "Sources" page, where controls
// being bound to GLib Setting object to save/restore functionality.
func ProfilePreferencesNew(win *gtk.ApplicationWindow, appSettings *SettingsStore,
	validator *UIValidator, profileID string, prefRow *PreferenceRow,
	profileChanged func(), initProfileName *string) (*gtk.Container, string, error) {

	sw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, "", err
	}
	sw.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	//SetScrolledWindowPropogatedHeight(sw, true)

	frame, err := gtk.FrameNew("")
	if err != nil {
		return nil, "", err
	}
	box0, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, "", err
	}
	SetAllMargins(box0, 0)
	frame.Add(box0)

	srclb, err := gtk.ListBoxNew()
	if err != nil {
		return nil, "", err
	}
	srclb.SetSelectionMode(gtk.SELECTION_NONE)
	srclb.SetHeaderFunc(func(row, before *gtk.ListBoxRow) {
		if before != nil {
			current := row.GetHeader()
			if current == nil {
				sep, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
				if err != nil {
					lg.Fatal(err)
				}
				row.SetHeader(&sep.Widget)
			}
		}
	})
	box0.Add(srclb)

	profileSettings, err := getProfileSettings(appSettings, profileID, profileChanged)
	if err != nil {
		return nil, "", err
	}

	sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)

	for _, srcID := range sarr.GetArrayIDs() {
		cntr, err := createBackupSourceBlock2(win, profileSettings, profileID,
			srcID, prefRow, validator, profileChanged)
		if err != nil {
			return nil, "", err
		}

		srclb.Add(cntr)
	}

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_FILL)
	row := 0

	var lbl *gtk.Label

	appBH := appSettings.NewBindingHelper()
	profileBH := profileSettings.NewBindingHelper()

	// Profile name
	markup := NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgProfileNameCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, "", err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	edProfileName, err := gtk.EntryNew()
	if err != nil {
		return nil, "", err
	}
	edProfileName.SetHExpand(true)
	edProfileName.SetHAlign(gtk.ALIGN_FILL)

	// UIValidator object is used to simplify and standardize communication
	// between UI and long running asynchronous processes. For instance, UIValidator
	// helps to run in background RSYNC, which may go on for minutes (in case of
	// network troubles), to verify that data source URL is valid.
	profileValidatorGroup := "ProfileName"
	profileValidatorIndex := "1"
	profileValidateIndex := validator.AddEntry(profileValidatorGroup, profileValidatorIndex,
		// 1st stage of UIValidator. Perform data initialization here, which will be used in next steps.
		// Synchronized call: can update GTK+ widgets here.
		func(data *ValidatorData, group []*ValidatorData) error {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			err := RemoveStyleClassesAll(&entry.Widget)
			if err != nil {
				return err
			}
			entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_SYNCHRONIZING_ICON)
			err = AddStyleClass(&entry.Widget, "entry-image-right-spin")
			if err != nil {
				return err
			}
			return nil
		},
		// 2nd stage of UIValidator. Execute long-running validation processes here.
		// Asynchronous call: doesn't allowed to change GTK+ widgets here (only read)!
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {

			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return nil, validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			profileName, err := entry.GetText()
			if err != nil {
				return nil, err
			}
			var warning *string
			if profileName == "" {
				groupLock.Lock()
				msg := locale.T(MsgPrefDlgProfileNameEmptyWarning, nil)
				groupLock.Unlock()
				warning = &msg
			} else {
				foundCollision := false
				for _, item := range group {
					entry2, ok := item.Items[0].(*gtk.Entry)
					if !ok {
						return nil, validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
					}
					profileName2, err := entry2.GetText()
					if err != nil {
						return nil, err
					}
					if entry != entry2 && profileName == profileName2 {
						foundCollision = true
						break
					}
				}
				if foundCollision {
					groupLock.Lock()
					msg := locale.T(MsgPrefDlgProfileNameExistsWarning,
						struct{ ProfileName string }{ProfileName: profileName})
					groupLock.Unlock()
					warning = &msg
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage of UIValidator. Final step of data validation.
		// Asynchronous call: can't update GTK+ widgets directly, but only when code is wrapped
		// to glib.IdleAdd method.
		// Use groupLock object, to limit simultaneous access to some not-thread-safe resources.
		func(groupLock *sync.Mutex, data *ValidatorData, results []interface{}) error {
			groupLock.Lock()
			profileNameHint := locale.T(MsgPrefDlgProfileNameHint, nil)
			groupLock.Unlock()
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			row, ok := data.Items[1].(*PreferenceRow)
			if !ok {
				return validatorConversionError("ValidatorData.Items[1]", "*PreferenceRow")
			}
			warning, ok := results[0].(*string)
			if !ok {
				return validatorConversionError("interface{}[0]", "*string")
			}

			MustIdleAdd(func() {
				if warning != nil {
					err := AddStyleClasses(&entry.Widget, []string{"entry-image-right-error", "entry-image-right-shake"})
					if err != nil {
						lg.Fatal(err)
					}
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_IMPORTANT_ICON)
					markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
						profileNameHint)
					entry.SetTooltipMarkup(markup.String())
					err = row.AddStatus(entry.Native(), ProfileStatusError, *warning)
					if err != nil {
						lg.Fatal(err)
					}
				} else {
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
					entry.SetTooltipText(profileNameHint)
					err = row.RemoveStatus(entry.Native())
					if err != nil {
						lg.Fatal(err)
					}
				}
			})

			return nil
		}, edProfileName, prefRow)
	profileBH.Bind(CFG_PROFILE_NAME, edProfileName, "text", glib.SETTINGS_BIND_DEFAULT)
	edProfileNameChangeTimer := time.AfterFunc(time.Millisecond*500, func() {
		MustIdleAdd(func() {
			name, err := edProfileName.GetText()
			if err != nil {
				lg.Fatal(err)
			}
			prefRow.SetName(name)
			err = validator.Validate(profileValidatorGroup, profileValidatorIndex)
			if err != nil {
				lg.Fatal(err)
			}
		})
	})
	_, err = edProfileName.Connect("changed", func(v *gtk.Entry, tmr *time.Timer) {
		RestartTimer(tmr, 500)
	}, edProfileNameChangeTimer)
	if err != nil {
		return nil, "", err
	}
	_, err = edProfileName.Connect("destroy", func(entry *gtk.Entry) {
		validator.RemoveEntry(profileValidateIndex)
		err = validator.Validate(profileValidatorGroup, profileValidatorIndex)
		if err != nil {
			lg.Fatal(err)
		}
	})
	if err != nil {
		return nil, "", err
	}

	if initProfileName != nil {
		edProfileName.SetText(*initProfileName)
	}

	grid.Attach(edProfileName, 1, row, 1, 1)
	row++

	// Destination root path
	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgDefaultDestPathCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, "", err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	destFolder, err := gtk.FileChooserButtonNew("Select default destination folder",
		gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	if err != nil {
		return nil, "", err
	}
	destFolder.SetTooltipText(locale.T(MsgPrefDlgDefaultDestPathHint, nil))
	destFolder.SetHExpand(true)
	destFolder.SetHAlign(gtk.ALIGN_FILL)
	folder := profileSettings.settings.GetString(CFG_PROFILE_DEST_ROOT_PATH)
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		destFolder.SetFilename(folder)
	}
	_, err = destFolder.Connect("file-set", func(fcb *gtk.FileChooserButton) {
		folder := fcb.GetFilename()
		if _, err := os.Stat(folder); !os.IsNotExist(err) {
			profileSettings.settings.SetString(CFG_PROFILE_DEST_ROOT_PATH, folder)
		}
	})
	if err != nil {
		return nil, "", err
	}
	grid.Attach(destFolder, 1, row, 1, 1)
	row++

	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgSourcesCaption, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, "", err
	}
	grid.Attach(lbl, 0, row, 1, 1)

	btnAddSource, err := SetupButtonWithThemedImage("list-add-symbolic")
	if err != nil {
		return nil, "", err
	}
	btnAddSource.SetTooltipText(locale.T(MsgPrefDlgAddBackupBlockHint, nil))
	_, err = btnAddSource.Connect("clicked", func() {
		sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
		sourceID, err := sarr.AddNode()
		if err != nil {
			lg.Fatal(err)
		}

		cntr, err := createBackupSourceBlock2(win, profileSettings, profileID,
			sourceID, prefRow, validator, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}

		srclb.Add(cntr)

		srclb.ShowAll()

		destSubPathValidatorGroup := "DestSubpath"
		destSubPathValidatorIndex := profileID
		err = validator.Validate(destSubPathValidatorGroup, destSubPathValidatorIndex)
		if err != nil {
			lg.Fatal(err)
		}
	})
	if err != nil {
		return nil, "", err
	}

	box2, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, "", err
	}
	SetAllMargins(box2, 18)
	box2.Add(grid)
	box2.Add(frame)
	box2.Add(btnAddSource)

	vp, err := gtk.ViewportNew(nil, nil)
	if err != nil {
		return nil, "", err
	}
	vp.Add(box2)

	sw.Add(vp)
	_, err = sw.Connect("destroy", func(b gtk.IWidget) {
		appBH.Unbind()
		profileBH.Unbind()
	})
	if err != nil {
		return nil, "", err
	}

	name := profileSettings.settings.GetString(CFG_PROFILE_NAME)
	return &sw.Container, name, nil
}

// AdvancedPreferencesNew create preference dialog with "Advanced" page, where controls
// bound to GLib Setting object for save/restore functionality.
func AdvancedPreferencesNew(appSettings *SettingsStore, prefRow *PreferenceRow) (*gtk.Container, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 18)

	if prefRow != nil {
		prefRow.Page = &box.Container
	}

	bh := appSettings.NewBindingHelper()

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	row := 0

	// ---------------------------------------------------------
	// Backup settings block
	// ---------------------------------------------------------
	markup := NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgAdvancedBackupSettingsSection, nil), "")
	lbl, err := SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Enable/disable automatic backup block size
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgAutoManageBackupBlockSizeCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err := gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbAutoManageBackupBlockSize, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbAutoManageBackupBlockSize.SetActive(!cbAutoManageBackupBlockSize.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbAutoManageBackupBlockSize.SetTooltipText(locale.T(MsgPrefDlgAutoManageBackupBlockSizeHint, nil))
	cbAutoManageBackupBlockSize.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE, cbAutoManageBackupBlockSize, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbAutoManageBackupBlockSize, DesignSecondCol, row, 1, 1)
	row++

	// Backup block size
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgBackupBlockSizeCaption, nil))
	if err != nil {
		return nil, err
	}
	bh.Bind(CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE, lbl, "sensitive",
		glib.SETTINGS_BIND_GET|glib.SETTINGS_BIND_INVERT_BOOLEAN)
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	sbBackupBlockSize, err := gtk.SpinButtonNewWithRange(50, 10000, 1)
	if err != nil {
		return nil, err
	}
	sbBackupBlockSize.SetTooltipText(locale.T(MsgPrefDlgBackupBlockSizeHint, nil))
	sbBackupBlockSize.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_MAX_BACKUP_BLOCK_SIZE_MB, sbBackupBlockSize, "value", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE, sbBackupBlockSize, "sensitive",
		glib.SETTINGS_BIND_GET|glib.SETTINGS_BIND_INVERT_BOOLEAN)
	grid.Attach(sbBackupBlockSize, DesignSecondCol, row, 1, 1)
	row++

	// Run notification script on backup completion
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgRunNotificationScriptCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbRunBackupCompletionNotificationScript, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbRunBackupCompletionNotificationScript.SetActive(!cbRunBackupCompletionNotificationScript.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbRunBackupCompletionNotificationScript.SetTooltipText(locale.T(MsgPrefDlgRunNotificationScriptHint, nil))
	cbRunBackupCompletionNotificationScript.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RUN_NOTIFICATION_SCRIPT, cbRunBackupCompletionNotificationScript,
		"active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbRunBackupCompletionNotificationScript, DesignSecondCol, row, 1, 1)
	row++

	sep, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	SetAllMargins(&sep.Widget, 6)
	grid.Attach(sep, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// ---------------------------------------------------------
	// Rsync general block
	// ---------------------------------------------------------
	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgAdvansedRsyncSettingsSection, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Rsync utility retry count
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgRsyncRetryCountCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	sbRetryCount, err := gtk.SpinButtonNewWithRange(0, 5, 1)
	if err != nil {
		return nil, err
	}
	sbRetryCount.SetTooltipText(locale.T(MsgPrefDlgRsyncRetryCountHint, nil))
	sbRetryCount.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_RETRY_COUNT, sbRetryCount, "value", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(sbRetryCount, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC low level log
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgRsyncLowLevelLogCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbLowLevelRsyncLog, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbLowLevelRsyncLog.SetActive(!cbLowLevelRsyncLog.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbLowLevelRsyncLog.SetTooltipText(locale.T(MsgPrefDlgRsyncLowLevelLogHint, nil))
	cbLowLevelRsyncLog.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC, cbLowLevelRsyncLog, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbLowLevelRsyncLog, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC intensive low level log
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgRsyncIntensiveLowLevelLogCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	bh.Bind(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC, eb, "sensitive", glib.SETTINGS_BIND_GET)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbIntensiveLowLevelRsyncLog, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbIntensiveLowLevelRsyncLog.SetActive(!cbIntensiveLowLevelRsyncLog.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbIntensiveLowLevelRsyncLog.SetTooltipText(locale.T(MsgPrefDlgRsyncIntensiveLowLevelLogHint, nil))
	cbIntensiveLowLevelRsyncLog.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_ENABLE_INTENSIVE_LOW_LEVEL_LOG_OF_RSYNC, cbIntensiveLowLevelRsyncLog,
		"active", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC, cbIntensiveLowLevelRsyncLog,
		"sensitive", glib.SETTINGS_BIND_GET)
	grid.Attach(cbIntensiveLowLevelRsyncLog, DesignSecondCol, row, 1, 1)
	row++

	sep, err = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	SetAllMargins(&sep.Widget, 6)
	grid.Attach(sep, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// ---------------------------------------------------------
	// Rsync deduplication block
	// ---------------------------------------------------------
	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgAdvancedRsyncDedupSettingsSection, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Use previous backup if found
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgUsePreviousBackupForDedupCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	eb.Add(lbl)
	grid.Attach(eb, DesignFirstCol, row, 1, 1)
	cbPrevBackupUsage, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	_, err = eb.Connect("button-press-event", func() {
		cbPrevBackupUsage.SetActive(!cbPrevBackupUsage.GetActive())
	})
	if err != nil {
		return nil, err
	}
	cbPrevBackupUsage.SetTooltipText(locale.T(MsgPrefDlgUsePreviousBackupForDedupHint, nil))
	cbPrevBackupUsage.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_ENABLE_USE_OF_PREVIOUS_BACKUP, cbPrevBackupUsage, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbPrevBackupUsage, DesignSecondCol, row, 1, 1)
	row++

	// Number of previous backup to use
	lbl, err = SetupLabelJustifyRight(locale.T(MsgPrefDlgNumberOfPreviousBackupToUseCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	sbNumberOfPreviousBackupToUse, err := gtk.SpinButtonNewWithRange(1, 20, 1)
	if err != nil {
		return nil, err
	}
	sbNumberOfPreviousBackupToUse.SetTooltipText(locale.T(MsgPrefDlgNumberOfPreviousBackupToUseHint, nil))
	sbNumberOfPreviousBackupToUse.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_NUMBER_OF_PREVIOUS_BACKUP_TO_USE, sbNumberOfPreviousBackupToUse, "value", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(sbNumberOfPreviousBackupToUse, DesignSecondCol, row, 1, 1)
	row++

	sep, err = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	SetAllMargins(&sep.Widget, 6)
	grid.Attach(sep, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// ---------------------------------------------------------
	// Rsync file transfer options block
	// ---------------------------------------------------------
	markup = NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgAdvancedRsyncFileTransferOptionsSection, nil), "")
	lbl, err = SetupLabelMarkupJustifyLeft(markup)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Enable/disable RSYNC transfer source owner
	cbTransferSourceOwner, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourceOwner.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourceOwnerCaption, nil))
	cbTransferSourceOwner.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourceOwnerHint, nil))
	cbTransferSourceOwner.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_OWNER, cbTransferSourceOwner, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferSourceOwner, DesignFirstCol, row, 1, 1)

	// Enable/disable RSYNC transfer source group
	cbTransferSourceGroup, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourceGroup.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourceGroupCaption, nil))
	cbTransferSourceGroup.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourceGroupHint, nil))
	cbTransferSourceGroup.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_GROUP, cbTransferSourceGroup, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferSourceGroup, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC transfer source permissions
	cbTransferSourcePermissions, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourcePermissions.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsCaption, nil))
	cbTransferSourcePermissions.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsHint, nil))
	cbTransferSourcePermissions.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS, cbTransferSourcePermissions, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferSourcePermissions, DesignFirstCol, row, 1, 1)

	// Enable/disable RSYNC symlinks recreation
	cbRecreateSymlinks, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbRecreateSymlinks.SetLabel(locale.T(MsgPrefDlgRsyncRecreateSymlinksCaption, nil))
	cbRecreateSymlinks.SetTooltipText(locale.T(MsgPrefDlgRsyncRecreateSymlinksHint, nil))
	cbRecreateSymlinks.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_RECREATE_SYMLINKS, cbRecreateSymlinks, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbRecreateSymlinks, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC transfer device files
	cbTransferDeviceFiles, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferDeviceFiles.SetLabel(locale.T(MsgPrefDlgRsyncTransferDeviceFilesCaption, nil))
	cbTransferDeviceFiles.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferDeviceFilesHint, nil))
	cbTransferDeviceFiles.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_DEVICE_FILES, cbTransferDeviceFiles, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferDeviceFiles, DesignFirstCol, row, 1, 1)

	// Enable/disable RSYNC transfer special files
	cbTransferSpecialFiles, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSpecialFiles.SetLabel(locale.T(MsgPrefDlgRsyncTransferSpecialFilesCaption, nil))
	cbTransferSpecialFiles.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSpecialFilesHint, nil))
	cbTransferSpecialFiles.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SPECIAL_FILES, cbTransferSpecialFiles, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferSpecialFiles, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC compress file transfer
	cbCompressFileTransfer, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbCompressFileTransfer.SetLabel(locale.T(MsgPrefDlgRsyncCompressFileTransferCaption, nil))
	cbCompressFileTransfer.SetTooltipText(locale.T(MsgPrefDlgRsyncCompressFileTransferHint, nil))
	cbCompressFileTransfer.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_COMPRESS_FILE_TRANSFER, cbCompressFileTransfer, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbCompressFileTransfer, DesignFirstCol, row, 1, 1)
	row++

	box.Add(grid)

	_, err = box.Connect("destroy", func(b *gtk.Box) {
		bh.Unbind()
	})
	if err != nil {
		return nil, err
	}

	return &box.Container, nil
}

// ProfileStatusState is used to denote profile validating status.
type ProfileStatusState int

const (
	ProfileStatusNone ProfileStatusState = 1 << iota
	ProfileStatusValidating
	ProfileStatusError
)

type ProfileStatus struct {
	Status      ProfileStatusState
	Description string
}

type RestartService struct {
	Revealer *gtk.Revealer
}

func (v *RestartService) show(show bool) {
	if show {
		v.Revealer.SetRevealChild(true)
		v.Revealer.ShowAll()
	} else {
		v.Revealer.SetRevealChild(false)
	}
}

type RsyncSource struct {
	// Keep delete RSYNC source module button
	// to enable/disable it, depending on module count.
	DeleteBtn *gtk.Button
	IndexLbl  *gtk.Label
	Index     int
}

// PreferenceRow keeps extra data globally
// for each page over multi-page preference dialog.
// In some cases implement some kind of DOM support
// to enable/disable controls, show statuses, etc.
type PreferenceRow struct {
	sync.RWMutex
	ID             string
	name           string
	Title          string
	Row            *gtk.ListBoxRow
	Container      *gtk.Box
	Label          *gtk.Label
	Icon           *gtk.Image
	Page           *gtk.Container
	Profile        bool
	RestartService *RestartService
	RsyncSources   map[uintptr]*RsyncSource
	Errors         map[uintptr]ProfileStatus
}

// PreferenceRowNew instantiate new PreferenceRow object.
func PreferenceRowNew(id, title string, page *gtk.Container,
	profile, restartService bool) (*PreferenceRow, error) {

	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 6)
	box.SetSpacing(6)

	lbl, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	lbl.SetHAlign(gtk.ALIGN_START)
	box.PackStart(lbl, false, true, 0)

	row, err := gtk.ListBoxRowNew()
	if err != nil {
		return nil, err
	}
	row.Add(box)

	errors := make(map[uintptr]ProfileStatus)
	rsyncSources := make(map[uintptr]*RsyncSource)

	pr := &PreferenceRow{ID: id, Title: title, Row: row,
		Container: box, Label: lbl, Page: page, Profile: profile,
		Errors: errors, RsyncSources: rsyncSources}

	pr.SetName(title)

	return pr, nil
}

// SetName set profile name as a template "Profile(<name>)"
func (v *PreferenceRow) SetName(name string) {
	v.Lock()
	defer v.Unlock()

	v.name = name
	if v.Profile {
		publicName := locale.T(MsgPrefDlgProfileTabName,
			struct{ ProfileName string }{ProfileName: name})
		MustIdleAdd(func() {
			v.Label.SetText(publicName)
		})
	} else {
		MustIdleAdd(func() {
			v.Label.SetText(name)
		})
	}
}

// GetName get name.
func (v *PreferenceRow) GetName() string {
	v.RLock()
	defer v.RUnlock()

	return v.name
}

// AddStatus add error status to the list box item.
func (v *PreferenceRow) AddStatus(sourceID uintptr,
	status ProfileStatusState, description string) error {

	v.Lock()
	defer v.Unlock()

	lastStatus := v.getCurrentStatus()

	v.Errors[sourceID] = ProfileStatus{Status: status, Description: description}

	err2 := v.updateErrorStatus(lastStatus)
	if err2 != nil {
		return err2
	}
	return nil
}

// RemoveStatus removes error status from the list box item.
func (v *PreferenceRow) RemoveStatus(sourceID uintptr) error {
	v.Lock()
	defer v.Unlock()

	lastStatus := v.getCurrentStatus()

	delete(v.Errors, sourceID)

	err2 := v.updateErrorStatus(lastStatus)
	if err2 != nil {
		return err2
	}
	return nil
}

// ActivateRestartService show/hide restart command panel
// located at the top of the form.
func (v *PreferenceRow) ActivateRestartService(activate bool) error {
	if v.RestartService != nil {
		v.RestartService.show(activate)
	}
	return nil
}

// EnableDisableDeleteButtonsAndRecalculateIndexes enable/dsiable delete button
// for RSYNC module (doesn't allow to delete last module).
// Additionally recalculate module's indexes
func (v *PreferenceRow) EnableDisableDeleteButtonsAndRecalculateIndexes() {
	labels := []struct {
		DeleteBtn *gtk.Button
		IndexLbl  *gtk.Label
		Index     int
	}{}
	for _, rs := range v.RsyncSources {
		labels = append(labels, struct {
			DeleteBtn *gtk.Button
			IndexLbl  *gtk.Label
			Index     int
		}{DeleteBtn: rs.DeleteBtn, IndexLbl: rs.IndexLbl, Index: rs.Index})
	}
	sort.SliceStable(labels, func(i, j int) bool {
		return labels[i].Index < labels[j].Index
	})
	j := 0
	for _, rs := range labels {
		markup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
			"", "", NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
				"", "", NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
					"", "", NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
						"", "", NewMarkup(MARKUP_SIZE_LARGER, 0, 0,
							strconv.Itoa(j+1), "")))))
		rs.DeleteBtn.SetSensitive(len(v.RsyncSources) > 1)
		rs.IndexLbl.SetText(markup.String())
		rs.IndexLbl.SetUseMarkup(true)
		j++
	}
}

// GetLastRsyncModuleIndex extract maximum Index field value,
// that exists for RSYNC modules in this specific backup profile.
func (v *PreferenceRow) GetLastRsyncModuleIndex() int {
	j := -1
	for _, rs := range v.RsyncSources {
		if rs.Index > j {
			j = rs.Index
		}
	}
	return j
}

// setThemedIcon assign icon to the right side of the list box item.
func (v *PreferenceRow) setThemedIcon(themedName string, cssClasses []string) error {
	img, err := gtk.ImageNew()
	if err != nil {
		return err
	}
	err = AddStyleClasses(&img.Widget, cssClasses)
	if err != nil {
		return err
	}
	img.SetFromIconName(themedName, gtk.ICON_SIZE_BUTTON)
	v.assignImage(img)
	return nil
}

/*
// setAssetsIconAnimation assign icon to the right side of the list box item.
func (v *PreferenceRow) setAssetsIconAnimation(assetName string, resizeToWidth, resizeToHeight int) error {
	img, err := AnimationImageFromAssetsNewWithResize(assetName, resizeToWidth, resizeToHeight)
	if err != nil {
		return err
	}
	v.assignImage(img)
	return nil
}

// setAssetsIcon assign icon to the right side of the list box item.
func (v *PreferenceRow) setAssetsIcon(assetName string, cssClasses []string) error {
	img, err := ImageFromAssetsNewWithResize(assetName, 16, 16)
	if err != nil {
		return err
	}
	err = AddStyleClasses(&img.Widget, cssClasses)
	if err != nil {
		return err
	}
	v.assignImage(img)
	return nil
}
*/

func (v *PreferenceRow) assignImage(image *gtk.Image) {
	MustIdleAdd(func() {
		v.clearIcon()
		v.Icon = image
		v.Container.PackEnd(image, false, false, 0)
		v.Container.ShowAll()
	})
}

// clearIcon removes icon from the list box item.
func (v *PreferenceRow) clearIcon() {
	if v.Icon != nil {
		v.Icon.Destroy()
		v.Icon = nil
	}
}

// setTooltipMarkup assign tooltip to the list box item.
func (v *PreferenceRow) setTooltipMarkup(tooltip string) {
	MustIdleAdd(func() {
		v.Row.SetTooltipMarkup(tooltip)
	})
}

// getCurrentStatus return bitmask which describe existing
// validation statuses for current profile.
func (v *PreferenceRow) getCurrentStatus() ProfileStatusState {
	var errorFound, validatingFound bool
	for _, v := range v.Errors {
		if v.Status == ProfileStatusError {
			lg.Debugf("PreferenceRow error %q", v)
			errorFound = true
		} else if v.Status == ProfileStatusValidating {
			lg.Debugf("PreferenceRow validating %q", v)
			validatingFound = true
		}
	}
	var flags ProfileStatusState
	if errorFound {
		flags |= ProfileStatusError
	}
	if validatingFound {
		flags |= ProfileStatusValidating
	}
	return flags
}

// updateErrorStatus clear or set validation statuses to the list box item.
// It proceeds, if only changes from last update found.
func (v *PreferenceRow) updateErrorStatus(lastStatus ProfileStatusState) error {
	newStatus := v.getCurrentStatus()

	var validatingChanged, errorChanged bool

	if lastStatus&ProfileStatusValidating != newStatus&ProfileStatusValidating {
		validatingChanged = true
	}
	if lastStatus&ProfileStatusError != newStatus&ProfileStatusError {
		errorChanged = true
	}

	if validatingChanged || errorChanged {
		if newStatus&ProfileStatusValidating != 0 {
			lg.Debug("Validating found")
			markup := NewMarkup(0, MARKUP_COLOR_SKY_BLUE, 0,
				locale.T(MsgPrefDlgSourceRsyncValidatingHint, nil), nil)
			v.setTooltipMarkup(markup.String())
			err := v.setThemedIcon(STOCK_SYNCHRONIZING_ICON, []string{"image-spin"})
			if err != nil {
				lg.Fatal(err)
			}
		} else if newStatus&ProfileStatusError != 0 {
			lg.Debug("Error found")
			markup := NewMarkup(0, MARKUP_COLOR_ORANGE_RED, 0,
				locale.T(MsgPrefDlgProfileConfigIssuesDetectedWarning, nil), nil)
			v.setTooltipMarkup(markup.String())
			err := v.setThemedIcon(STOCK_IMPORTANT_ICON, []string{"image-error", "image-shake"})
			if err != nil {
				lg.Fatal(err)
			}
		} else {
			lg.Debug("No errors found")
			v.setTooltipMarkup("")
			MustIdleAdd(func() {
				v.clearIcon()
			})
		}
	}
	return nil
}

// PreferenceRowList keep a link between GtkListBoxRow
// and specific PreferenceRow object.
type PreferenceRowList struct {
	m      map[uintptr]*PreferenceRow
	sorted []uintptr
}

func PreferenceRowListNew() *PreferenceRowList {
	var m = make(map[uintptr]*PreferenceRow)
	v := &PreferenceRowList{m: m}
	return v
}

func (v *PreferenceRowList) Append(row *PreferenceRow) {
	v.m[row.Row.Native()] = row
	v.sorted = append(v.sorted, row.Row.Native())
}

func (v *PreferenceRowList) Delete(rowID uintptr) {
	delete(v.m, rowID)
	for ind, val := range v.sorted {
		if val == rowID {
			v.sorted = append(v.sorted[:ind], v.sorted[ind+1:]...)
			break
		}
	}
}

func (v *PreferenceRowList) Get(rowID uintptr) *PreferenceRow {
	return v.m[rowID]
}

func (v *PreferenceRowList) GetLastProfileListIndex() int {
	lastIndex := -1
	for _, rowID := range v.sorted {
		if v.m[rowID].Profile && v.m[rowID].Row.GetIndex() > lastIndex {
			lastIndex = v.m[rowID].Row.GetIndex()
		}
	}
	return lastIndex
}

func (v *PreferenceRowList) GetProfileCount() int {
	count := 0
	for _, rowID := range v.sorted {
		if v.m[rowID].Profile {
			count++
		}
	}
	return count
}

func (v *PreferenceRowList) GetProfiles() []*PreferenceRow {
	var rows []*PreferenceRow
	for _, rowID := range v.sorted {
		if v.m[rowID].Profile {
			rows = append(rows, v.m[rowID])
		}
	}
	return rows
}

// addProfilePage build UI on the top of profile taken from GlibSettings.
func addProfilePage(win *gtk.ApplicationWindow, profileID string, initProfileName *string,
	appSettings *SettingsStore, list *PreferenceRowList, validator *UIValidator,
	lbSide *gtk.ListBox, pages *gtk.Stack, selectNew bool, profileChanged func()) error {

	prefRow, err := PreferenceRowNew(profileID,
		locale.T(MsgPrefDlgGeneralProfileTabName, nil), nil, true, false)
	if err != nil {
		return err
	}
	page, profileName, err := ProfilePreferencesNew(win, appSettings, validator,
		profileID, prefRow, profileChanged, initProfileName)
	if err != nil {
		return err
	}
	prefRow.SetName(profileName)
	prefRow.Page = page
	pages.AddTitled(page, profileID, "Profile")
	list.Append(prefRow)
	index := list.GetLastProfileListIndex()
	lbSide.Insert(prefRow.Row, index+1)
	lbSide.ShowAll()
	pages.ShowAll()
	if selectNew {
		lbSide.SelectRow(prefRow.Row)
	}
	return nil
}

// CreatePreferenceDialog creates multi-page preference dialog
// with save/restore functionality to/from the GLib Setting object.
func CreatePreferenceDialog(settingsID, settingsPath string, mainWin *gtk.ApplicationWindow,
	profileChanged func()) (*gtk.ApplicationWindow, error) {

	app, err := mainWin.GetApplication()
	if err != nil {
		return nil, err
	}
	win, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		return nil, err
	}

	// General window settings
	win.SetTransientFor(mainWin)
	win.SetDestroyWithParent(false)
	win.SetShowMenubar(false)
	appSettings, err := NewSettingsStore(settingsID, settingsPath, nil)
	if err != nil {
		return nil, err
	}

	// Create window header
	hbMain, err := SetupHeader("", "", true)
	if err != nil {
		return nil, err
	}
	hbMain.SetHExpand(true)

	hbSide, err := SetupHeader(locale.T(MsgPrefDlgPreferencesDialogCaption, nil), "", false)
	if err != nil {
		return nil, err
	}
	hbSide.SetHExpand(false)

	bTitle, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	bTitle.Add(hbSide)
	sTitle, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	bTitle.Add(sTitle)
	bTitle.Add(hbMain)

	win.SetTitlebar(bTitle)

	var list = PreferenceRowListNew()
	// TODO: better to create and keep this variable in global context
	// to skip possible race issues, in case of multiple preference
	// windows opened simultaneously.
	var validator = UIValidatorNew(context.Background())

	_, err = win.Connect("destroy", func() {
		validator.CancelAll()
	})
	if err != nil {
		return nil, err
	}

	// Create Stack and boxes
	pages, err := gtk.StackNew()
	if err != nil {
		return nil, err
	}
	pages.SetHExpand(true)
	pages.SetVExpand(true)

	// Create ListBox
	lbSide, err := gtk.ListBoxNew()
	if err != nil {
		return nil, err
	}
	lbSide.SetCanFocus(true)
	lbSide.SetSelectionMode(gtk.SELECTION_BROWSE)
	lbSide.SetVExpand(true)

	var pr *PreferenceRow

	profileSettingsArray := appSettings.NewSettingsArray(CFG_BACKUP_LIST)
	profileList := profileSettingsArray.GetArrayIDs()
	if len(profileList) == 0 {
		profileID, err := profileSettingsArray.AddNode()
		if err != nil {
			return nil, err
		}
		profileSettings, err := getProfileSettings(appSettings, profileID, nil)
		if err != nil {
			return nil, err
		}
		sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
		_, err = sarr.AddNode()
		if err != nil {
			return nil, err
		}
		profileName := profileID
		if i, err := strconv.Atoi(profileID); err == nil {
			profileName = strconv.Itoa(i + 1)
		}
		err = addProfilePage(win, profileID, &profileName, appSettings, list,
			validator, lbSide, pages, false, profileChanged)
		if err != nil {
			return nil, err
		}
	} else {
		for _, profileID := range profileList {
			err = addProfilePage(win, profileID, nil, appSettings, list,
				validator, lbSide, pages, false, profileChanged)
			if err != nil {
				return nil, err
			}
		}
	}

	pr, err = PreferenceRowNew("General_ID", locale.T(MsgPrefDlgGeneralTabName, nil), nil, false, true)
	if err != nil {
		return nil, err
	}
	gp, err := GeneralPreferencesNew(win, appSettings, &mainWin.ActionMap, pr)
	if err != nil {
		return nil, err
	}
	pages.AddTitled(gp, "General_ID", locale.T(MsgPrefDlgGeneralTabName, nil))
	list.Append(pr)
	lbSide.Add(pr.Row)

	pr, err = PreferenceRowNew("Advanced_ID", locale.T(MsgPrefDlgAdvancedTabName, nil), nil, false, true)
	if err != nil {
		return nil, err
	}
	ap, err := AdvancedPreferencesNew(appSettings, pr)
	if err != nil {
		return nil, err
	}
	pages.AddTitled(ap, "Advanced_ID", locale.T(MsgPrefDlgAdvancedTabName, nil))
	list.Append(pr)
	lbSide.Add(pr.Row)

	sw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, err
	}
	sw.Add(lbSide)
	sw.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	sw.SetShadowType(gtk.SHADOW_NONE)
	sw.SetSizeRequest(220, -1)

	vp, err := gtk.ViewportNew(nil, nil)
	if err != nil {
		return nil, err
	}
	vp.Add(sw)

	bSide, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}
	bSide.Add(vp)

	bButtons, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	SetAllMargins(bButtons, 6)
	bSide.Add(bButtons)
	btnAddProfile, err := SetupButtonWithThemedImage("list-add-symbolic")
	if err != nil {
		return nil, err
	}
	btnAddProfile.SetTooltipText(locale.T(MsgPrefDlgAddProfileHint, nil))
	_, err = btnAddProfile.Connect("clicked", func() {
		profileID, err := profileSettingsArray.AddNode()
		if err != nil {
			lg.Fatal(err)
		}
		profileSettings, err := getProfileSettings(appSettings, profileID, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}
		sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
		_, err = sarr.AddNode()
		if err != nil {
			lg.Fatal(err)
		}

		profileName := profileID
		if i, err := strconv.Atoi(profileID); err == nil {
			profileName = strconv.Itoa(i + 1)
		}
		err = addProfilePage(win, profileID, &profileName, appSettings, list,
			validator, lbSide, pages, true, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}
		if profileChanged != nil {
			profileChanged()
		}
	})
	if err != nil {
		return nil, err
	}
	bButtons.PackStart(btnAddProfile, false, false, 0)

	// Function to manage (enable/disable) "delete backup profile" button.
	updateBtnDeleteProfileSensitive := func(deleteBtn *gtk.Button, row *gtk.ListBoxRow) {
		var pr *PreferenceRow
		if row != nil {
			pr = list.Get(row.Native())
			pages.SetVisibleChildName(pr.ID)
			hbMain.SetTitle(pr.Title)
		}
		deleteBtn.SetSensitive(pr != nil && pr.Profile && list.GetProfileCount() > 1)
	}

	btnDeleteProfile, err := SetupButtonWithThemedImage("list-remove-symbolic")
	if err != nil {
		return nil, err
	}
	btnDeleteProfile.SetTooltipText(locale.T(MsgPrefDlgDeleteProfileHint, nil))
	_, err = btnDeleteProfile.Connect("clicked", func() {
		title := locale.T(MsgPrefDlgDeleteProfileDialogTitle, nil)
		titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
			NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
		yesButtonCaption := locale.T(MsgDialogYesButton, nil)
		yesButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, removeUndescore(yesButtonCaption), nil)
		textMarkup := locale.T(MsgPrefDlgDeleteProfileDialogText,
			struct{ YesButton string }{YesButton: yesButtonMarkup.String()})
		responseYes, err := questionDialog(&win.Window, titleMarkup.String(), textMarkup, true, true, false)
		if err != nil {
			lg.Fatal(err)
		}

		if responseYes {
			sr := lbSide.GetSelectedRow()
			sri := sr.GetIndex()
			pr := list.Get(sr.Native())
			if pr.Profile {
				profileID := pr.ID
				profileSettings, err := getProfileSettings(appSettings, profileID, profileChanged)
				if err != nil {
					lg.Fatal(err)
				}
				sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
				ids := sarr.GetArrayIDs()
				for _, sourceID := range ids {
					sourceSettings, err := getBackupSourceSettings(profileSettings, sourceID, profileChanged)
					if err != nil {
						lg.Fatal(err)
					}
					err = sarr.DeleteNode(sourceSettings, sourceID)
					if err != nil {
						lg.Fatal(err)
					}
				}

				err = profileSettingsArray.DeleteNode(profileSettings, profileID)
				if err != nil {
					lg.Fatal(err)
				}
				nsr := lbSide.GetRowAtIndex(sri + 1)
				lbSide.SelectRow(nsr)
				pages.Remove(pr.Page)
				list.Delete(sr.Native())
				pr.Page.Destroy()
				sr.Destroy()
				updateBtnDeleteProfileSensitive(btnDeleteProfile, lbSide.GetSelectedRow())

				if profileChanged != nil {
					profileChanged()
				}
			}
		}
	})
	if err != nil {
		return nil, err
	}
	bButtons.PackStart(btnDeleteProfile, false, false, 0)

	_, err = lbSide.Connect("row-selected", func(lb *gtk.ListBox, row *gtk.ListBoxRow) {
		lg.Debugf("Row at index %d selected", row.GetIndex())
		updateBtnDeleteProfileSensitive(btnDeleteProfile, row)
	})
	if err != nil {
		return nil, err
	}

	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	box.Add(bSide)
	div, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	box.Add(div)
	box.Add(pages)

	win.Add(box)

	sgSide, err := gtk.SizeGroupNew(gtk.SIZE_GROUP_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	sgSide.AddWidget(hbSide)
	sgSide.AddWidget(bSide)

	sgMain, err := gtk.SizeGroupNew(gtk.SIZE_GROUP_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	sgMain.AddWidget(hbMain)
	sgMain.AddWidget(pages)

	return win, nil
}
