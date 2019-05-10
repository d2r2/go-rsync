package gtkui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
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
	ASSET_SYNCHRONIZING_ICON = "emblem-synchronizing-cyan.gif"
	STOCK_IMPORTANT_ICON     = "emblem-important-symbolic"
	ASSET_IMPORTANT_ICON     = "emblem-important-red.gif"
	STOCK_NETWORK_ERROR_ICON = "network-error-symbolic"
)

// return error describing issue with conversion from one type to another.
func validatorConversionError(fromType, toType string) error {
	msg := spew.Sprintf("Can't convert %[1]v to %[2]v", fromType, toType)
	err := errors.New(msg)
	return err
}

// setupLabelJustifyRight create GtkLabel with justification to the right by default.
func setupLabelJustifyRight(caption string) (*gtk.Label, error) {
	lbl, err := gtk.LabelNew(caption)
	if err != nil {
		return nil, err
	}
	lbl.SetHAlign(gtk.ALIGN_END)
	lbl.SetJustify(gtk.JUSTIFY_RIGHT)
	return lbl, nil
}

const (
	DesignIndentCol     = 0
	DesignFirstCol      = 4
	DesignSecondCol     = 5
	DesignTotalColCount = 6
)

// Create preference dialog with "General" page, where controls
// being bound to GLib Setting object to save/restore functionality.
func GeneralPreferencesNew(gsSettings *glib.Settings) (*gtk.Container, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 18)

	bh := BindingHelperNew(gsSettings)

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	row := 0

	// ---------------------------------------------------------
	// Interface options block
	// ---------------------------------------------------------
	markup := NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgGeneralUserInterfaceOptionsSecion, nil), "")
	lbl, err := gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Option to show about dialog on application startup
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgDoNotShowAtAppStartupCaption, nil))
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgPerformDesktopNotificationCaption, nil))
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgLanguageCaption, nil))
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
	row++

	// Session log font size
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgSessionLogControlFontSizeCaption, nil))
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
	lbl, err = gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Ignore file signature
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgSkipFolderBackupFileSignatureCaption, nil))
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

	/*
		// ---------------------------------------------------------
		// Debug section
		// ---------------------------------------------------------
		lbl, err = gtk.LabelNew(NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
			"DEBUG SECTION", "").String())
		if err != nil {
			return nil, err
		}
		lbl.SetUseMarkup(true)
		lbl.SetHAlign(gtk.ALIGN_START)
		if err != nil {
			return nil, err
		}
		grid.Attach(lbl, 0, row, 2, 1)
		row++

		pb, err := gtk.ProgressBarNew()
		if err != nil {
			return nil, err
		}
		pb.SetPulseStep(0.1)
		err = FixProgressBarCSSStyle(pb)
		if err != nil {
			return nil, err
		}
		grid.Attach(pb, 0, row, 2, 1)
		row++

		btn, err := gtk.ButtonNew()
		if err != nil {
			return nil, err
		}
		lbl, err = gtk.LabelNew("1")
		if err != nil {
			return nil, err
		}
		btn.Add(lbl)
		btn.Connect("clicked", func(btn *gtk.Button) {
			pb.Pulse()
		})
		grid.Attach(btn, 0, row, 1, 1)
		row++

		btn, err = gtk.ButtonNew()
		if err != nil {
			return nil, err
		}
		lbl, err = gtk.LabelNew("2")
		if err != nil {
			return nil, err
		}
		btn.Add(lbl)
		btn.Connect("clicked", func(btn *gtk.Button) {
			pb.SetFraction(0.3)
		})
		grid.Attach(btn, 0, row, 1, 1)
		row++
	*/

	box.Add(grid)

	_, err = box.Connect("destroy", func(b *gtk.Box) {
		bh.Unbind()
	})
	if err != nil {
		return nil, err
	}

	return &box.Container, nil
}

// GetSubpathRegexp verify that proposed file system path expression is valid.
// Understand path separator for different OS, taking path separator setting from runtime.
//
// Use Microsoft Windows restriction list taken from here:
// https://stackoverflow.com/questions/1976007/what-characters-are-forbidden-in-windows-and-linux-directory-names
//
// Linux/Unix:
// / (forward slash)
//
// Windows:
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
func GetSubpathRegexp() (*regexp.Regexp, error) {
	template := spew.Sprintf(`^\%[1]c?([^\<\>\:\"\|\?\*\%[1]c]+\%[1]c?)*$`, os.PathSeparator)
	lg.Debugf("Subpath regex template: %s", template)
	rexp, err := regexp.Compile(template)
	if err != nil {
		return nil, err
	}
	return rexp, nil
}

// RestartTimer restart timer with call fire after specific millisecond period.
// Used as a trigger for validation events.
func RestartTimer(timer *time.Timer, milliseconds time.Duration) {
	timer.Stop()
	timer.Reset(time.Millisecond * milliseconds)
}

