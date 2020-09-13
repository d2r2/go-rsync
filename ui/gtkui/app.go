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

package gtkui

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/backup"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	shell "github.com/d2r2/go-shell"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
)

// interruptBackupProcess open dialog to confirm backup process interruption.
func interruptBackupProcess(win *gtk.Window, backupSync *BackupSessionStatus) (bool, error) {
	quit := true
	if backupSync.IsRunning() {
		done := make(chan struct{})
		dialog, err := createInterruptBackupDialog(win)
		if err != nil {
			return false, err
		}

		go func() {
			terminate := false
			for {
				select {
				case <-done:
					terminate = true
				// Check every 0,5 sec that backup process still alive.
				case <-time.After(time.Millisecond * 500):
					if !backupSync.IsRunning() {
						MustIdleAdd(func() {
							dialog.dialog.Response(gtk.RESPONSE_NO)
						})
					}
				}
				if terminate {
					break
				}
			}
		}()

		response := dialog.Run(false)
		quit = IsResponseYes(response)
		close(done)
	}
	return quit, nil
}

// createQuitAction creates exit app action.
func createQuitAction(win *gtk.Window, backupSync *BackupSessionStatus, supplimentary *RunningContexts) (glib.IAction, error) {
	act, err := glib.SimpleActionNew("QuitAction", nil)
	if err != nil {
		return nil, err
	}

	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		quit, err := interruptBackupProcess(win, backupSync)
		if err != nil {
			lg.Fatal(err)
		}

		if quit {
			app, err := win.GetApplication()
			if err != nil {
				lg.Fatal(err)
			}
			if backupSync.IsRunning() {
				backupSync.Stop()
			}
			// terminate all supplementary services if running
			supplimentary.CancelAll()
			// loop through and close all windows currently opened
			for {
				win2 := app.GetActiveWindow()
				if win2 != nil {
					win2.Close()
					for gtk.EventsPending() {
						gtk.MainIterationDo(true)
					}
				} else {
					break
				}
			}
			// quit application
			app.Quit()
		}
	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// createAboutAction creates "about dialog" action.
func createAboutAction(win *gtk.Window, appSettings *SettingsStore) (glib.IAction, error) {
	act, err := glib.SimpleActionNew("AboutAction", nil)
	if err != nil {
		return nil, err
	}

	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		dlg, err := CreateAboutDialog(appSettings)
		if err != nil {
			lg.Fatal(err)
		}

		dlg.SetTransientFor(win)
		dlg.SetModal(true)
		dlg.ShowNow()
	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// createHelpAction pop up default browser with URI link to project github README.md file.
func createHelpAction(win *gtk.Window) (glib.IAction, error) {
	act, err := glib.SimpleActionNew("HelpAction", nil)
	if err != nil {
		return nil, err
	}

	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		// ignore error
		_ = ShowUri(win, "https://gorsync.github.io")
	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// createMenuModelForPopover construct menu for popover button.
func createMenuModelForPopover() (glib.IMenuModel, error) {
	main, err := glib.MenuNew()
	if err != nil {
		return nil, err
	}

	var section *glib.Menu

	// New menu section (with buttons)
	section, err = glib.MenuNew()
	if err != nil {
		return nil, err
	}
	section.Append(locale.T(MsgAppWindowAboutMenuCaption, nil), "win.AboutAction")
	section.Append(locale.T(MsgAppWindowHelpMenuCaption, nil), "win.HelpAction")
	main.AppendSection("", section)

	section, err = glib.MenuNew()
	if err != nil {
		return nil, err
	}
	section.Append(locale.T(MsgAppWindowPreferencesMenuCaption, nil), "win.PreferenceAction")
	main.AppendSection("", section)

	section, err = glib.MenuNew()
	if err != nil {
		return nil, err
	}
	section.Append(locale.T(MsgAppWindowQuitMenuCaption, nil), "win.QuitAction")
	main.AppendSection("", section)

	//	main.AppendItem(section.MenuModel)

	return main, nil
}

// createPreferenceAction constructs multi-page preference dialog
// with save/restore functionality to/from the GLib GSettings object.
// Action activation require to have GLib Setting Schema
// preliminary installed, otherwise will not work raising error.
// Installation bash script from app folder must be performed in advance.
func createPreferenceAction(mainWin *gtk.ApplicationWindow, profile *gtk.ComboBox) (glib.IAction, error) {
	act, err := glib.SimpleActionNew("PreferenceAction", nil)
	if err != nil {
		return nil, err
	}

	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		app, err := mainWin.GetApplication()
		if err != nil {
			lg.Fatal(err)
		}

		extraMsg := locale.T(MsgSchemaConfigDlgSchemaErrorAdvise,
			struct{ ScriptName string }{ScriptName: "gs_schema_install.sh"})
		found, err := CheckSchemaSettingsIsInstalled(SETTINGS_SCHEMA_ID, app, &extraMsg)
		if err != nil {
			lg.Fatal(err)
		}

		if found {

			profileChanged := make(chan struct{})
			var once sync.Once
			changedFunc := func() {
				once.Do(func() {
					close(profileChanged)
				})
			}

			win, err := CreatePreferenceDialog(SETTINGS_SCHEMA_ID, SETTINGS_SCHEMA_PATH, mainWin, changedFunc)
			if err != nil {
				lg.Fatal(err)
			}

			win.ShowAll()
			win.Show()

			_, err = win.Connect("destroy", func(window *gtk.ApplicationWindow) {
				win.Close()
				win.Destroy()
				lg.Debug("Destroy window")

				changed := false
				select {
				case <-profileChanged:
					changed = true
				default:
				}

				if changed {
					lst, err := getProfileList()
					if err != nil {
						lg.Fatal(err)
					}
					err = UpdateNameValueCombo(profile, lst)
					if err != nil {
						lg.Fatal(err)
					}
					profile.SetActiveID("")
				}
			})
			if err != nil {
				lg.Fatal(err)
			}
		}

	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// enableAction finds GAction by name and enable/disable it.
func enableAction(win *gtk.ApplicationWindow, actionName string, enable bool) error {
	act := win.LookupAction(actionName)
	if act == nil {
		err := errors.New(locale.T(MsgActionDoesNotFound,
			struct{ ActionName string }{ActionName: actionName}))
		return err
	}
	action, err := glib.SimpleActionFromAction(act)
	if err != nil {
		return err
	}
	action.SetEnabled(enable)
	return nil
}

// EmptySpaceRecover used to try to recover from RSYNC critical error, caused
// by out of space state. Main entry ErrorHook is trying heuristically
// identify out of space symptoms and then check free space size.
type EmptySpaceRecover struct {
	main      *gtk.ApplicationWindow
	backupLog logger.PackageLog
}

// ErrorHook is a main entry hook to recover from RSYNC out of space issue.
func (v *EmptySpaceRecover) ErrorHook(err error, paths core.SrcDstPath, predictedSize *core.FolderSize,
	repeated int, retryLeft int) (newRetryLeft int, criticalError error) {

	if rsync.IsCallFailedError(err) {
		erro := err.(*rsync.CallFailedError)
		freeSpace, err2 := shell.GetFreeSpace(paths.DestPath)
		if err2 != nil {
			return retryLeft, err2
		}

		var predSize uint64
		if predictedSize != nil {
			predSize = predictedSize.GetByteCount()
		}

		lg.Debugf("Exit code = %d, error = %v, predicted size = %d MB, retry left = %d, space left: %d kB",
			erro.ExitCode, erro, predSize/core.MB, retryLeft, freeSpace/core.KB)

		if (erro.ExitCode == 23 || erro.ExitCode == 11) &&
			(predictedSize == nil && freeSpace < 1*core.MB || predictedSize.GetByteCount() > freeSpace) {

			v.backupLog.Notifyf(locale.T(MsgLogBackupStageOutOfSpaceWarning,
				struct{ SizeLeft string }{SizeLeft: core.FormatSize(freeSpace, true)}))

			response, err2 := outOfSpaceDialogAsync(&v.main.Window, paths, freeSpace)
			if err2 != nil {
				lg.Fatal(err2)
			}

			if response == OutOfSpaceRetry {
				// retry
				if retryLeft == 0 {
					retryLeft++
				}
				return retryLeft, nil
			} else if response == OutOfSpaceTerminate {
				// terminated backup process
				return 0, err
			} else {
				// ignore this call and continue
				return 0, nil
			}
		}
	}
	return retryLeft, nil
}

// traceLongRunningContext monitor system signals to cancel context finally if signal raised.
func traceLongRunningContext(ctx *ContextPack) chan struct{} {
	// Build actual signals list to control
	signals := []os.Signal{os.Kill}
	if shell.IsLinuxMacOSFreeBSD() {
		signals = append(signals, syscall.SIGTERM, os.Interrupt)
	}
	done := make(chan struct{})
	shell.CloseContextOnSignals(ctx.Cancel, done, signals...)
	return done
}

// performFullBackup run backup process, which include 1st and 2nd passes.
func performFullBackup(backupSync *BackupSessionStatus, notifier *NotifierUI,
	win *gtk.ApplicationWindow, config *backup.Config, modules []backup.Module, destPath string) {

	ctx := backupSync.Start()
	done := traceLongRunningContext(ctx)
	defer close(done)
	defer backupSync.Done(ctx.Context)

	backupLog := core.NewProxyLog(backup.LocalLog, "backup", 6, "15:04:05",
		func(line string) error {
			err := notifier.UpdateTextViewLog(line)
			if err != nil {
				return err
			}
			return nil
		}, logger.InfoLevel,
	)

	// Run 1st stage to prepare backup plan.
	plan, progress, err := backup.BuildBackupPlan(ctx.Context, backupLog, config, modules, notifier)
	if err == nil {
		lg.Debugf("Backup node's dir trees: %+v", plan)

		// Create empty space recover hook.
		emptySpaceRecover := &EmptySpaceRecover{main: win, backupLog: backupLog}
		// Run 2nd stage to perform backup itself.
		err = plan.RunBackup(progress, destPath, emptySpaceRecover.ErrorHook)

		notifier.ReportCompletion(1, err, progress, true)
		progress.Close()
	} else {
		notifier.ReportCompletion(0, err, nil, true)
	}
}

// setControlStateOnBackupStarted enable/disable actions according to backup
// process status. Actions in its turns associated with GTK widgets.
func setControlStateOnBackupStarted(win *gtk.ApplicationWindow,
	selectFolder *gtk.FileChooserButton, profile *gtk.ComboBox) {

	err := enableAction(win, "RunBackupAction", false)
	if err != nil {
		lg.Fatal(err)
	}
	err = enableAction(win, "PreferenceAction", false)
	if err != nil {
		lg.Fatal(err)
	}
	err = enableAction(win, "StopBackupAction", true)
	if err != nil {
		lg.Fatal(err)
	}
	profile.SetSensitive(false)
	selectFolder.SetSensitive(false)
}

// setControlStateOnBackupEnded enable/disable actions according to backup
// process status. Actions in its turns associated with GTK widgets.
func setControlStateOnBackupEnded(win *gtk.ApplicationWindow, selectFolder *gtk.FileChooserButton,
	profile *gtk.ComboBox, notifier *NotifierUI) {

	call := func() {
		profile.SetSensitive(true)
		selectFolder.SetSensitive(true)
		err := enableAction(win, "StopBackupAction", false)
		if err != nil {
			lg.Fatal(err)
		}
		err = enableAction(win, "PreferenceAction", true)
		if err != nil {
			lg.Fatal(err)
		}
		err = enableAction(win, "RunBackupAction", true)
		if err != nil {
			lg.Fatal(err)
		}
	}

	<-notifier.Done()
	MustIdleAdd(call)
}

// createRunBackupAction creates action - entry point for data backup process start.
func createRunBackupAction(win *gtk.ApplicationWindow, gridUI *gtk.Grid,
	destPath *string, selectFolder *gtk.FileChooserButton, profile *gtk.ComboBox,
	backupSync *BackupSessionStatus) (glib.IAction, error) {

	act, err := glib.SimpleActionNew("RunBackupAction", nil)
	if err != nil {
		return nil, err
	}

	act.SetEnabled(false)
	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		profileID := profile.GetActiveID()
		lg.Debugf("BackupID = %v", profileID)

		if profileID != "" {
			config, modules, err := readBackupConfig(profileID)
			if err != nil {
				lg.Fatal(err)
			}
			// verify that RSYNC modules configuration is valid, otherwise show error dialog
			if errFound, msg := isModulesConfigError(modules, true); errFound {
				title := locale.T(MsgAppWindowCannotStartBackupProcessTitle, nil)
				titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
					NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
				err = ErrorMessage(&win.Window, titleMarkup.String(), []*DialogParagraph{NewDialogParagraph(msg)})
				if err != nil {
					lg.Fatal(err)
				}
			} else if errFound, msg := isDestPathError(*destPath, true); errFound {
				title := locale.T(MsgAppWindowCannotStartBackupProcessTitle, nil)
				titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
					NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
				var text string
				if *destPath == "" {
					text = locale.T(MsgAppWindowDestPathIsEmptyError2, nil)
				} else {
					text = msg
				}
				err = ErrorMessage(&win.Window, titleMarkup.String(), []*DialogParagraph{NewDialogParagraph(text)})
				if err != nil {
					lg.Fatal(err)
				}
			} else {
				// enable/disable corresponding UI elements
				setControlStateOnBackupStarted(win, selectFolder, profile)

				appSettings, err := glib.SettingsNew(SETTINGS_SCHEMA_ID)
				if err != nil {
					lg.Fatal(err)
				}
				val, err := GetComboValue(profile, 0)
				if err != nil {
					lg.Fatal(err)
				}
				profileName, err := val.GetString()
				if err != nil {
					lg.Fatal(err)
				}
				notifier := NewNotifierUI(profileName, gridUI)
				err = notifier.ClearProgressGrid()
				if err != nil {
					lg.Fatal(err)
				}
				fontSize := appSettings.GetString(CFG_SESSION_LOG_WIDGET_FONT_SIZE)
				err = notifier.CreateProgressControls(fontSize)
				if err != nil {
					lg.Fatal(err)
				}
				err = notifier.UpdateBackupProgress(nil, locale.T(MsgAppWindowBackupProgressStartMessage, nil), false)
				if err != nil {
					lg.Fatal(err)
				}

				go func() {
					// perform a full backup cycle in one closure
					performFullBackup(backupSync, notifier, win, config, modules, *destPath)
					// enable/disable corresponding UI elements
					setControlStateOnBackupEnded(win, selectFolder, profile, notifier)
				}()
			}
		}
	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// createStopBackupAction creates entry point for action which terminate live backup session.
func createStopBackupAction(win *gtk.ApplicationWindow, grid *gtk.Grid,
	selectFolder *gtk.FileChooserButton, profile *gtk.ComboBox,
	backupSync *BackupSessionStatus) (glib.IAction, error) {

	act, err := glib.SimpleActionNew("StopBackupAction", nil)
	if err != nil {
		return nil, err
	}

	act.SetEnabled(false)
	_, err = act.Connect("activate", func(action *glib.SimpleAction, param *glib.Variant) {
		name, state, err := GetActionNameAndState(action)
		if err != nil {
			lg.Fatal(err)
		}
		lg.Debugf("%v action activated with current state %v and args %v",
			name, state, param)

		err = enableAction(win, "StopBackupAction", false)
		if err != nil {
			lg.Fatal(err)
		}

		quit, err := interruptBackupProcess(&win.Window, backupSync)
		if err != nil {
			lg.Fatal(err)
		}

		if quit {
			if backupSync.IsRunning() {
				backupSync.Stop()

				profile.SetSensitive(true)
				selectFolder.SetSensitive(true)
				err = enableAction(win, "PreferenceAction", true)
				if err != nil {
					lg.Fatal(err)
				}
				err = enableAction(win, "RunBackupAction", true)
				if err != nil {
					lg.Fatal(err)
				}
			}
		} else {
			if backupSync.IsRunning() {
				err = enableAction(win, "StopBackupAction", true)
				if err != nil {
					lg.Fatal(err)
				}
			}
		}

	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// getProfileList reads from app configuration profile's identifiers and names
// to use as a source for GtkComboBox widget.
func getProfileList() ([]struct{ value, key string }, error) {
	appSettings, err := NewSettingsStore(SETTINGS_SCHEMA_ID, SETTINGS_SCHEMA_PATH, nil)
	if err != nil {
		return nil, err
	}
	sarr := appSettings.NewSettingsArray(CFG_BACKUP_LIST)
	lst := sarr.GetArrayIDs()
	arr := []struct{ value, key string }{{locale.T(MsgAppWindowNoneProfileEntry, nil), ""}}
	for _, item := range lst {
		profileSettings, err := getProfileSettings(appSettings, item, nil)
		if err != nil {
			return nil, err
		}
		name := profileSettings.settings.GetString(CFG_PROFILE_NAME)
		arr = append(arr, struct{ value, key string }{name, item})
	}
	return arr, nil
}

// readBackupConfig reads from app glib.Settings configuration to Config object
// which contains all settings necessary to run new backup session.
func readBackupConfig(profileID string) (*backup.Config, []backup.Module, error) {
	appSettings, err := NewSettingsStore(SETTINGS_SCHEMA_ID, SETTINGS_SCHEMA_PATH, nil)
	if err != nil {
		return nil, nil, err
	}

	cfg := &backup.Config{}

	cfg.SigFileIgnoreBackup = appSettings.settings.GetString(CFG_IGNORE_FILE_SIGNATURE)

	autoManageBackupBLockSize := appSettings.settings.GetBoolean(CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE)
	cfg.AutoManageBackupBlockSize = &autoManageBackupBLockSize

	maxBackupBlockSize := appSettings.settings.GetInt(CFG_MAX_BACKUP_BLOCK_SIZE_MB)
	cfg.MaxBackupBlockSizeMb = &maxBackupBlockSize

	usePreviousBackup := appSettings.settings.GetBoolean(CFG_ENABLE_USE_OF_PREVIOUS_BACKUP)
	cfg.UsePreviousBackup = &usePreviousBackup

	numberOfPreviousBackupToUse := appSettings.settings.GetInt(CFG_NUMBER_OF_PREVIOUS_BACKUP_TO_USE)
	cfg.NumberOfPreviousBackupToUse = &numberOfPreviousBackupToUse

	enableLowLevelLog := appSettings.settings.GetBoolean(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC)
	cfg.EnableLowLevelLogForRsync = &enableLowLevelLog

	enableIntensiveLowLevelLog := appSettings.settings.GetBoolean(CFG_ENABLE_INTENSIVE_LOW_LEVEL_LOG_OF_RSYNC)
	cfg.EnableIntensiveLowLevelLogForRsync = &enableIntensiveLowLevelLog

	transferSourceOwner := appSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_OWNER)
	cfg.RsyncTransferSourceOwner = &transferSourceOwner

	transferSourceGroup := appSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_GROUP)
	cfg.RsyncTransferSourceGroup = &transferSourceGroup

	transferSourcePermissions := appSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS)
	cfg.RsyncTransferSourcePermissions = &transferSourcePermissions

	recreateSymlinks := appSettings.settings.GetBoolean(CFG_RSYNC_RECREATE_SYMLINKS)
	cfg.RsyncRecreateSymlinks = &recreateSymlinks

	transferDeviceFiles := appSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_DEVICE_FILES)
	cfg.RsyncTransferDeviceFiles = &transferDeviceFiles

	transferSpecialFiles := appSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SPECIAL_FILES)
	cfg.RsyncTransferSpecialFiles = &transferSpecialFiles

	compressFileTransfer := appSettings.settings.GetBoolean(CFG_RSYNC_COMPRESS_FILE_TRANSFER)
	cfg.RsyncCompressFileTransfer = &compressFileTransfer

	retry := appSettings.settings.GetInt(CFG_RSYNC_RETRY_COUNT)
	cfg.RsyncRetryCount = &retry

	modules := []backup.Module{}

	profileSettings, err := getProfileSettings(appSettings, profileID, nil)
	if err != nil {
		return nil, nil, err
	}
	sarr := profileSettings.NewSettingsArray(CFG_SOURCE_LIST)
	sourceIDs := sarr.GetArrayIDs()

	for _, sid := range sourceIDs {
		sourceSettings, err := getBackupSourceSettings(profileSettings, sid, nil)
		if err != nil {
			return nil, nil, err
		}
		if sourceSettings.settings.GetBoolean(CFG_MODULE_ENABLED) {
			module := backup.Module{}

			module.SourceRsync = strings.TrimSpace(sourceSettings.settings.GetString(CFG_MODULE_RSYNC_SOURCE_PATH))
			subpath := sourceSettings.settings.GetString(CFG_MODULE_DEST_SUBPATH)
			module.DestSubPath = normalizeSubpath(subpath)

			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_OWNER_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_OWNER)
				module.RsyncTransferSourceOwner = &value
			}
			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_GROUP_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_GROUP)
				module.RsyncTransferSourceGroup = &value
			}
			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS)
				module.RsyncTransferSourcePermissions = &value
			}
			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_RECREATE_SYMLINKS_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_RECREATE_SYMLINKS)
				module.RsyncRecreateSymlinks = &value
			}
			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_DEVICE_FILES_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_DEVICE_FILES)
				module.RsyncTransferDeviceFiles = &value
			}
			if !sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SPECIAL_FILES_INCONSISTENT) {
				value := sourceSettings.settings.GetBoolean(CFG_RSYNC_TRANSFER_SPECIAL_FILES)
				module.RsyncTransferSpecialFiles = &value
			}

			module.ChangeFilePermission = sourceSettings.settings.GetString(CFG_MODULE_CHANGE_FILE_PERMISSION)
			authPass := sourceSettings.settings.GetString(CFG_MODULE_AUTH_PASSWORD)
			if authPass != "" {
				module.AuthPassword = &authPass
			}
			modules = append(modules, module)
		}

	}

	return cfg, modules, nil
}

