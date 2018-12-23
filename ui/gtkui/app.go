package gtkui

import (
	"context"
	"errors"
	"os"
	"syscall"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/backup"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/data"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	shell "github.com/d2r2/go-shell"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// createQuitAction creates regular exit app action.
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

		quit := true
		if backupSync.IsRunning() {
			quit, err = interruptBackupDialog(win)
			if err != nil {
				lg.Fatal(err)
			}
		}

		if quit {
			application, err := win.GetApplication()
			if err != nil {
				lg.Fatal(err)
			}
			if backupSync.IsRunning() {
				backupSync.Stop()
			}
			supplimentary.CancelAll()
			application.Quit()
		}
	})
	if err != nil {
		return nil, err
	}

	return act, nil
}

// createAboutAction creates "about dialog" action.
func createAboutAction(win *gtk.Window, appSettings *glib.Settings) (glib.IAction, error) {
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

// createMenuModelForPopover construct menu for popover button
func createMenuModelForPopover() (glib.IMenuModel, error) {
	main, err := glib.MenuNew()
	if err != nil {
		return nil, err
	}

	var section *glib.Menu
	//var item *glib.MenuItem

	// New menu section (with buttons)
	section, err = glib.MenuNew()
	if err != nil {
		return nil, err
	}
	section.Append(locale.T(MsgAppWindowAboutMenuCaption, nil), "win.AboutAction")
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

// createToolbar creates GTK Toolbar.
func createToolbar() (*gtk.Toolbar, error) {
	tbx, err := gtk.ToolbarNew()
	if err != nil {
		return nil, err
	}
	tbx.SetStyle(gtk.TOOLBAR_BOTH_HORIZ)

	var tbtn *gtk.ToolButton
	var tdvd *gtk.SeparatorToolItem
	var img *gtk.Image

	img, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	img.SetFromIconName("application-exit-symbolic", gtk.ICON_SIZE_BUTTON)

	tbtn, err = gtk.ToolButtonNew(img, "")
	if err != nil {
		return nil, err
	}
	tbtn.SetActionName("win.QuitAction")
	tbx.Add(tbtn)

	tdvd, err = gtk.SeparatorToolItemNew()
	if err != nil {
		return nil, err
	}
	tbx.Add(tdvd)

	img, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	img.SetFromIconName("preferences-other-symbolic", gtk.ICON_SIZE_BUTTON)

	/*
		file, err := data.Assets.Open("ajax-loader-gears_32x32.gif")
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}
		b2, err := glib.BytesNew(b)
		if err != nil {
			return nil, err
		}
		ms, err := glib.MemoryInputStreamFromBytesNew(b2)
		if err != nil {
			return nil, err
		}
		pb, err := gdk.PixbufNewFromStream(&ms.InputStream, nil)
		if err != nil {
			return nil, err
		}
		img, err = gtk.ImageNewFromPixbuf(pb)
		if err != nil {
			return nil, err
		}
	*/

	tbtn, err = gtk.ToolButtonNew(img, "")
	if err != nil {
		return nil, err
	}
	tbtn.SetActionName("win.PreferenceAction")
	tbx.Add(tbtn)

	img, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	img.SetFromIconName("help-about-symbolic", gtk.ICON_SIZE_BUTTON)
	tbtn, err = gtk.ToolButtonNew(img, "")
	if err != nil {
		return nil, err
	}
	tbtn.SetActionName("win.AboutAction")
	tbx.Add(tbtn)

	tdvd, err = gtk.SeparatorToolItemNew()
	if err != nil {
		return nil, err
	}
	tbx.Add(tdvd)

	img, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	img.SetFromIconName("media-playback-start-symbolic", gtk.ICON_SIZE_BUTTON)
	tbtn, err = gtk.ToolButtonNew(img, "")
	if err != nil {
		return nil, err
	}
	tbtn.SetActionName("win.RunBackupAction")
	tbx.Add(tbtn)

	img, err = gtk.ImageNew()
	if err != nil {
		return nil, err
	}
	img.SetFromIconName("media-playback-stop-symbolic", gtk.ICON_SIZE_BUTTON)
	tbtn, err = gtk.ToolButtonNew(img, "")
	if err != nil {
		return nil, err
	}
	tbtn.SetActionName("win.StopBackupAction")
	tbx.Add(tbtn)

	return tbx, nil
}

// createPreferenceAction creates multi-page preference dialog
// with save/restore functionality to/from the GLib GSettings object.
// Action activation require to have GLib Setting Schema
// preliminary installed, otherwise will not work raising error.
// Installation bash script from app folder must be performed in advance.
func createPreferenceAction(win *gtk.Window, profile *gtk.ComboBox) (glib.IAction, error) {
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

		app, err := win.GetApplication()
		if err != nil {
			lg.Fatal(err)
		}

		extraMsg := locale.T(MsgSchemaConfigDlgSchemaErrorAdvise,
			struct{ ScriptName string }{ScriptName: "gs_schema_install.sh"})
		found, err := CheckSchemaSettingsIsInstalled(core.SETTINGS_ID, app, &extraMsg)
		if err != nil {
			lg.Fatal(err)
		}

		if found {
			var profileChanged bool
			win, err := CreatePreferenceDialog(core.SETTINGS_ID, app, &profileChanged)
			if err != nil {
				lg.Fatal(err)
			}

			win.ShowAll()
			win.Show()

			_, err = win.Connect("destroy", func(window *gtk.ApplicationWindow) {
				win.Close()
				win.Destroy()
				lg.Debug("Destroy window")

				if profileChanged {
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
	if act != nil {
		action, err := glib.SimpleActionFromAction(act)
		if err != nil {
			return err
		}
		action.SetEnabled(enable)
	} else {
		err := errors.New(spew.Sprintf("action %q does not found", actionName))
		return err
	}
	return nil
}

type EmptySpaceRecover struct {
	main      *gtk.ApplicationWindow
	backupLog logger.PackageLog
}

func (v *EmptySpaceRecover) ErrorHook(err error, paths core.SrcDstPath, predictedSize *core.FolderSize,
	repeated int, retryLeft int) (newRetryLeft int, criticalError error) {

	if rsync.IsRsyncCallFailedError(err) {
		erro := err.(*rsync.RsyncCallFailedError)
		freeSpace, err2 := shell.GetFreeSpace(paths.DestPath)
		if err2 != nil {
			return retryLeft, err2
		}

		// if space/1000/1000 < 10 {
		var predSize uint64
		if predictedSize != nil {
			predSize = predictedSize.GetByteCount()
		}

		const MB = 1000 * 1000

		lg.Debugf("Exit code = %d, error = %v, predicted size = %d MB, retry left = %d, space left: %d kB",
			erro.ExitCode, erro, predSize/1000/1000, retryLeft, freeSpace/1000)

		if (erro.ExitCode == 23 || erro.ExitCode == 11) &&
			(predictedSize == nil && freeSpace < 1*MB || predictedSize.GetByteCount() > freeSpace) {

			ch := make(chan OutOfSpaceResponse)
			defer close(ch)

			v.backupLog.Notifyf(locale.T(MsgLogBackupStageOutOfSpaceWarning,
				struct{ SizeLeft string }{SizeLeft: core.FormatSize(freeSpace, true)}))

			_, err2 = glib.IdleAdd(func() {

				response, err2 := outOfSpaceDialog(&v.main.Window, paths, freeSpace)
				if err2 != nil {
					lg.Fatal(err2)
				}
				ch <- response

			})
			if err2 != nil {
				lg.Fatal(err2)
			}

			response, _ := <-ch
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
		// }
	}
	return retryLeft, nil
}

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

func performFullBackup(backupSync *BackupSessionStatus, notifier *NotifierUI,
	win *gtk.ApplicationWindow, config *backup.Config, destPath string) {

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

	plan, progress, err := backup.BuildBackupPlan(ctx.Context, backupLog, config, notifier)
	if err == nil {
		lg.Debugf("Backup node's dir trees: %+v", plan)

		emptySpaceRecover := &EmptySpaceRecover{main: win, backupLog: backupLog}
		err = plan.RunBackup(progress, destPath, emptySpaceRecover.ErrorHook)

		notifier.ReportCompletion(1, err, progress, true)
		progress.Close()
	} else {
		notifier.ReportCompletion(0, err, nil, true)
	}
}

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
	_, err := glib.IdleAdd(call)
	if err != nil {
		lg.Fatal(err)
	}
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

		backupID := profile.GetActiveID()
		lg.Debugf("BackupID = %v", backupID)

		if backupID != "" {
			if ok, msg := isDestPathError(*destPath, true); ok {
				title := locale.T(MsgAppWindowCannotStartBackupProcessTitle, nil)
				var text string
				if *destPath == "" {
					text = locale.T(MsgAppWindowDestPathIsEmptyError2, nil)
				} else {
					text = msg
				}
				err = ErrorMessage(&win.Window, title, []*DialogParagraph{NewDialogParagraph(text)})
			} else {
				config, err := readBackupConfig(backupID)
				if err != nil {
					lg.Fatal(err)
				}
				// enable/disable corresponding UI elements
				setControlStateOnBackupStarted(win, selectFolder, profile)

				appSettings, err := glib.SettingsNew(core.SETTINGS_ID)
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
					performFullBackup(backupSync, notifier, win, config, *destPath)
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

		if ok, err := interruptBackupDialog(&win.Window); err != nil || ok {
			if err != nil {
				lg.Fatal(err)
			}

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
		} else {
			err = enableAction(win, "StopBackupAction", true)
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

// getProfileList reads from app configuration profile's identifiers and names
// to use as a source for GtkComboBox widget.
func getProfileList() ([]struct{ value, key string }, error) {
	appSettings, err := glib.SettingsNew(core.SETTINGS_ID)
	if err != nil {
		return nil, err
	}
	sarr := NewSettingsArray(appSettings, CFG_BACKUP_LIST)
	lst := sarr.GetArrayIDs()
	arr := []struct{ value, key string }{{locale.T(MsgAppWindowNoneProfileEntry, nil), ""}}
	for _, item := range lst {
		backupSettings, err := getBackupSettings(item, nil)
		if err != nil {
			return nil, err
		}
		name := backupSettings.GetString(CFG_PROFILE_NAME)
		arr = append(arr, struct{ value, key string }{name, item})
	}
	return arr, nil
}

// readBackupConfig reads from app configuration Config object which
// contains all settings necessary to run new backup session.
func readBackupConfig(backupID string) (*backup.Config, error) {
	appSettings, err := glib.SettingsNew(core.SETTINGS_ID)
	if err != nil {
		return nil, err
	}
	backupSettings, err := getBackupSettings(backupID, nil)
	if err != nil {
		return nil, err
	}
	sarr := NewSettingsArray(backupSettings, CFG_SOURCE_LIST)
	sourceIDs := sarr.GetArrayIDs()

	cfg := &backup.Config{}

	cfg.SigFileIgnoreBackup = appSettings.GetString(CFG_IGNORE_FILE_SIGNATURE)

	autoManageBackupBLockSize := appSettings.GetBoolean(CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE)
	cfg.AutoManageBackupBlockSize = &autoManageBackupBLockSize

	maxBackupBlockSize := appSettings.GetInt(CFG_MAX_BACKUP_BLOCK_SIZE_MB)
	cfg.MaxBackupBlockSizeMb = &maxBackupBlockSize

	usePreviousBackup := appSettings.GetBoolean(CFG_ENABLE_USE_OF_PREVIOUS_BACKUP)
	cfg.UsePreviousBackup = &usePreviousBackup

	numberOfPreviousBackupToUse := appSettings.GetInt(CFG_NUMBER_OF_PREVIOUS_BACKUP_TO_USE)
	cfg.NumberOfPreviousBackupToUse = &numberOfPreviousBackupToUse

	enableLowLevelLog := appSettings.GetBoolean(CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC)
	cfg.EnableLowLevelLogForRsync = &enableLowLevelLog

	enableIntensiveLowLevelLog := appSettings.GetBoolean(CFG_ENABLE_INTENSIVE_LOW_LEVEL_LOG_OF_RSYNC)
	cfg.EnableIntensiveLowLevelLogForRsync = &enableIntensiveLowLevelLog

	compressFileTransfer := appSettings.GetBoolean(CFG_RSYNC_COMPRESS_FILE_TRANSFER)
	cfg.RsyncCompressFileTransfer = &compressFileTransfer

	recreateSymlinks := appSettings.GetBoolean(CFG_RSYNC_RECREATE_SYMLINKS)
	cfg.RsyncRecreateSymlinks = &recreateSymlinks

	transferDeviceFiles := appSettings.GetBoolean(CFG_RSYNC_TRANSFER_DEVICE_FILES)
	cfg.RsyncTransferDeviceFiles = &transferDeviceFiles

	transferSpecialFiles := appSettings.GetBoolean(CFG_RSYNC_TRANSFER_SPECIAL_FILES)
	cfg.RsyncTransferSpecialFiles = &transferSpecialFiles

	transferSourceOwner := appSettings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_OWNER)
	cfg.RsyncTransferSourceOwner = &transferSourceOwner

	transferSourceGroup := appSettings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_GROUP)
	cfg.RsyncTransferSourceGroup = &transferSourceGroup

	transferSourcePermissions := appSettings.GetBoolean(CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS)
	cfg.RsyncTransferSourcePermissions = &transferSourcePermissions

	retry := appSettings.GetInt(CFG_RSYNC_RETRY_COUNT)
	cfg.RsyncRetryCount = &retry

	for _, sid := range sourceIDs {
		sourceSettings, err := getBackupSourceSettings(backupID, sid, nil)
		if err != nil {
			return nil, err
		}
		if sourceSettings.GetBoolean(CFG_SOURCE_ENABLED) {
			node := backup.BackupNode{}
			node.SourceRsync = sourceSettings.GetString(CFG_SOURCE_RSYNC_SOURCE_PATH)
			subpath := sourceSettings.GetString(CFG_SOURCE_DEST_SUBPATH)
			node.DestSubPath = normalizeSubpath(subpath)
			cfg.BackupNodes = append(cfg.BackupNodes, node)
		}

	}

	return cfg, nil
}

// getPlanInfoMarkup formats backup process totals.
func getPlanInfoMarkup(plan *backup.BackupPlan) *Markup {
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
	/*
		s := spew.Sprintf("RSYNC sources: %v; Total size: %v; Skip size: %v; Directory count: %v",
			MarkupTag("big", sourceCount),
			MarkupTag("big", hum.Bytes(totalSize.GetByteCount())),
			MarkupTag("big", hum.Bytes(ignoreSize.GetByteCount())),
			MarkupTag("big", dirCount))
	*/
	return mp
}

func getSpaceBox() (*gtk.Box, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	box.SetSizeRequest(3, -1)
	return box, nil
}

// createHeader creates GtkHeader widget filled with children controls.
func createHeader(title, subtitle string, showCloseButton bool) (*gtk.HeaderBar, error) {
	hdr, err := SetupHeader(title, subtitle, showCloseButton)
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

	var box *gtk.Box

	btn, err := SetupButtonWithThemedImage("preferences-other-symbolic")
	//btn, err := SetupButtonWithAssetAnimationImage("ajax-loader-gears_32x32.gif")
	if err != nil {
		return nil, err
	}
	btn.SetActionName("win.PreferenceAction")
	btn.SetTooltipText(locale.T(MsgAppWindowPreferencesHint, nil))
	hdr.PackStart(btn)

	box, err = getSpaceBox()
	if err != nil {
		return nil, err
	}
	hdr.PackStart(box)

	div, err := gtk.SeparatorNew(gtk.ORIENTATION_VERTICAL)
	if err != nil {
		return nil, err
	}
	hdr.PackStart(div)

	box, err = getSpaceBox()
	if err != nil {
		return nil, err
	}
	hdr.PackStart(box)

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

	box, err = getSpaceBox()
	if err != nil {
		return nil, err
	}
	hdr.PackStart(box)

	return hdr, nil
}

func createBoxWithThemedIcon(themedIconName string) (*gtk.Box, error) {
	img, err := gtk.ImageNew()
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
	img, err := ImageFromAssetsNew(assetIconName, 16, 16)
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

func createBoxWithSpinner() (*gtk.Box, error) {
	spinner, err := gtk.SpinnerNew()
	if err != nil {
		return nil, err
	}
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 0)
	if err != nil {
		return nil, err
	}
	box.Add(spinner)
	spinner.Start()
	return box, nil
}

func updateDestPathWidget(destWidget *gtk.FileChooserButton, destFolderBox *gtk.Box, destFolderStatusBox **gtk.Box) error {
	destPath := destWidget.GetFilename()
	if *destFolderStatusBox != nil {
		(*destFolderStatusBox).Destroy()
		*destFolderStatusBox = nil
	}
	DEST_PATH_DESCRIPTION := locale.T(MsgAppWindowDestPathHint, nil)
	if ok, msg := isDestPathError(destPath, false); ok {
		destWidget.SetFilename("")
		markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0, msg, nil),
			DEST_PATH_DESCRIPTION)
		destWidget.SetTooltipMarkup(markup.String())
		var err error
		// *destFolderStatusBox, err = createBoxWithThemedIcon(STOCK_IMPORTANT_ICON)
		*destFolderStatusBox, err = createBoxWithAssetIcon(ASSET_IMPORTANT_ICON)
		if err != nil {
			return err
		}
		destFolderBox.Add(*destFolderStatusBox)
		destFolderBox.ShowAll()
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

func getProfileWidgetHint() string {
	return locale.T(MsgAppWindowProfileHint, nil)
}

func performBackupPlanStage(ctx *ContextPack, supplimentary *RunningContexts, config *backup.Config,
	cbProfile *gtk.ComboBox, cbProfileBox *gtk.Box, cbProfileStatusBox *gtk.Box) error {

	supplimentary.AddContext(ctx)
	done := traceLongRunningContext(ctx)
	defer close(done)
	defer supplimentary.RemoveContext(ctx.Context)

	backupLog := core.NewProxyLog(backup.LocalLog, "backup",
		6, "15:04:05", nil, logger.InfoLevel)

	plan, _, err2 := backup.BuildBackupPlan(ctx.Context, backupLog, config, nil)
	if err2 == nil || !rsync.IsRsyncProcessTerminatedError(err2) {
		_, err := glib.IdleAdd(func() {
			if cbProfileStatusBox != nil {
				cbProfileStatusBox.Destroy()
				cbProfileStatusBox = nil
			}
			if err2 == nil {
				lg.Debugf("%+v", plan)
				markup := markupTooltip(getPlanInfoMarkup(plan), getProfileWidgetHint())
				cbProfile.SetTooltipMarkup(markup.String())
			} else {
				msg := err2.Error()
				var err error
				cbProfileStatusBox, err = createBoxWithAssetIcon(ASSET_IMPORTANT_ICON)
				if err != nil {
					lg.Fatal(err)
				}
				markup := markupTooltip(NewMarkup(MARKUP_WEIGHT_BOLD, MARKUP_COLOR_ORANGE_RED, 0,
					msg, nil), getProfileWidgetHint())
				cbProfile.SetTooltipMarkup(markup.String())
			}
			if cbProfileStatusBox != nil {
				cbProfileBox.Add(cbProfileStatusBox)
				cbProfileBox.ShowAll()
			}
		})
		if err != nil {
			return err
		}
	} else {
		_, err := glib.IdleAdd(func() {
			cbProfile.SetActiveID("")
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// createMainForm creates main form of application.
// This method is a main entry point for all GUI activity construction and display.
func createMainForm(parent context.Context, cancel func(),
	app *gtk.Application, appSettings *glib.Settings) (*gtk.ApplicationWindow, error) {

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
			quit, err = interruptBackupDialog(&window.Window)
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
	grid.SetRowSpacing(6)
	row := 0

	lbl, err := setupLabelJustifyRight(locale.T(MsgAppWindowProfileCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)

	cbProfileBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		return nil, err
	}
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
	cbProfileBox.Add(cbProfile)
	cbProfileBox.SetHExpand(true)
	grid.Attach(cbProfileBox, 1, row, 1, 1)
	row++

	box2.Add(grid)

	div, err = gtk.SeparatorNew(gtk.ORIENTATION_HORIZONTAL)
	if err != nil {
		return nil, err
	}
	box2.Add(div)

	box.Add(box2)

	box3, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
	if err != nil {
		return nil, err
	}
	box3.SetVExpand(true)
	box3.SetVAlign(gtk.ALIGN_FILL)
	box3.SetSensitive(false)

	box2.Add(box3)

	grid, err = gtk.GridNew()
	if err != nil {
		return nil, err
	}
	grid.SetColumnSpacing(12)
	grid.SetRowSpacing(6)
	row = 0
	box3.Add(grid)

	lbl, err = setupLabelJustifyRight(locale.T(MsgAppWindowDestPathCaption, nil))
	if err != nil {
		return nil, err
	}
	grid.Attach(lbl, 0, row, 1, 1)
	destFolderBox, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		return nil, err
	}
	destFolder, err := gtk.FileChooserButtonNew("Select destination folder", gtk.FILE_CHOOSER_ACTION_SELECT_FOLDER)
	if err != nil {
		return nil, err
	}
	DEST_PATH_DESCRIPTION := locale.T(MsgAppWindowDestPathHint, nil)
	destFolder.SetTooltipText(DEST_PATH_DESCRIPTION)
	destFolder.SetHExpand(true)
	destFolder.SetHAlign(gtk.ALIGN_FILL)
	destFolderBox.Add(destFolder)
	destFolderBox.SetHExpand(true)

	grid.Attach(destFolderBox, 1, row, 1, 1)
	grid.ShowAll()
	row++

	var cbProfileStatusBox *gtk.Box
	var destFolderStatusBox *gtk.Box
	var lastDestPath string

	_, err = destFolder.Connect("file-set", func(dest *gtk.FileChooserButton, lastDestPath *string) {
		destPath := dest.GetFilename()
		//destFolder.SetTooltipText(destPath)

		if *lastDestPath != destPath {
			err := updateDestPathWidget(dest, destFolderBox, &destFolderStatusBox)
			if err != nil {
				lg.Fatal(err)
			}
			*lastDestPath = destPath
			lg.Debugf("file-set: assign last dest path to %q", *lastDestPath)
		}
	}, &lastDestPath)
	if err != nil {
		return nil, err
	}

	_, err = cbProfile.Connect("changed", func(profile *gtk.ComboBox, lastDestPath *string) {
		cbProfile.SetTooltipText(getProfileWidgetHint())
		backupID := profile.GetActiveID()
		if backupID != "" {
			val, err := GetComboValue(profile, 0)
			if err != nil {
				lg.Fatal(err)
			}
			profileName, err := val.GetString()
			if err != nil {
				lg.Fatal(err)
			}

			backupSettings, err := getBackupSettings(backupID, nil)
			if err != nil {
				lg.Fatal(err)
			}
			box3.SetSensitive(true)
			destPath := backupSettings.GetString(CFG_PROFILE_DEST_ROOT_PATH)
			*lastDestPath = destPath
			lg.Debugf("changed: assign last dest path to %q", *lastDestPath)
			destFolder.SetFilename(destPath)
			err = updateDestPathWidget(destFolder, destFolderBox, &destFolderStatusBox)
			if err != nil {
				lg.Fatal(err)
			}

			err = enableAction(win, "RunBackupAction", true)
			if err != nil {
				lg.Fatal(err)
			}

			if cbProfileStatusBox != nil {
				cbProfileStatusBox.Destroy()
				cbProfileStatusBox = nil
			}
			msg := locale.T(MsgAppWindowInquiringProfileStatus,
				struct{ ProfileName string }{ProfileName: profileName})
			markup := markupTooltip(NewMarkup(0, MARKUP_COLOR_SKY_BLUE, 0, msg, nil), getProfileWidgetHint())
			cbProfile.SetTooltipMarkup(markup.String())
			cbProfileStatusBox, err = createBoxWithSpinner()
			if err != nil {
				lg.Fatal(err)
			}
			cbProfileBox.Add(cbProfileStatusBox)
			cbProfileBox.ShowAll()

			config, err := readBackupConfig(backupID)
			if err != nil {
				lg.Fatal(err)
			}

			supplimentary.CancelAll()

			go func() {
				ctx := ForkContext(parent)

				// perform backup plan stage in one closure
				err := performBackupPlanStage(ctx, supplimentary, config, cbProfile, cbProfileBox, cbProfileStatusBox)
				if err != nil {
					lg.Fatal(err)
				}
			}()

		} else {
			box3.SetSensitive(false)
			err = enableAction(win, "RunBackupAction", false)
			if err != nil {
				lg.Fatal(err)
			}
			supplimentary.CancelAll()
			if cbProfileStatusBox != nil {
				cbProfileStatusBox.Destroy()
				cbProfileStatusBox = nil
			}
		}

	}, &lastDestPath)
	if err != nil {
		return nil, err
	}

	act, err = createPreferenceAction(&win.Window, cbProfile)
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
	grid3.SetRowSpacing(6)
	//grid3.SetSensitive(false)
	row = 0
	box3.Add(grid3)

	act, err = createQuitAction(&win.Window, backupSync, supplimentary)
	if err != nil {
		return nil, err
	}
	win.AddAction(act)

	act, err = createRunBackupAction(win, grid3,
		&lastDestPath, destFolder, cbProfile, backupSync)
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

	file, err := data.Assets.Open("./ajax-loader-gears_32x32.gif")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	app, err := gtk.ApplicationNew(core.APP_ID, glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, err
	}

	extraMsg := locale.T(MsgSchemaConfigDlgSchemaErrorAdvise,
		struct{ ScriptName string }{ScriptName: "gs_schema_install.sh"})
	found, err := CheckSchemaSettingsIsInstalled(core.SETTINGS_ID, app, &extraMsg)
	if err != nil {
		return nil, err
	}
	if !found {
		// exit app
		return app, nil
	}
	err = rsync.IsInstalled()
	if err != nil {
		text := locale.T(MsgAppWindowRsyncUtilityDlgNotFoundError, nil)
		err = ErrorMessage(nil, locale.T(MsgAppWindowRsyncUtilityDlgTitle, nil),
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
		appSettings, err := glib.SettingsNew(core.SETTINGS_ID)
		if err != nil {
			lg.Fatal(err)
		}

		win, err := createMainForm(ctx, cancel, application, appSettings)
		if err != nil {
			lg.Fatal(err)
		}

		win.ShowAll()
		//win.SetGravity(gdk.GDK_GRAVITY_CENTER)
		//win.Move(gdk.ScreenWidth()/2, gdk.ScreenHeight()/2)
		win.SetPosition(gtk.WIN_POS_CENTER_ON_PARENT)

		// Run code, when app message queue becomes empty.
		if !appSettings.GetBoolean(CFG_DONT_SHOW_ABOUT_ON_STARTUP) {
			_, err := glib.IdleAdd(func() {
				action := win.LookupAction("AboutAction")
				if action != nil {
					action.Activate(nil)
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

	locale.Localizer.MustLocalize(&i18n.LocalizeConfig{MessageID: "HelloWorld"})

	return app, nil
}

func GetLanguagePreference() (string, error) {
	appSettings, err := glib.SettingsNew(core.SETTINGS_ID)
	if err != nil {
		return "", err
	}
	lang := appSettings.GetString(CFG_UI_LANGUAGE)
	return lang, nil
}