func createBackupSourceBlock(profileID, sourceID string, gsSettings *glib.Settings,
	prefRow *PreferenceRow, validator *UIValidator, profileChanged *bool) (gtk.IWidget, error) {
	//	frame, err := gtk.FrameNew("")
	//	if err != nil {
	//		return nil, err
	//	}
	//	frame.SetShadowType(gtk.SHADOW_ETCHED_OUT)

	rsyncPathGroupName := spew.Sprintf("RsyncPath_%v_%v", profileID, sourceID)
	destSubPathGroupName := spew.Sprintf("DestSubpath_%v", profileID)

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 18)

	bh := BindingHelperNew(gsSettings)

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_FILL)
	row := 0

	// Source rsync path
	lbl, err := setupLabelJustifyRight(locale.T(MsgPrefDlgSourceRsyncPathCaption, nil))
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	grid.Attach(lbl, 0, row, 1, 1)
	edRsyncPath, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	edRsyncPath.SetHExpand(true)
	//edRsyncPath.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_QUESTION_ICON)
	edRsyncPath.SetIconTooltipText(gtk.ENTRY_ICON_SECONDARY, locale.T(MsgPrefDlgSourceRsyncPathRetryHint, nil))

	grid.Attach(edRsyncPath, 1, row, 1, 1)
	row++

	// Destination root path
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgDestinationSubpathCaption, nil))
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	grid.Attach(lbl, 0, row, 1, 1)
	edDestSubpath, err := gtk.EntryNew()
	if err != nil {
		return nil, err
	}
	//edDestSubpath.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_OK_ICON)
	edDestSubpath.SetTooltipText(locale.T(MsgPrefDlgDestinationSubpathHint, nil))
	grid.Attach(edDestSubpath, 1, row, 1, 1)
	row++

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
	SetAllMargins(box2, 9)
	box2.SetMarginEnd(0)
	box2.SetMarginStart(0)
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
	SetAllMargins(box3, 18)
	frame.Add(box3)

	grid2, err := gtk.GridNew()
	grid2.SetColumnSpacing(12)
	grid2.SetRowSpacing(6)
	grid2.SetHAlign(gtk.ALIGN_FILL)
	box3.PackStart(grid2, true, true, 0)
	row2 := 0
	/*
		// Authenticate user
		lbl, err = setupLabelJustifyRight("Auth. user")
		if err != nil {
			return nil, err
		}
		lbl.SetUseMarkup(true)
		grid2.Attach(lbl, 0, row2, 1, 1)
		edAuthUser, err := gtk.EntryNew()
		if err != nil {
			return nil, err
		}
		edAuthUser.SetTooltipText("")
		edAuthUser.SetHExpand(true)
		grid2.Attach(edAuthUser, 1, row2, 1, 1)
		row2++
	*/
	// Authenticate password
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgAuthPasswordCaption, nil))
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgChangeFilePermissionCaption, nil))
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgEnableBackupBlockCaption, nil))
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	grid.Attach(lbl, 0, row, 1, 1)
	swEnabled, err := gtk.SwitchNew()
	if err != nil {
		return nil, err
	}
	swEnabled.SetTooltipText(locale.T(MsgPrefDlgEnableBackupBlockHint, nil))
	swEnabled.SetHAlign(gtk.ALIGN_START)
	grid.Attach(swEnabled, 1, row, 1, 1)
	row++

	rsyncPathGroupIndex := validator.AddEntry(rsyncPathGroupName,
		// 1st stage.
		// Initialize data validation.
		// Synchronized call: can update GTK widgets from here.
		func(data *ValidatorData, group []*ValidatorData) error {
			entry, ok := data.Items[0].(*gtk.Entry)
			if !ok {
				return validatorConversionError("ValidatorData.Items[0]", "*gtk.Entry")
			}
			// entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_SYNCHRONIZING_ICON)
			err := SetEntryIconWithAssetImage(entry, gtk.ENTRY_ICON_SECONDARY, ASSET_SYNCHRONIZING_ICON)
			if err != nil {
				return err
			}
			RsyncSourcePathDescription := locale.T(MsgPrefDlgSourceRsyncPathDescriptionHint, nil)
			markup := markupTooltip(NewMarkup(0, MARKUP_COLOR_SKY_BLUE, 0,
				locale.T(MsgPrefDlgSourceRsyncValidatingHint, nil), nil), RsyncSourcePathDescription)
			entry.SetTooltipMarkup(markup.String())
			return nil
		},
		// 2nd stage.
		// Execute validation.
		// Asynchronous call: doesn't allowed to change GTK widgets from here (only read)!
		func(ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {
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
					msg := locale.T(MsgPrefDlgSourceRsyncPathEmptyError, nil)
					warning = &msg
				} else {
					lg.Debugf("Start rsync utility to validate rsync source")
					sourceSettings, err := getBackupSourceSettings(profileID, sourceID, profileChanged)
					var authPass *string
					ap := sourceSettings.GetString(CFG_MODULE_AUTH_PASSWORD)
					if ap != "" {
						authPass = &ap
					}
					err = rsync.GetPathStatus(ctx, authPass, rsyncURL, false)
					if err != nil {
						lg.Debug(err)
						if !rsync.IsRsyncProcessTerminatedError(err) {
							msg := err.Error()
							warning = &msg
						}
					}
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage.
		// Finalize data validation.
		// Synchronized call: can update GTK widgets from here.
		func(data *ValidatorData, results []interface{}) error {
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
			RsyncSourcePathDescription := locale.T(MsgPrefDlgSourceRsyncPathDescriptionHint, nil)
			if swtch.GetActive() {
				warning, ok := results[0].(*string)
				if !ok {
					return validatorConversionError("interface{}[0]", "*string")
				}
				if warning != nil {
					err := SetEntryIconWithAssetImage(entry, gtk.ENTRY_ICON_SECONDARY, ASSET_IMPORTANT_ICON)
					if err != nil {
						return err
					}
					markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
						RsyncSourcePathDescription)
					entry.SetTooltipMarkup(markup.String())
					//entry.SetTooltipText(spew.Sprintf("Error: %s", *warning))
					err = row.AddErrorStatus(entry.Native(), *warning)
					if err != nil {
						return err
					}
				} else {
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_OK_ICON)
					entry.SetTooltipText(RsyncSourcePathDescription)
					//fgColor := "Cyan"
					//fntWeight := "bold"
					//markup := markupTooltip("Verified", &fgColor, &fntWeight, RsyncSourcePathDescription)
					//entry.SetTooltipMarkup(markup)
					entry.SetTooltipText(RsyncSourcePathDescription)
					//					entry.SetIconTooltipText(gtk.ENTRY_ICON_SECONDARY, "Press to validate rsync source")
					err := row.AddErrorStatus(entry.Native(), "")
					if err != nil {
						return err
					}
				}
			} else {
				entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
				markup := markupTooltip(NewMarkup(0, 0 /*MARKUP_COLOR_LIGHT_GRAY*/, 0,
					locale.T(MsgPrefDlgSourceRsyncPathNotValidatedHint, nil), nil), RsyncSourcePathDescription)
				entry.SetTooltipMarkup(markup.String())
				err := row.AddErrorStatus(entry.Native(), "")
				if err != nil {
					return err
				}
			}
			return nil
		}, edRsyncPath, swEnabled, prefRow)

	rsyncPathTimer := time.AfterFunc(time.Millisecond*1000, func() {
		_, err := glib.IdleAdd(func() {
			err := validator.Validate(rsyncPathGroupName)
			if err != nil {
				lg.Fatal(err)
			}
		})
		if err != nil {
			lg.Fatal(err)
		}
	})
	rsyncPathTimer.Stop()
	_, err = edRsyncPath.Connect("changed", func(v *gtk.Entry) {
		RestartTimer(rsyncPathTimer, 1000)
	})
	if err != nil {
		return nil, err
	}
	_, err = edRsyncPath.Connect("icon-press", func(v *gtk.Entry) {
		RestartTimer(rsyncPathTimer, 50)
	})
	if err != nil {
		return nil, err
	}
	_, err = edRsyncPath.Connect("destroy", func(entry *gtk.Entry) {
		lg.Debug("Destroy edRsyncPath")
		err := prefRow.RemoveErrorStatus(entry.Native())
		if err != nil {
			lg.Fatal(err)
		}
		validator.RemoveEntry(rsyncPathGroupIndex)
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

	rexp, err := GetSubpathRegexp()
	if err != nil {
		return nil, err
	}
	destSubPathGroupIndex := validator.AddEntry(destSubPathGroupName,
		// 1st stage.
		// Initialize data validation.
		// Synchronized call: can update GTK widgets from here.
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
				// entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_SYNCHRONIZING_ICON)
				SetEntryIconWithAssetImage(entry, gtk.ENTRY_ICON_SECONDARY, ASSET_SYNCHRONIZING_ICON)
				if err != nil {
					return err
				}
			}
			return nil
		},
		// 2nd stage.
		// Execute validation.
		// Asynchronous call: doesn't allowed to change GTK widgets from here (only read)!
		func(ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {
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
			if swtch.GetActive() && !rexp.MatchString(destSubPath) {
				msg := locale.T(MsgPrefDlgDestinationSubpathExpressionError, nil)
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
					msg := locale.T(MsgPrefDlgDestinationSubpathNotUniqueError, nil)
					warning = &msg
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage.
		// Finalize data validation.
		// Synchronized call: can update GTK widgets from here.
		func(data *ValidatorData, results []interface{}) error {
			var DestSubpathHint = locale.T(MsgPrefDlgDestinationSubpathHint, nil)

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
				warning, ok := results[0].(*string)
				if !ok {
					return validatorConversionError("interface{}[0]", "*string")
				}
				if warning != nil {
					err := SetEntryIconWithAssetImage(entry, gtk.ENTRY_ICON_SECONDARY, ASSET_IMPORTANT_ICON)
					if err != nil {
						return err
					}
					markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
						DestSubpathHint)
					entry.SetTooltipMarkup(markup.String())
					//entry.SetTooltipText(*warning)
					err = row.AddErrorStatus(entry.Native(), *warning)
					if err != nil {
						return err
					}
				} else {
					entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, STOCK_OK_ICON)
					//fgcolor := "Royal Blue"
					//fgColor := "Cyan"
					//fntWeight := "bold"
					//markup := markupTooltip("Verified", &fgColor, &fntWeight, DestSubpathHint)
					//entry.SetTooltipMarkup(markup)
					entry.SetTooltipText(DestSubpathHint)
					err = row.AddErrorStatus(entry.Native(), "")
					if err != nil {
						return err
					}
				}
			} else {
				entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
				markup := markupTooltip(NewMarkup(0, 0 /*MARKUP_COLOR_LIGHT_GRAY*/, 0,
					locale.T(MsgPrefDlgDestinationSubpathNotValidatedHint, nil), nil), DestSubpathHint)
				entry.SetTooltipMarkup(markup.String())
				err := row.AddErrorStatus(entry.Native(), "")
				if err != nil {
					lg.Fatal(err)
				}
			}
			return nil
		}, edDestSubpath, swEnabled, prefRow)
	destSubpathTimer := time.AfterFunc(time.Millisecond*500, func() {
		_, err := glib.IdleAdd(func() {
			err := validator.Validate(destSubPathGroupName)
			if err != nil {
				lg.Fatal(err)
			}
		})
		if err != nil {
			lg.Fatal(err)
		}
	})
	destSubpathTimer.Stop()
	_, err = edDestSubpath.Connect("changed", func(v *gtk.Entry) {
		RestartTimer(destSubpathTimer, 500)
	})
	if err != nil {
		return nil, err
	}
	_, err = edDestSubpath.Connect("destroy", func(entry *gtk.Entry) {
		lg.Debug("Destroy edDestSubpath")
		err := prefRow.RemoveErrorStatus(entry.Native())
		if err != nil {
			lg.Fatal(err)
		}
		validator.RemoveEntry(destSubPathGroupIndex)
		RestartTimer(destSubpathTimer, 50)
	})
	if err != nil {
		return nil, err
	}

	_, err = edAuthPasswd.Connect("changed", func(v *gtk.Entry) {
		if swEnabled.GetActive() {
			RestartTimer(rsyncPathTimer, 1000)
		}
	})
	if err != nil {
		return nil, err
	}

	bh.Bind(CFG_MODULE_DEST_SUBPATH, edDestSubpath, "text", glib.SETTINGS_BIND_DEFAULT)

	bh.Bind(CFG_MODULE_CHANGE_FILE_PERMISSION, edChmod, "text", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_MODULE_AUTH_PASSWORD, edAuthPasswd, "text", glib.SETTINGS_BIND_DEFAULT)

	sourceSettings, err := getBackupSourceSettings(profileID, sourceID, profileChanged)
	if err != nil {
		return nil, err
	}
	ap := sourceSettings.GetString(CFG_MODULE_AUTH_PASSWORD)
	cfp := sourceSettings.GetString(CFG_MODULE_CHANGE_FILE_PERMISSION)
	expExtraOptions.SetExpanded(ap != "" || cfp != "")

	_, err = swEnabled.Connect("state-set", func(v *gtk.Switch) {
		RestartTimer(rsyncPathTimer, 50)
		RestartTimer(destSubpathTimer, 50)
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

	RestartTimer(rsyncPathTimer, 50)
	RestartTimer(destSubpathTimer, 50)

	return box, nil
}

// getBackupSettings create GlibSettings object with change event
// connected to specific indexed profile[profileID].
func getBackupSettings(profileID string, profileChanged *bool) (*glib.Settings, error) {
	path := fmt.Sprintf(core.SETTINGS_PROFILE_PATH, profileID)
	gs, err := glib.SettingsNewWithPath(core.SETTINGS_PROFILE_ID, path)
	if err != nil {
		return nil, err
	}
	_, err = gs.Connect("changed", func() {
		if profileChanged != nil {
			*profileChanged = true
		}
	})
	if err != nil {
		return nil, err
	}
	return gs, nil
}

// getBackupSourceSettings create GlibSettings object with change event
// connected to specific indexed source[profile[profileID], sourceID].
func getBackupSourceSettings(profileID, sourceID string, profileChanged *bool) (*glib.Settings, error) {
	path := fmt.Sprintf(core.SETTINGS_SOURCE_PATH, profileID, sourceID)
	gs, err := glib.SettingsNewWithPath(core.SETTINGS_SOURCE_ID, path)
	if err != nil {
		return nil, err
	}
	_, err = gs.Connect("changed", func() {
		if profileChanged != nil {
			*profileChanged = true
		}
	})
	if err != nil {
		return nil, err
	}
	return gs, nil
}

func createBackupSourceBlock2(profileID, sourceID string, prefRow *PreferenceRow,
	validator *UIValidator, profileChanged *bool) (*gtk.Container, error) {

	backupSettings, err := getBackupSettings(profileID, profileChanged)
	if err != nil {
		return nil, err
	}
	sourceSettings, err := getBackupSourceSettings(profileID, sourceID, profileChanged)
	if err != nil {
		lg.Fatal(err)
	}

	box2, err := createBackupSourceBlock(profileID, sourceID, sourceSettings, prefRow, validator, profileChanged)
	if err != nil {
		return nil, err
	}

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	box3, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		return nil, err
	}
	box3.Add(box2)
	box.Add(box3)
	btnDeleteSource, err := gtk.ButtonNew()
	if err != nil {
		return nil, err
	}
	lbl, err := gtk.LabelNew(locale.T(MsgPrefDlgDeleteBackupBlockCaption, nil))
	if err != nil {
		return nil, err
	}
	btnDeleteSource.Add(lbl)
	btnDeleteSource.SetVAlign(gtk.ALIGN_CENTER)
	btnDeleteSource.SetTooltipText(locale.T(MsgPrefDlgDeleteBackupBlockHint, nil))
	_, err = btnDeleteSource.Connect("clicked", func(btn *gtk.Button, box *gtk.Box) {
		box.Destroy()

		sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
		err = sarr.DeleteNode(sourceSettings, sourceID)
		if err != nil {
			lg.Fatal(err)
		}
	}, box)
	if err != nil {
		return nil, err
	}
	//	if sourceID == "0" {
	//		btnDeleteSource.SetSensitive(false)
	//	}
	box3.PackStart(btnDeleteSource, false, false, 0)
	box3.SetHExpand(true)
	box3.SetVExpand(false)
	sep, err := gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	box.Add(sep)

	return &box.Container, nil
}