// getPlanInfoMarkup formats backup process totals.
func getPlanInfoMarkup(plan *backup.Plan) *Markup {
	var sourceCount int = len(plan.Nodes)
	var totalSize core.FolderSize
	var ignoreSize core.FolderSize
	var dirCount int
	for _, node := range plan.Nodes {
		totalSize += node.RootDir.GetTotalSize()
		ignoreSize += node.RootDir.GetIgnoreSize()
		dirCount += node.RootDir.GetFoldersCount()
	}
	mp := NewMarkup(0, MARKUP_COLOR_CHARTREUSE, 0, nil, nil,
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowProfileBackupPlanInfoSourceCount, nil), " "),
		NewMarkup( /*MARKUP_SIZE_LARGER*/ 0, 0, 0, sourceCount, nil),
		NewMarkup(0, 0, 0, spew.Sprintf("; %s", locale.T(MsgAppWindowProfileBackupPlanInfoTotalSize, nil)), " "),
		NewMarkup( /*MARKUP_SIZE_LARGER*/ 0, 0, 0, core.GetReadableSize(totalSize), nil),
		NewMarkup(0, 0, 0, spew.Sprintf("; %s", locale.T(MsgAppWindowProfileBackupPlanInfoSkipSize, nil)), " "),
		NewMarkup( /*MARKUP_SIZE_LARGER*/ 0, 0, 0, core.GetReadableSize(ignoreSize), nil),
		NewMarkup(0, 0, 0, spew.Sprintf("; %s", locale.T(MsgAppWindowProfileBackupPlanInfoDirectoryCount, nil)), " "),
		NewMarkup( /*MARKUP_SIZE_LARGER*/ 0, 0, 0, dirCount, nil),
	)
	return mp
}