// Create preference dialog with "Sources" page, where controls
// being bound to GLib Setting object to save/restore functionality.
func BackupPreferencesNew(appSettings *glib.Settings, list *PreferenceRowList,
	validator *UIValidator, profileID string, prefRow *PreferenceRow,
	profileChanged *bool, initProfileName *string) (*gtk.Container, string, error) {

	sw, err := gtk.ScrolledWindowNew(nil, nil)
	if err != nil {
		return nil, "", err
	}
	sw.SetPolicy(gtk.POLICY_NEVER, gtk.POLICY_AUTOMATIC)
	//SetScrolledWindowPropogatedHeight(sw, true)

	//box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	frame, err := gtk.FrameNew(locale.T(MsgPrefDlgSourcesCaption, nil))
	if err != nil {
		return nil, "", err
	}
	//frame.SetLabelAlign(0.01, 0.5)
	box0, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, "", err
	}
	SetAllMargins(box0, 18)
	frame.Add(box0)

	box1, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, "", err
	}
	box0.Add(box1)

	backupSettings, err := getBackupSettings(profileID, profileChanged)
	if err != nil {
		return nil, "", err
	}

	sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
	for _, srcID := range sarr.GetArrayIDs() {
		box2, err := createBackupSourceBlock2(profileID, srcID, prefRow, validator, profileChanged)
		if err != nil {
			return nil, "", err
		}
		box1.Add(box2)
	}

	btnAddSource, err := SetupButtonWithThemedImage("list-add-symbolic")
	if err != nil {
		return nil, "", err
	}
	btnAddSource.SetTooltipText(locale.T(MsgPrefDlgAddBackupBlockHint, nil))
	_, err = btnAddSource.Connect("clicked", func() {
		sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
		sourceID, err := sarr.AddNode()
		if err != nil {
			lg.Fatal(err)
		}

		box3, err := createBackupSourceBlock2(profileID, sourceID, prefRow, validator, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}

		box1.Add(box3)
		box1.ShowAll()

		destSubPathGroupName := spew.Sprintf("DestSubpath_%v", profileID)
		err = validator.Validate(destSubPathGroupName)
		if err != nil {
			lg.Fatal(err)
		}
	})
	if err != nil {
		return nil, "", err
	}

	box0.Add(btnAddSource)

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_FILL)
	row := 0

	var lbl *gtk.Label

	appBH := BindingHelperNew(appSettings)
	backupBH := BindingHelperNew(backupSettings)

	// Profile name
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgProfileNameCaption, nil))
	if err != nil {
		return nil, "", err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	edProfileName, err := gtk.EntryNew()
	if err != nil {
		return nil, "", err
	}
	//	edProfileName.SetTooltipText("Profile name")
	edProfileName.SetHExpand(true)
	edProfileName.SetHAlign(gtk.ALIGN_FILL)
	profileGroupIndex := validator.AddEntry("ProfileName",
		// 1st stage.
		// Initialize data validation.
		// Synchronized call: can update GTK widgets here.
		func(data *ValidatorData, group []*ValidatorData) error {
			return nil
		},
		// 2nd stage.
		// Execute validation.
		// Asynchronous call: doesn't allowed to change GTK widgets here (only read)!
		func(ctx context.Context, data *ValidatorData, group []*ValidatorData) ([]interface{}, error) {
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
				msg := locale.T(MsgPrefDlgProfileNameEmptyWarning, nil)
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
					msg := locale.T(MsgPrefDlgProfileNameExistsWarning,
						struct{ ProfileName string }{ProfileName: profileName})
					warning = &msg
				}
			}
			return []interface{}{warning}, nil
		},
		// 3rd stage.
		// Finalize data validation.
		// Synchronized call: can update GTK widgets here.
		func(data *ValidatorData, results []interface{}) error {
			var ProfileNameHint = locale.T(MsgPrefDlgProfileNameHint, nil)
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
			if warning != nil {
				err := SetEntryIconWithAssetImage(entry, gtk.ENTRY_ICON_SECONDARY, ASSET_IMPORTANT_ICON)
				if err != nil {
					return err
				}
				markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, *warning, nil),
					ProfileNameHint)
				entry.SetTooltipMarkup(markup.String())
				err = row.AddErrorStatus(entry.Native(), *warning)
				if err != nil {
					return err
				}
			} else {
				entry.SetIconFromIconName(gtk.ENTRY_ICON_SECONDARY, "")
				entry.SetTooltipText(ProfileNameHint)
				err = row.RemoveErrorStatus(entry.Native())
				if err != nil {
					return err
				}
			}
			return nil
		}, edProfileName, prefRow)
	backupBH.Bind(CFG_PROFILE_NAME, edProfileName, "text", glib.SETTINGS_BIND_DEFAULT)
	timer := time.AfterFunc(time.Millisecond*500, func() {
		_, err := glib.IdleAdd(func() {
			name, err := edProfileName.GetText()
			if err != nil {
				lg.Fatal(err)
			}
			prefRow.SetName(name)
			err = validator.Validate("ProfileName")
			if err != nil {
				lg.Fatal(err)
			}
		})
		if err != nil {
			lg.Fatal(err)
		}
	})
	_, err = edProfileName.Connect("changed", func(v *gtk.Entry, tmr *time.Timer) {
		tmr.Stop()
		tmr.Reset(time.Millisecond * 500)
	}, timer)
	if err != nil {
		return nil, "", err
	}
	_, err = edProfileName.Connect("destroy", func(entry *gtk.Entry) {
		validator.RemoveEntry(profileGroupIndex)
		err = validator.Validate("ProfileName")
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgDefaultDestPathCaption, nil))
	if err != nil {
		return nil, "", err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	destFolder, err := gtk.FileChooserButtonNew("Select default destination folder", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	if err != nil {
		return nil, "", err
	}
	destFolder.SetTooltipText(locale.T(MsgPrefDlgDefaultDestPathHint, nil))
	destFolder.SetHExpand(true)
	destFolder.SetHAlign(gtk.ALIGN_FILL)
	folder := backupSettings.GetString(CFG_PROFILE_DEST_ROOT_PATH)
	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		// lg.Println(spew.Sprintf("File %q found", filename))
		destFolder.SetFilename(folder)
	}
	_, err = destFolder.Connect("file-set", func(fcb *gtk.FileChooserButton) {
		folder := fcb.GetFilename()
		if _, err := os.Stat(folder); !os.IsNotExist(err) {
			backupSettings.SetString(CFG_PROFILE_DEST_ROOT_PATH, folder)
		}
	})
	if err != nil {
		return nil, "", err
	}
	grid.Attach(destFolder, 1, row, 1, 1)
	row++

	box2, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, "", err
	}
	SetAllMargins(box2, 18)
	box2.Add(grid)
	box2.Add(frame)

	vp, err := gtk.ViewportNew(nil, nil)
	if err != nil {
		return nil, "", err
	}
	vp.Add(box2)

	sw.Add(vp)
	_, err = sw.Connect("destroy", func(b gtk.IWidget) {
		appBH.Unbind()
		backupBH.Unbind()
	})
	if err != nil {
		return nil, "", err
	}

	/*
		act, err := createAddNewBackupSourceAction(profileID, box, btnAddSource)
		if err != nil {
			return nil, nil, nil, err
		}
		actionMap.AddAction(act)
	*/

	name := backupSettings.GetString(CFG_PROFILE_NAME)
	return &sw.Container, name, nil
}