// createHeader creates GtkHeader widget filled with children controls.
func createHeader(title, subtitle string, showCloseButton bool) (*gtk.HeaderBar, error) {
	hdr, err := SetupHeader(title, subtitle, showCloseButton)
	if err != nil {
		return nil, err
	}
	err = AddStyleClasses(&hdr.Widget, []string{"themed"})
	if err != nil {
		return nil, err
	}

	menu, err := createMenuModelForPopover()
	if err != nil {
		return nil, err
	}
	menuBtn, err := SetupMenuButtonWithThemedImage("open-menu-symbolic")
	if err != nil {
		return nil, err
	}
	menuBtn.SetUsePopover(true)
	menuBtn.SetMenuModel(menu)
	hdr.PackEnd(menuBtn)

	btn, err := SetupButtonWithThemedImage("preferences-other-symbolic")
	if err != nil {
		return nil, err
	}
	btn.SetActionName("win.PreferenceAction")
	btn.SetTooltipText(locale.T(MsgAppWindowPreferencesHint, nil))
	hdr.PackStart(btn)

	div, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	hdr.PackStart(div)

	btn, err = SetupButtonWithThemedImage("media-playback-start-symbolic")
	if err != nil {
		return nil, err
	}
	btn.SetActionName("win.RunBackupAction")
	btn.SetTooltipText(locale.T(MsgAppWindowRunBackupHint, nil))
	hdr.PackStart(btn)

	btn, err = SetupButtonWithThemedImage("media-playback-stop-symbolic")
	if err != nil {
		return nil, err
	}
	btn.SetActionName("win.StopBackupAction")
	btn.SetTooltipText(locale.T(MsgAppWindowStopBackupHint, nil))
	hdr.PackStart(btn)

	return hdr, nil
}

func createBoxWithThemedIcon(themedIconName string, cssClasses []string) (*gtk.Box, error) {
	img, err := gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	err = AddStyleClasses(&img.Widget, cssClasses)
	if err != nil {
		return nil, err
	}
	img.SetFromIconName(themedIconName, gtk.ICON_SIZE_BUTTON)
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	box.Add(img)
	return box, nil
}

func createBoxWithAssetIcon(assetIconName string) (*gtk.Box, error) {
	img, err := ImageFromAssetsNewWithResize(assetIconName, 16, 16)
	if err != nil {
		return nil, err
	}
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	box.Add(img)
	return box, nil
}

// updateDestPathWidget validate destination path and gives corresponding hint or
// error message if needed.
func updateDestPathWidget(destWidget *gtk.FileChooserButton, destControl *ControlWithStatus) error {
	destPath := destWidget.GetFilename()
	destControl.ReplaceStatus(nil)
	DEST_PATH_DESCRIPTION := locale.T(MsgAppWindowDestPathHint, nil)
	if ok, msg := isDestPathError(destPath, false); ok {
		destWidget.SetFilename("")
		markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, msg, nil),
			DEST_PATH_DESCRIPTION)
		destWidget.SetTooltipMarkup(markup.String())
		var err error
		statusBox, err := createBoxWithThemedIcon(STOCK_IMPORTANT_ICON,
			[]string{"image-error", "image-shake"})
		if err != nil {
			return err
		}
		destControl.ReplaceStatus(statusBox)
	} else {
		markup := markupTooltip(NewMarkup(0, 0, 0, nil, nil,
			NewMarkup(0, MARKUP_COLOR_CHARTREUSE, 0,
				spew.Sprintf("%s ", locale.T(MsgAppWindowDestPathIsValidStatusPart1, nil)),
				spew.Sprintf(" %s", locale.T(MsgAppWindowDestPathIsValidStatusPart2, nil)),
				NewMarkup(0, MARKUP_COLOR_CHARTREUSE, 0, spew.Sprintf("%q", destPath), nil),
			),
		), DEST_PATH_DESCRIPTION)
		destWidget.SetTooltipMarkup(markup.String())
	}
	return nil
}