// AdvancedPreferencesNew create preference dialog with "Advanced" page, where controls
// bound to GLib Setting object for save/restore functionality.
func AdvancedPreferencesNew(gsSettings *glib.Settings) (*gtk.Container, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box, 18)

	bh := BindingHelperNew(gsSettings)

	grid, err := gtk.GridNew()
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	row := 0

	// ---------------------------------------------------------
	// Backup settings block
	// ---------------------------------------------------------
	markup := NewMarkup(MARKUP_WEIGHT_BOLD, 0, 0,
		locale.T(MsgPrefDlgAdvancedBackupSettingsSection, nil), "")
	lbl, err := gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Enable/disable automatic backup block size
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgAutoManageBackupBlockSizeCaption, nil))
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgBackupBlockSizeCaption, nil))
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgRunNotificationScriptCaption, nil))
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
	lbl, err = gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Rsync utility retry count
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgRsyncRetryCountCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignFirstCol, row, 1, 1)
	sbRetryCount, err := gtk.SpinButtonNewWithRange(0, 5, 1)
	if err != nil {
		return nil, err
	}
	sbRetryCount.SetTooltipText(locale.T(MsgPrefDlgRsyncRetryCountHint, nil))
	//sbRetryCount.SetHExpand(false)
	sbRetryCount.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_RETRY_COUNT, sbRetryCount, "value", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(sbRetryCount, DesignSecondCol, row, 1, 1)
	row++

	// Enable/disable RSYNC low level log
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgRsyncLowLevelLogCaption, nil))
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
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgRsyncIntensiveLowLevelLogCaption, nil))
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
	bh.Bind(CFG_ENABLE_INTENSIVE_LOW_LEVEL_LOG_OF_RSYNC, cbIntensiveLowLevelRsyncLog, "active", glib.SETTINGS_BIND_DEFAULT)
	bh.Bind(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC, cbIntensiveLowLevelRsyncLog, "sensitive", glib.SETTINGS_BIND_GET)
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
	lbl, err = gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
	row++

	// Use previous backup if found
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgUsePreviousBackupForDedupCaption, nil))
	if err != nil {
		return nil, err
	}
	eb, err = gtk.EventBoxNew()
	if err != nil {
		return nil, err
	}
	//eb.AddEvents(int(gdk.BUTTON_PRESS_MASK))
	//eb.AddEvents(int(gdk.BUTTON_RELEASE_MASK))
	//eb.AddEvents(int(gdk.EVENT_BUTTON_PRESS))
	//eb.AddEvents(int(gdk.EVENT_BUTTON_RELEASE))
	//eb.AddEvents(int(gdk.EVENT_DOUBLE_BUTTON_PRESS))
	//	eb.AddEvents(int(gdk.EVENT_DOUBLE_BUTTON_PRESS))
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
	//eb.Connect("button-release-event", func() {
	//	cbPrevBackupUsage.SetActive(!cbPrevBackupUsage.GetActive())
	//})
	//eb.Connect("toggled", func() {
	//	cbPrevBackupUsage.SetActive(!cbPrevBackupUsage.GetActive())
	//})
	cbPrevBackupUsage.SetTooltipText(locale.T(MsgPrefDlgUsePreviousBackupForDedupHint, nil))
	cbPrevBackupUsage.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_ENABLE_USE_OF_PREVIOUS_BACKUP, cbPrevBackupUsage, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbPrevBackupUsage, DesignSecondCol, row, 1, 1)
	row++

	// Number of previous backup to use
	lbl, err = setupLabelJustifyRight(locale.T(MsgPrefDlgNumberOfPreviousBackupToUseCaption, nil))
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
	lbl, err = gtk.LabelNew(markup.String())
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, DesignIndentCol, row, DesignTotalColCount, 1)
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

	// Enable/disable RSYNC transfer source permissions
	cbTransferSourcePermissions, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbTransferSourcePermissions.SetLabel(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsCaption, nil))
	cbTransferSourcePermissions.SetTooltipText(locale.T(MsgPrefDlgRsyncTransferSourcePermissionsHint, nil))
	cbTransferSourcePermissions.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS, cbTransferSourcePermissions, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbTransferSourcePermissions, DesignSecondCol, row, 1, 1)
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

	// Enable/disable RSYNC symlinks recreation
	cbRecreateSymlinks, err := gtk.CheckButtonNew()
	if err != nil {
		return nil, err
	}
	cbRecreateSymlinks.SetLabel(locale.T(MsgPrefDlgRsyncRecreateSymlinksCaption, nil))
	cbRecreateSymlinks.SetTooltipText(locale.T(MsgPrefDlgRsyncRecreateSymlinksHint, nil))
	cbRecreateSymlinks.SetHAlign(gtk.ALIGN_START)
	bh.Bind(CFG_RSYNC_RECREATE_SYMLINKS, cbRecreateSymlinks, "active", glib.SETTINGS_BIND_DEFAULT)
	grid.Attach(cbRecreateSymlinks, DesignFirstCol, row, 1, 1)
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

	box.Add(grid)

	_, err = box.Connect("destroy", func(b *gtk.Box) {
		bh.Unbind()
	})
	if err != nil {
		return nil, err
	}

	return &box.Container, nil

	//bh := BindingHelperNew(gsSettings)

	_, err = box.Connect("destroy", func() {
		//bh.Unbind()
	})
	if err != nil {
		return nil, err
	}

	return &box.Container, nil
}