// ProfileObjects keeps main form controls and settings
// in one place to pass to the functions as single parameter.
type ProfileObjects struct {
	sync.Mutex
	profileControl *ControlWithStatus
	destControl    *ControlWithStatus
	lastDestPath   string
	reselect       chan struct{}
}

func (v *ProfileObjects) CheckAndClearReselect() bool {
	select {
	case <-v.reselect:
		return true
	default:
	}
	return false
}

func (v *ProfileObjects) SetReselect() {
	v.reselect <- struct{}{}
}

func getProfileWidgetHint() string {
	return locale.T(MsgAppWindowProfileHint, nil)
}

func (v *ProfileObjects) PerformBackupPlanStage(ctx *ContextPack, supplimentary *RunningContexts,
	config *backup.Config, modules []backup.Module, cbProfile *gtk.ComboBox) error {

	supplimentary.AddContext(ctx)
	done := traceLongRunningContext(ctx)
	defer close(done)
	defer supplimentary.RemoveContext(ctx.Context)

	backupLog := core.NewProxyLog(backup.LocalLog, "backup",
		6, "15:04:05", nil, logger.InfoLevel)

	v.Lock()
	defer func() {
		v.CheckAndClearReselect()
		v.Unlock()
	}()
	v.CheckAndClearReselect()
	plan, _, err2 := backup.BuildBackupPlan(ctx.Context, backupLog, config, modules, nil)
	if err2 == nil || !rsync.IsProcessTerminatedError(err2) {
		var statusBox *gtk.Box
		if err2 == nil {
			lg.Debugf("%+v", plan)
			markup := markupTooltip(getPlanInfoMarkup(plan), getProfileWidgetHint())
			MustIdleAdd(func() {
				cbProfile.SetTooltipMarkup(markup.String())
				v.profileControl.ReplaceStatus(statusBox)
			})
		} else {
			msg := err2.Error()
			markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0,
				msg, nil), getProfileWidgetHint())
			MustIdleAdd(func() {
				statusBox, err := createBoxWithThemedIcon(STOCK_IMPORTANT_ICON,
					[]string{"image-error", "image-shake"})
				if err != nil {
					lg.Fatal(err)
				}
				cbProfile.SetTooltipMarkup(markup.String())
				v.profileControl.ReplaceStatus(statusBox)
			})
		}
	} else {
		// if termination event is not raised by profile
		// re-selection, then reset profile selection to None
		if !v.CheckAndClearReselect() {
			MustIdleAdd(func() {
				cbProfile.SetActiveID("")
			})
		}
	}
	return nil
}

func setWidgetsSensitive(sensitive bool, widgets []*gtk.Widget) {
	for _, item := range widgets {
		item.SetSensitive(sensitive)
	}
}

// createMainForm creates main form of application.
// This method is a main entry point for all GUI activity construction and display.
func createMainForm(parent context.Context, cancel func(),
	app *gtk.Application, appSettings *SettingsStore) (*gtk.ApplicationWindow, error) {

	backupSync := NewBackupSessionStatus(parent)
	supplimentary := &RunningContexts{}

	win, err := gtk.ApplicationWindowNew(app)
	if err != nil {
		return nil, err
	}
	win.SetDefaultSize(800, 150)

	_, err = win.Connect("destroy", func(window *gtk.ApplicationWindow) {
		application, err := window.GetApplication()
		if err != nil {
			lg.Fatal(err)
		}
		if backupSync.IsRunning() {
			backupSync.Stop()
		}
		supplimentary.CancelAll()
		application.Quit()
	})
	if err != nil {
		return nil, err
	}

	_, err = win.Connect("delete-event", func(window *gtk.ApplicationWindow) bool {
		quit := true
		if backupSync.IsRunning() {
			quit, err = interruptBackupProcess(&win.Window, backupSync)
			if err != nil {
				lg.Fatal(err)
			}
		}
		return !quit
	})
	if err != nil {
		return nil, err
	}

	var act glib.IAction
	var div *gtk.Separator

	act, err = createAboutAction(&win.Window, appSettings)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	act, err = createHelpAction(&win.Window)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	hdr, err := createHeader(core.GetAppTitle(), core.GetAppExtraTitle(), true)
	if err != nil {
		return nil, err
	}
	win.SetTitlebar(hdr)

	box, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		return nil, err
	}
	box.SetVAlign(gtk.ALIGN_FILL)

	box2, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	SetAllMargins(box2, 18)
	box2.SetVExpand(true)
	box2.SetVAlign(gtk.ALIGN_FILL)

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(9)
	row := 0

	lbl, err := SetupLabelJustifyRight(locale.T(MsgAppWindowProfileCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)

	lst, err := getProfileList()
	if err != nil {
		return nil, err
	}
	cbProfile, err := CreateNameValueCombo(lst)
	if err != nil {
		return nil, err
	}
	cbProfile.SetTooltipText(getProfileWidgetHint())
	cbProfile.SetActiveID("")
	cbProfile.SetHExpand(true)
	profileCtrl, err := NewControlWithStatus(&cbProfile.Widget)
	if err != nil {
		return nil, err
	}
	grid.Attach(profileCtrl.GetBox(), 1, row, 1, 1)
	row++

	box2.Add(grid)

	box.Add(box2)

	box3, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	box3.SetVExpand(true)
	box3.SetVAlign(gtk.ALIGN_FILL)

	box2.Add(box3)

	lblDestFolder, err := SetupLabelJustifyRight(locale.T(MsgAppWindowDestPathCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lblDestFolder, 0, row, 1, 1)
	destFolder, err := gtk.FileChooserButtonNew("Select destination folder", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	if err != nil {
		return nil, err
	}
	DEST_PATH_DESCRIPTION := locale.T(MsgAppWindowDestPathHint, nil)
	destFolder.SetTooltipText(DEST_PATH_DESCRIPTION)
	destFolder.SetHExpand(true)
	destFolder.SetHAlign(gtk.ALIGN_FILL)
	destCtrl, err := NewControlWithStatus(&destFolder.Widget)
	if err != nil {
		return nil, err
	}
	grid.Attach(destCtrl.GetBox(), 1, row, 1, 1)
	grid.ShowAll()
	row++

	// Make widgets disabled, until backup profile not selected.
	setWidgetsSensitive(false, []*gtk.Widget{&box3.Widget, &lblDestFolder.Widget, &destFolder.Widget})

	profileObjects := &ProfileObjects{profileControl: profileCtrl, destControl: destCtrl,
		reselect: make(chan struct{}, 1)}

	_, err = destFolder.Connect("file-set", func(dest *gtk.FileChooserButton, profileObjects *ProfileObjects) {
		destPath := dest.GetFilename()

		if profileObjects.lastDestPath != destPath {
			err := updateDestPathWidget(dest, profileObjects.destControl)
			if err != nil {
				lg.Fatal(err)
			}
			profileObjects.lastDestPath = destPath
			lg.Debugf("file-set: assign last dest path to %q", profileObjects.lastDestPath)
		}
	}, profileObjects)
	if err != nil {
		return nil, err
	}

	_, err = cbProfile.Connect("changed", func(profile *gtk.ComboBox, profileObjects *ProfileObjects) {
		cbProfile.SetTooltipText(getProfileWidgetHint())
		profileID := profile.GetActiveID()
		if profileID != "" {
			val, err := GetComboValue(profile, 0)
			if err != nil {
				lg.Fatal(err)
			}
			profileName, err := val.GetString()
			if err != nil {
				lg.Fatal(err)
			}

			profileSettings, err := getProfileSettings(appSettings, profileID, nil)
			if err != nil {
				lg.Fatal(err)
			}
			setWidgetsSensitive(true, []*gtk.Widget{&box3.Widget, &lblDestFolder.Widget, &destFolder.Widget})
			destPath := profileSettings.settings.GetString(CFG_PROFILE_DEST_ROOT_PATH)
			profileObjects.lastDestPath = destPath
			lg.Debugf("changed: assign last dest path to %q", profileObjects.lastDestPath)
			destFolder.SetFilename(destPath)
			err = updateDestPathWidget(destFolder, profileObjects.destControl)
			if err != nil {
				lg.Fatal(err)
			}

			err = enableAction(win, "RunBackupAction", true)
			if err != nil {
				lg.Fatal(err)
			}

			msg := locale.T(MsgAppWindowInquiringProfileStatus,
				struct{ ProfileName string }{ProfileName: profileName})
			markup := markupTooltip(NewMarkup(0, MARKUP_COLOR_SKY_BLUE, 0, msg, nil), getProfileWidgetHint())
			cbProfile.SetTooltipMarkup(markup.String())
			statusBox, err := createBoxWithThemedIcon(STOCK_SYNCHRONIZING_ICON, []string{"image-spin"})
			if err != nil {
				lg.Fatal(err)
			}
			profileObjects.profileControl.ReplaceStatus(statusBox)

			config, modules, err := readBackupConfig(profileID)
			if err != nil {
				lg.Fatal(err)
			}
			lg.Debugf("Modules: %+v", modules)

			// Verify that RSYNC modules configuration is valid, otherwise show error in cbProfile hint.
			if errFound, msg := isModulesConfigError(modules, false); errFound {
				markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, msg, nil),
					getProfileWidgetHint())
				cbProfile.SetTooltipMarkup(markup.String())
				var err error
				statusBox, err = createBoxWithThemedIcon(STOCK_IMPORTANT_ICON,
					[]string{"image-error", "image-shake"})
				if err != nil {
					lg.Fatal(err)
				}
				profileObjects.profileControl.ReplaceStatus(statusBox)
			} else {

				profileObjects.SetReselect()
				supplimentary.CancelAll()

				go func() {
					ctx := ForkContext(parent)

					// perform backup plan stage in one closure
					err := profileObjects.PerformBackupPlanStage(ctx, supplimentary,
						config, modules, cbProfile)
					if err != nil {
						lg.Fatal(err)
					}
				}()
			}

		} else {
			setWidgetsSensitive(false, []*gtk.Widget{&box3.Widget, &lblDestFolder.Widget, &destFolder.Widget})
			err = enableAction(win, "RunBackupAction", false)
			if err != nil {
				lg.Fatal(err)
			}
			supplimentary.CancelAll()
			profileObjects.profileControl.ReplaceStatus(nil)
		}

	}, profileObjects)
	if err != nil {
		return nil, err
	}

	act, err = createPreferenceAction(win, cbProfile)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	div, err = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	box3.Add(div)

	grid3, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid3.SetVExpand(true)
	grid3.SetVAlign(gtk.ALIGN_FILL)
	grid3.SetColumnSpacing(12)
	grid3.SetRowSpacing(9)
	row = 0
	box3.Add(grid3)

	act, err = createQuitAction(&win.Window, backupSync, supplimentary)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	act, err = createRunBackupAction(win, grid3,
		&profileObjects.lastDestPath, destFolder, cbProfile, backupSync)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	act, err = createStopBackupAction(win, grid3,
		destFolder, cbProfile, backupSync)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	win.Add(box)

	return win, nil
}