// PreferenceRow keeps here extra data for each page of multi-page preference dialog.
type PreferenceRow struct {
	sync.RWMutex
	ID        string
	name      string
	Title     string
	Row       *gtk.ListBoxRow
	Container *gtk.Box
	Label     *gtk.Label
	Icon      *gtk.Image
	Page      *gtk.Container
	Profile   bool
	Errors    map[uintptr]string
}

// PreferenceRowNew instantiate new PreferenceRow object.
func PreferenceRowNew(id, title string, page *gtk.Container,
	profile bool) (*PreferenceRow, error) {

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

	errors := make(map[uintptr]string)

	pr := &PreferenceRow{ID: id, Title: title, Row: row,
		Container: box, Label: lbl, Page: page,
		Profile: profile, Errors: errors}

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
		v.Label.SetText(publicName)
	} else {
		v.Label.SetText(name)
	}
}

// GetName get name.
func (v *PreferenceRow) GetName() string {
	v.RLock()
	defer v.RUnlock()

	return v.name
}

// setThemedIcon assign icon to the right side of the list box item.
func (v *PreferenceRow) setThemedIcon(themedName string) error {
	img, err := gtk.ImageNew()
	if err != nil {
		return err
	}
	img.SetFromIconName(themedName, gtk.ICON_SIZE_BUTTON)
	v.clearIcon()
	v.Icon = img
	v.Container.PackEnd(img, false, false, 0)
	v.Container.ShowAll()
	return nil
}

// setAssetsIcon assign icon to the right side of the list box item.
func (v *PreferenceRow) setAssetsIcon(assetName string) error {
	img, err := ImageFromAssetsNew(assetName, 16, 16)
	if err != nil {
		return err
	}
	v.clearIcon()
	v.Icon = img
	v.Container.PackEnd(img, false, false, 0)
	v.Container.ShowAll()
	return nil
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
	v.Row.SetTooltipMarkup(tooltip)
}

// updateErrorStatus clear or set error status to the list box item.
func (v *PreferenceRow) updateErrorStatus() error {
	found := false
	for _, v := range v.Errors {
		if v != "" {
			lg.Debugf("PreferenceRow error %q", v)
			found = true
			break
		}
	}
	//	glib.IdleAdd(func() {
	if found {
		lg.Debug("Error found")
		markup := NewMarkup(0, MARKUP_COLOR_ORANGE_RED, 0,
			locale.T(MsgPrefDlgProfileConfigIssuesDetectedWarning, nil), nil)
		v.setTooltipMarkup(markup.String())
		err := v.setAssetsIcon(ASSET_IMPORTANT_ICON)
		if err != nil {
			lg.Fatal(err)
		}
	} else {
		lg.Debug("No errors found")
		v.setTooltipMarkup("")
		v.clearIcon()
	}
	//	})
	return nil
}