// CreateApp creates GtkApplication instance to run.
func CreateApp() (*gtk.Application, error) {
	app, err := gtk.ApplicationNew(APP_SCHEMA_ID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, err
	}

	extraMsg := locale.T(MsgSchemaConfigDlgSchemaErrorAdvise,
		struct{ ScriptName string }{ScriptName: "gs_schema_install.sh"})
	found, err := CheckSchemaSettingsIsInstalled(SETTINGS_SCHEMA_ID, app, &extraMsg)
	if err != nil {
		return nil, err
	}
	if !found {
		// Exit app because of critical issue.
		return app, nil
	}
	err = rsync.IsInstalled()
	if err != nil {
		title := locale.T(MsgAppWindowRsyncUtilityDlgTitle, nil)
		titleMarkup := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, nil, nil,
			NewMarkup(MARKUP_SIZE_LARGER, 0, 0, title, nil))
		text := locale.T(MsgAppWindowRsyncUtilityDlgNotFoundError, nil)
		err = ErrorMessage(nil, titleMarkup.String(),
			[]*DialogParagraph{NewDialogParagraph(text)})
		if err != nil {
			return nil, err
		}
		return app, nil
	}

	lang, err := GetLanguagePreference()
	if err != nil {
		lg.Fatal(err)
	}
	locale.SetLanguage(lang)

	ctx, cancel := context.WithCancel(context.Background())

	_, err = app.Application.Connect("activate", func(application *gtk.Application) {
		appSettings, err := NewSettingsStore(SETTINGS_SCHEMA_ID, SETTINGS_SCHEMA_PATH, nil)
		if err != nil {
			lg.Fatal(err)
		}

		win, err := createMainForm(ctx, cancel, application, appSettings)
		if err != nil {
			lg.Fatal(err)
		}

		win.ShowAll()
		win.SetPosition(gtk.WIN_POS_CENTER_ON_PARENT)

		// Run code, when app message queue becomes empty.
		if !appSettings.settings.GetBoolean(CFG_DONT_SHOW_ABOUT_ON_STARTUP) {
			MustIdleAdd(func() {
				actionName := "AboutAction"
				action := win.LookupAction(actionName)
				if action == nil {
					err := errors.New(locale.T(MsgActionDoesNotFound,
						struct{ ActionName string }{ActionName: actionName}))
					lg.Fatal(err)
				}
				action.Activate(nil)
			})
		}

	})
	if err != nil {
		return nil, err
	}

	// locale.GlobalLocalizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "HelloWorld"})

	return app, nil
}

// GetLanguagePreference reads application language preference customized by user.
func GetLanguagePreference() (string, error) {
	appSettings, err := glib.SettingsNew(SETTINGS_SCHEMA_ID)
	if err != nil {
		return "", err
	}
	lang := appSettings.GetString(CFG_UI_LANGUAGE)
	return lang, nil
}