// AddErrorStatus add error status to the list box item.
func (v *PreferenceRow) AddErrorStatus(sourceID uintptr, err string) error {
	v.Lock()
	defer v.Unlock()

	v.Errors[sourceID] = err

	err2 := v.updateErrorStatus()
	if err2 != nil {
		return err2
	}
	return nil
}

// RemoveErrorStatus removes error status from the list box item.
func (v *PreferenceRow) RemoveErrorStatus(sourceID uintptr) error {
	v.Lock()
	defer v.Unlock()

	delete(v.Errors, sourceID)

	err2 := v.updateErrorStatus()
	if err2 != nil {
		return err2
	}
	return nil
}

// PreferenceRowList keeps a link between GtkListBoxRow
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

/*
func findProfilesByNameTillCurrent(list []*PreferenceRow,
	current *PreferenceRow, profileName string) []*PreferenceRow {

	rows := []*PreferenceRow{}
	for _, pr := range list {
		if pr == current {
			break
		}
		if pr.GetName() == profileName {
			rows = append(rows, pr)
		}
	}
	return rows
}
*/

// addProfilePage build UI on the top of profile taken from GlibSettings.
func addProfilePage(profileID string, initProfileName *string, appSettings *glib.Settings,
	list *PreferenceRowList, validator *UIValidator, lbSide *gtk.ListBox, pages *gtk.Stack, selectNew bool,
	profileChanged *bool) error {

	prefRow, err := PreferenceRowNew(profileID,
		locale.T(MsgPrefDlgGeneralProfileTabName, nil), nil, true)
	if err != nil {
		return err
	}
	page, profileName, err := BackupPreferencesNew(appSettings, list, validator,
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

// Create multi-page preference dialog
// with save/restore functionality to/from the GLib Setting object.
func CreatePreferenceDialog(settingsID string, app *gtk.Application, profileChanged *bool) (*gtk.ApplicationWindow, error) {
	parentWin := app.GetActiveWindow()
	win, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		return nil, err
	}

	// Settings
	win.SetTransientFor(parentWin)
	win.SetDestroyWithParent(false)
	win.SetShowMenubar(false)
	appSettings, err := glib.SettingsNew(settingsID)
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
	bTitle.Add(hbSide)
	sTitle, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	bTitle.Add(sTitle)
	bTitle.Add(hbMain)

	win.SetTitlebar(bTitle)

	var list = PreferenceRowListNew()
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

	profileSettingsArray := NewSettingsArray(appSettings, CFG_BACKUP_LIST)
	profileList := profileSettingsArray.GetArrayIDs()
	if len(profileList) == 0 {
		profileID, err := profileSettingsArray.AddNode()
		if err != nil {
			return nil, err
		}
		backupSettings, err := getBackupSettings(profileID, nil)
		if err != nil {
			return nil, err
		}
		sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
		_, err = sarr.AddNode()
		if err != nil {
			return nil, err
		}
		profileName := profileID
		if i, err := strconv.Atoi(profileID); err == nil {
			profileName = strconv.Itoa(i + 1)
		}
		err = addProfilePage(profileID, &profileName, appSettings, list,
			validator, lbSide, pages, false, profileChanged)
		if err != nil {
			return nil, err
		}
	} else {
		for _, profileID := range profileList {
			err = addProfilePage(profileID, nil, appSettings, list,
				validator, lbSide, pages, false, profileChanged)
			if err != nil {
				return nil, err
			}
		}
	}

	gp, err := GeneralPreferencesNew(appSettings)
	if err != nil {
		return nil, err
	}
	pages.AddTitled(gp, "General_ID", locale.T(MsgPrefDlgGeneralTabName, nil))
	pr, err = PreferenceRowNew("General_ID", locale.T(MsgPrefDlgGeneralTabName, nil), gp, false)
	if err != nil {
		return nil, err
	}
	list.Append(pr)
	lbSide.Add(pr.Row)

	ap, err := AdvancedPreferencesNew(appSettings)
	if err != nil {
		return nil, err
	}
	pages.AddTitled(ap, "Advanced_ID", locale.T(MsgPrefDlgAdvancedTabName, nil))
	pr, err = PreferenceRowNew("Advanced_ID", locale.T(MsgPrefDlgAdvancedTabName, nil), ap, false)
	if err != nil {
		return nil, err
	}
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
		backupSettings, err := getBackupSettings(profileID, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}
		sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
		_, err = sarr.AddNode()
		if err != nil {
			lg.Fatal(err)
		}

		profileName := profileID
		if i, err := strconv.Atoi(profileID); err == nil {
			profileName = strconv.Itoa(i + 1)
		}
		err = addProfilePage(profileID, &profileName, appSettings, list,
			validator, lbSide, pages, true, profileChanged)
		if err != nil {
			lg.Fatal(err)
		}
		if profileChanged != nil {
			*profileChanged = true
		}
	})
	if err != nil {
		return nil, err
	}
	bButtons.PackStart(btnAddProfile, false, false, 0)
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
		yesButtonMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, yesButtonCaption, nil)
		textMarkup := locale.T(MsgPrefDlgDeleteProfileDialogText, struct{ YesButton string }{YesButton: yesButtonMarkup.String()})
		responseYes, err := questionDialog(&win.Window, titleMarkup.String(), textMarkup, true, true, false)
		// responseYes, err := QuestionDialog(&win.Window, title,
		// 	[]*DialogParagraph{NewDialogParagraph(text)}, false)
		if err != nil {
			lg.Fatal(err)
		}

		if responseYes {
			sr := lbSide.GetSelectedRow()
			sri := sr.GetIndex()
			pr := list.Get(sr.Native())
			if pr.Profile {
				profileID := pr.ID
				backupSettings, err := getBackupSettings(profileID, profileChanged)
				if err != nil {
					lg.Fatal(err)
				}
				sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
				ids := sarr.GetArrayIDs()
				for _, sourceID := range ids {
					sourceSettings, err := getBackupSourceSettings(profileID, sourceID, profileChanged)
					if err != nil {
						lg.Fatal(err)
					}
					err = sarr.DeleteNode(sourceSettings, sourceID)
					if err != nil {
						lg.Fatal(err)
					}
				}

				err = profileSettingsArray.DeleteNode(backupSettings, profileID)
				if err != nil {
					lg.Fatal(err)
				}
				nsr := lbSide.GetRowAtIndex(sri + 1)
				lbSide.SelectRow(nsr)
				pages.Remove(pr.Page)
				list.Delete(sr.Native())
				pr.Page.Destroy()
				sr.Destroy()
				if profileChanged != nil {
					*profileChanged = true
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
		pr := list.Get(row.Native())
		//lg.Println(spew.Sprintf("%+v", r1))
		pages.SetVisibleChildName(pr.ID)
		hbMain.SetTitle(pr.Title)
		btnDeleteProfile.SetSensitive(pr.Profile && list.GetProfileCount() > 1)
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

	// Set initial title
	//hbMain.SetTitle("Global")

	//win.SetDefaultSize(1000, -1)
	//	win.SetDefaultSize(500, -1)

	return win, nil
}
