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
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/backup"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
	shell "github.com/d2r2/go-shell"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/d2r2/gotk3/libnotify"
	"github.com/d2r2/gotk3/pango"
	"github.com/davecgh/go-spew/spew"
)

// NotifierUI is an object, than bind backup process
// notifications with application GUI controls.
type NotifierUI struct {
	profileName string
	gridUI      *gtk.Grid
	totalDone   core.FolderSize
	// keep overall progress percentage
	progress *float32
	// flag informing that backup process is finalized in asynchronous GUI controls
	done chan struct{}
	// GUI GTK widgets
	pbm         *ProgressBarManage
	statusLabel *gtk.Label
	logTextView *gtk.TextView
	logViewPort *gtk.Viewport
}

// Static cast to verify that struct implement specific interface.
var _ backup.Notifier = &NotifierUI{}

func NewNotifierUI(profileName string, gridUI *gtk.Grid) *NotifierUI {
	v := &NotifierUI{profileName: profileName, gridUI: gridUI, done: make(chan struct{})}
	return v
}

func (v *NotifierUI) Done() chan struct{} {
	return v.done
}

func formatInqueryProgress(sourceID int, sourceRsync string) string {
	mp := NewMarkup(0, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, locale.T(MsgAppWindowBackupProgressInquiringSourceID,
			struct{ SourceID int }{SourceID: sourceID + 1}), spew.Sprintln()),
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowBackupProgressInquiringSourceDescription,
			struct{ RsyncSource string }{RsyncSource: sourceRsync}), nil),
	)
	return mp.String()
}

// NotifyPlanStage_NodeStructureStartInquiry implements core.BackupNotifier interface method.
func (v *NotifierUI) NotifyPlanStage_NodeStructureStartInquiry(sourceID int,
	sourceRsync string) error {
	msg := formatInqueryProgress(sourceID, sourceRsync)
	err := v.UpdateBackupProgress(nil, msg, true)
	if err != nil {
		lg.Fatal(err)
	}
	return nil
}

// NotifyPlanStage_NodeStructureDoneInquiry implements core.BackupNotifier interface method.
func (v *NotifierUI) NotifyPlanStage_NodeStructureDoneInquiry(sourceID int,
	sourceRsync string, dir *core.Dir) error {
	return nil
}

// formatBackupProgress build markup text to detail progress status.
func formatBackupProgress(backupType core.FolderBackupType, totalDone, leftToBackup core.FolderSize,
	timePassed time.Duration, eta *time.Duration, path string) string {

	sections := 2
	etaStr := "*"
	if eta != nil {
		etaStr = core.FormatDurationToDaysHoursMinsSecs(*eta, true, &sections)
	}
	passedStr := core.FormatDurationToDaysHoursMinsSecs(timePassed, true, &sections)
	mp := NewMarkup(0, 0, 0, nil, nil,
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, passedStr, " "),
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowBackupProgressTimePassedSuffix, nil), " | "),
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, etaStr, " "),
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowBackupProgressETASuffix, nil), "\n"),
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, core.GetReadableSize(totalDone), " "),
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowBackupProgressSizeCompletedSuffix, nil), " | "),
		NewMarkup(MARKUP_SIZE_LARGER, 0, 0, core.GetReadableSize(leftToBackup), " "),
		NewMarkup(0, 0, 0, locale.T(MsgAppWindowBackupProgressSizeLeftToProcessSuffix, nil), "\n"),
		NewMarkup(0, 0, 0, spew.Sprintf("%s: %q", backup.GetBackupTypeDescription(backupType), path),
			nil),
	)
	return mp.String()
}

// NotifyBackupStage_FolderStartBackup implements core.BackupNotifier interface method.
// Called by backup process when next piece of data backup started.
func (v *NotifierUI) NotifyBackupStage_FolderStartBackup(rootDest string,
	paths core.SrcDstPath, backupType core.FolderBackupType,
	leftToBackup core.FolderSize,
	timePassed time.Duration, eta *time.Duration) error {

	path, err := core.GetRelativePath(rootDest, paths.DestPath)
	if err != nil {
		return err
	}

	msg := formatBackupProgress(backupType, v.totalDone, leftToBackup, timePassed, eta, path)

	err = v.UpdateBackupProgress(v.progress, msg, true)
	if err != nil {
		lg.Fatal(err)
	}

	return err
}

// NotifyBackupStage_FolderDoneBackup implements core.BackupNotifier interface method.
// Called by backup process when next piece of data backup ended.
func (v *NotifierUI) NotifyBackupStage_FolderDoneBackup(rootDest string,
	paths core.SrcDstPath, backupType core.FolderBackupType,
	leftToBackup core.FolderSize, sizeDone core.SizeProgress,
	timePassed time.Duration, eta *time.Duration,
	sessionErr error) error {

	path, err := core.GetRelativePath(rootDest, paths.DestPath)
	if err != nil {
		return err
	}

	v.totalDone = v.totalDone.AddSizeProgress(sizeDone)

	msg := formatBackupProgress(backupType, v.totalDone, leftToBackup, timePassed, eta, path)

	lg.Debugf("Total done: %v", v.totalDone)
	lg.Debugf("Left to backup: %v", leftToBackup.GetByteCount())
	progress := float32(float64(v.totalDone) / float64(v.totalDone+leftToBackup))
	const minProgress = 0.002
	if progress < minProgress {
		progress = minProgress
	}
	v.progress = &progress

	err = v.UpdateBackupProgress(v.progress, msg, true)
	if err != nil {
		lg.Fatal(err)
	}

	return err
}

// ClearProgressGrid remove and delete GTK widgets containing information about previous backup session.
func (v *NotifierUI) ClearProgressGrid() error {
	v.statusLabel = nil
	if v.pbm != nil {
		v.pbm.StopPulse()
		v.pbm = nil
	}
	v.logTextView = nil
	v.logViewPort = nil
	lst := v.gridUI.GetChildren()
	lst.Foreach(func(item interface{}) {
		if wdg, ok := item.(*gtk.Widget); ok {
			wdg.Destroy()
		}
	})
	/*
		for gtk.EventsPending() {
			lg.Info("Pending events 2")
			gtk.MainIteration()
		}
	*/
	return nil
}

// CreateProgressControls create GTK widgets which will indicate backup session progress.
func (v *NotifierUI) CreateProgressControls(sessionLogFontSize string) error {
	row := 0
	if v.pbm == nil {
		lbl, err := gtk.LabelNew(locale.T(MsgAppWindowOverallProgressCaption, nil))
		if err != nil {
			return err
		}
		lbl.SetHAlign(gtk.ALIGN_START)
		v.gridUI.Attach(lbl, 0, row, 1, 1)
		progressBar, err := gtk.ProgressBarNew()
		if err != nil {
			return err
		}
		// err = ApplyStyleCSS(&progressBar.Widget, css)
		// if err != nil {
		// 	return err
		// }
		progressBar.SetHAlign(gtk.ALIGN_FILL)
		progressBar.SetHExpand(true)
		// AddStyleClass(&progressBar.Widget, "run-animation")
		v.pbm = NewProgressBarManage(progressBar)
		_, err = progressBar.Connect("destroy", func(pb *gtk.ProgressBar, pbm *ProgressBarManage) {
			pbm.StopPulse()
		}, v.pbm)
		if err != nil {
			return err
		}

		v.gridUI.Attach(progressBar, 1, row, 1, 1)
	}
	row++

	if v.statusLabel == nil {
		lbl, err := gtk.LabelNew(locale.T(MsgAppWindowProgressStatusCaption, nil))
		if err != nil {
			return err
		}
		lbl.SetHAlign(gtk.ALIGN_START)
		lbl.SetVAlign(gtk.ALIGN_START)
		v.gridUI.Attach(lbl, 0, row, 1, 1)
		v.statusLabel, err = gtk.LabelNew("")
		if err != nil {
			return err
		}
		v.statusLabel.SetHAlign(gtk.ALIGN_START)
		v.statusLabel.SetHExpand(true)
		v.statusLabel.SetEllipsize(pango.ELLIPSIZE_MIDDLE)
		v.gridUI.Attach(v.statusLabel, 1, row, 1, 1)
	}
	row++

	if v.logTextView == nil {
		lbl, err := gtk.LabelNew(locale.T(MsgAppWindowSessionLogCaption, nil))
		if err != nil {
			return err
		}
		lbl.SetHAlign(gtk.ALIGN_START)
		v.gridUI.Attach(lbl, 0, row, 2, 1)
		row++
		v.logTextView, err = gtk.TextViewNew()
		if err != nil {
			return err
		}
		buffer, err := v.logTextView.GetBuffer()
		if err != nil {
			return err
		}
		err = addColorTags(buffer)
		if err != nil {
			return err
		}

		css := `
textview {
    font: %s "Monospace";
}
		`
		err = ApplyStyleCSS(&v.logTextView.Widget, spew.Sprintf(css, sessionLogFontSize))
		if err != nil {
			return err
		}
		v.logTextView.SetEditable(false)
		v.logViewPort, err = gtk.ViewportNew(nil, nil)
		if err != nil {
			return err
		}
		sw, err := gtk.ScrolledWindowNew(nil, nil)
		if err != nil {
			return err
		}
		sw.SetSizeRequest(-1, 120)
		sw.SetVAlign(gtk.ALIGN_FILL)
		sw.SetVExpand(true)
		sw.Add(v.logViewPort)
		v.logViewPort.Add(v.logTextView)
		v.gridUI.Attach(sw, 0, row, 2, 1)
	}
	row++

	v.gridUI.ShowAll()
	return nil
}

// ScrollView scroll down multiline GTK widget, which keep backup session log data,
// to show the most recent line.
func (v *NotifierUI) ScrollView() error {
	adj, err := v.logViewPort.GetVAdjustment()
	if err != nil {
		return err
	}
	adj.SetValue(adj.GetUpper())
	//v.grid.QueueDraw()
	//v.logViewPort.QueueDraw()
	return nil
}

// addColorTags add special format tags to colorize TextView control.
func addColorTags(buffer *gtk.TextBuffer) error {
	table, err := buffer.GetTagTable()
	if err != nil {
		return err
	}

	tag, err := gtk.TextTagNew("BlueColor")
	if err != nil {
		return err
	}
	err = tag.SetProperty("foreground", "Dodger Blue")
	if err != nil {
		return err
	}
	table.Add(tag)

	tag, err = gtk.TextTagNew("RedColor")
	if err != nil {
		return err
	}
	err = tag.SetProperty("foreground", "Red")
	if err != nil {
		return err
	}
	table.Add(tag)

	tag, err = gtk.TextTagNew("AquaColor")
	if err != nil {
		return err
	}
	err = tag.SetProperty("foreground", "Aqua")
	if err != nil {
		return err
	}
	table.Add(tag)

	tag, err = gtk.TextTagNew("YellowColor")
	if err != nil {
		return err
	}
	err = tag.SetProperty("foreground", "Goldenrod")
	if err != nil {
		return err
	}
	table.Add(tag)

	tag, err = gtk.TextTagNew("OrangeRedColor")
	if err != nil {
		return err
	}
	err = tag.SetProperty("foreground", "Orange Red")
	if err != nil {
		return err
	}
	table.Add(tag)

	tag, err = gtk.TextTagNew("Path")
	if err != nil {
		return err
	}
	err = tag.SetProperty("underline", pango.UNDERLINE_SINGLE)
	if err != nil {
		return err
	}
	table.Add(tag)

	return nil
}

// getRuneIndex finds index of UTF-8 character by character byte offset in line string.
func getRuneIndex(line string, byteOffset int) int {
	runeIndex := 0
	var index int
	// var runeValue rune
	for index = range line {
		// lg.Infof("rune=%v, offset=%d", runeValue, index)
		if index == byteOffset {
			return runeIndex
		}
		runeIndex++
	}
	if index+1 == byteOffset {
		return runeIndex
	}
	return -1
}

// lToU gets logger.LogLevel string representation and convert it to upper case.
func lToU(level logger.LogLevel) string {
	return strings.ToUpper(level.ShortStr())
}

// getLogEventsRegex recognize logger.LogLevel entry in backup session log line output
// to colorize it in TextVide GTK widget.
func getLogEventsRegex(events []struct {
	Level   logger.LogLevel
	TagName string
}) *regexp.Regexp {

	var buf bytes.Buffer
	for i, event := range events {
		buf.WriteString(lToU(event.Level))
		if i < len(events)-1 {
			buf.WriteString("|")
		}
	}
	re := regexp.MustCompile(fmt.Sprintf(`\[.+\]\s+(?P<Event>(%s))`, buf.String()))
	return re
}

// addLineToBuffer get next log line received from backup session process
// to process and display this line in application GUI.
func (v *NotifierUI) addLineToBuffer(buffer *gtk.TextBuffer, line string) {
	end := buffer.GetEndIter()
	endOffset := end.GetOffset()
	buffer.Insert(end, line)

	events := []struct {
		Level   logger.LogLevel
		TagName string
	}{
		{logger.InfoLevel, "BlueColor"},
		{logger.NotifyLevel, "YellowColor"},
		{logger.WarnLevel, "OrangeRedColor"},
		{logger.ErrorLevel, "RedColor"},
		{logger.FatalLevel, "RedColor"},
		{logger.PanicLevel, "RedColor"},
	}
	re := getLogEventsRegex(events)
	m := core.FindStringSubmatchIndexes(re, line)
	if a, ok := m["Event"]; ok {
		value := line[a[0]:a[1]]
		p1 := buffer.GetIterAtOffset(getRuneIndex(line, a[0]) + endOffset)
		p2 := buffer.GetIterAtOffset(getRuneIndex(line, a[1]) + endOffset)
		for _, event := range events {
			if value == lToU(event.Level) {
				buffer.ApplyTagByName(event.TagName, p1, p2)
			}
		}
	}
	/*
			var err error
		   	re, err = getSubpathRegexp()
		   	if err != nil {
		   		return err
		   	}
		   	m = core.FindStringSubmatchIndexes(re, line)
		   	if b, ok := m["Path"]; ok {
		   		link, err := gtk.LabelNew("/home/")
		   		if err != nil {
		   			return err
		   		}
		   		// SetAllMargins(link, 0)
		   		css := `
		   * {
		       margin: 0;
		       padding: 0;
		       border-style: none;
		       border-radius: 0;
		       border-width: 0;
		       outline-style: none;
		       outline-offset: 0px;
		   }
		   		`
		   		// end := buffer.GetEndIter()
		   		p1 := buffer.GetIterAtOffset(getRuneIndex(line, b[0]) + endOffset)
		   		// p2 := buffer.GetIterAtOffset(getRuneIndex(line, b[1]) + endOffset)
		   		anchor, err := buffer.CreateChildAnchor(p1)
		   		if err != nil {
		   			return err
		   		}
		   		v.logTextView.AddChildAtAnchor(link, anchor)
		   		link.ShowAll()
		   		applyStyleCSS(&link.Widget, css)
		   		// buffer.ApplyTagByName("Path", p1, p2)
		   	}
	*/
}

// UpdateTextViewLog add log line to the end of
// Session Log GTK widget.
func (v *NotifierUI) UpdateTextViewLog(line string) error {
	call := func() {
		//if v.logTextView != nil {
		buffer, err := v.logTextView.GetBuffer()
		if err != nil {
			lg.Fatal(err)
		}
		v.addLineToBuffer(buffer, line)

		err = v.ScrollView()
		if err != nil {
			lg.Fatal(err)
		}
		//}
	}
	MustIdleAdd(call)
	return nil
}

// UpdateBackupProgress updates visual progress of backup
// with status and percent progresses.
func (v *NotifierUI) UpdateBackupProgress(progress *float32,
	progressStr string, fromAsync bool) error {

	call := func() {
		if progress == nil {
			v.pbm.StartPulse()
			err := v.pbm.AddProgressBarStyleClass("run-animation")
			if err != nil {
				lg.Fatal(err)
			}
		} else {
			prg := float64(*progress)
			err := v.pbm.SetFraction(prg)
			if err != nil {
				lg.Fatal(err)
			}
			if prg == 1 {
				err := v.pbm.RemoveProgressBarStyleClass("run-animation")
				if err != nil {
					lg.Fatal(err)
				}
			}
		}
		v.statusLabel.SetMarkup(progressStr)
	}
	if fromAsync {
		MustIdleAdd(call)
	} else {
		call()
	}
	return nil
}

// BackupCompletionType signify all possible states of backup session completion.
type BackupCompletionType int

// It could be 4 possible exit status:
// 1) backup failed (due to some issue recognized as critical);
// 2) backup terminated by user or by system request;
// 3) backup successfully completed without any issues;
// 4) backup completed, but some data are not backed up
// due to errors happened during backup process;
const (
	BackupFailed BackupCompletionType = iota
	BackupTerminated
	BackupSucessfullyCompleted
	BackupCompletedWithErrors
)

func (v *NotifierUI) decodeBackupCompletionType(err error,
	backupProgress *backup.Progress) BackupCompletionType {
	if err != nil {
		if rsync.IsProcessTerminatedError(err) {
			return BackupTerminated
		} else {
			return BackupFailed
		}
	} else {
		if backupProgress.TotalProgress.Failed != nil {
			return BackupCompletedWithErrors
		} else {
			return BackupSucessfullyCompleted
		}
	}
}

// getDesktopNotificationSummaryAndBody prepares desktop notification subject and body text.
func (v *NotifierUI) getDesktopNotificationSummaryAndBody(completionType BackupCompletionType,
	backupProgress *backup.Progress) (string, string) {

	var summary, body string
	switch completionType {
	case BackupSucessfullyCompleted:
		summary = locale.T(
			MsgDesktopNotificationBackupSuccessfullyCompleted,
			struct{ ProfileName string }{ProfileName: v.profileName})
	case BackupCompletedWithErrors:
		summary = locale.T(
			MsgDesktopNotificationBackupCompletedWithErrors,
			struct{ ProfileName string }{ProfileName: v.profileName})
	case BackupFailed:
		summary = locale.T(
			MsgDesktopNotificationBackupFailed,
			struct{ ProfileName string }{ProfileName: v.profileName})
	case BackupTerminated:
		summary = locale.T(
			MsgDesktopNotificationBackupTerminated,
			struct{ ProfileName string }{ProfileName: v.profileName})
	}

	var buf bytes.Buffer
	if completionType != BackupFailed && completionType != BackupTerminated &&
		backupProgress != nil && backupProgress.TotalProgress != nil {

		if backupProgress.TotalProgress.Completed != nil {
			buf.WriteString(fmt.Sprintln(locale.T(MsgDesktopNotificationTotalSize,
				struct{ TotalSize string }{TotalSize: core.GetReadableSize(
					*backupProgress.TotalProgress.Completed)})))
		}
		if backupProgress.TotalProgress.Failed != nil {
			buf.WriteString(fmt.Sprintln(locale.T(MsgDesktopNotificationFailedToBackupSize,
				struct{ FailedToBackupSize string }{FailedToBackupSize: core.GetReadableSize(
					*backupProgress.TotalProgress.Failed)})))
		}
		if backupProgress.TotalProgress.Skipped != nil {
			buf.WriteString(fmt.Sprintln(locale.T(MsgDesktopNotificationSkippedSize,
				struct{ SkippedSize string }{SkippedSize: core.GetReadableSize(
					*backupProgress.TotalProgress.Skipped)})))
		}
	}
	if backupProgress != nil {
		timeTaken := backupProgress.GetTotalTimeTaken()
		sections := 2
		buf.WriteString(fmt.Sprintln(locale.T(MsgDesktopNotificationTimeTaken,
			struct{ TimeTaken string }{TimeTaken: core.FormatDurationToDaysHoursMinsSecs(
				timeTaken, true, &sections)})))
	}
	body = buf.String()

	return summary, body
}

func (v *NotifierUI) checkDesktopNotificationEnabled() (bool, error) {
	appSettings, err := glib.SettingsNew(SETTINGS_SCHEMA_ID)
	if err != nil {
		return false, err
	}
	enabled := appSettings.GetBoolean(CFG_PERFORM_DESKTOP_NOTIFICATION)
	return enabled, nil
}

func (v *NotifierUI) sendDesktopNotification(completionType BackupCompletionType,
	backupProgress *backup.Progress) error {

	summary, body := v.getDesktopNotificationSummaryAndBody(completionType, backupProgress)
	notif, err := libnotify.NotifyNotificationNew(summary, body, "")
	if err != nil {
		return err
	}
	err = notif.Show()
	if err != nil {
		return err
	}
	return nil
}

func (v *NotifierUI) checkNotificationScriptEnabled() (bool, error) {
	appSettings, err := glib.SettingsNew(SETTINGS_SCHEMA_ID)
	if err != nil {
		return false, err
	}
	enabled := appSettings.GetBoolean(CFG_RUN_NOTIFICATION_SCRIPT)
	return enabled, nil
}

func buildEnvVars(completionType BackupCompletionType,
	backupProgress *backup.Progress) []string {

	var status string
	switch completionType {
	case BackupTerminated:
		status = "terminated"
	case BackupFailed:
		status = "failed"
	case BackupSucessfullyCompleted:
		status = "done"
	case BackupCompletedWithErrors:
		status = "done_with_errors"
	}

	var vars []string
	vars = append(vars, fmt.Sprintf("BACKUP_STATUS=%s", status))
	if backupProgress != nil {
		if backupProgress.TotalProgress.Completed != nil {
			vars = append(vars, fmt.Sprintf("SIZE_BACKEDUP_MB=%d",
				backupProgress.TotalProgress.Completed.GetByteCount()/core.MB))
		}
		if backupProgress.TotalProgress.Failed != nil {
			vars = append(vars, fmt.Sprintf("SIZE_FAILED_MB=%d",
				backupProgress.TotalProgress.Failed.GetByteCount()/core.MB))
		}
		if backupProgress.TotalProgress.Skipped != nil {
			vars = append(vars, fmt.Sprintf("SIZE_SKIPPED_MB=%d",
				backupProgress.TotalProgress.Skipped.GetByteCount()/core.MB))
		}
		timeTaken := backupProgress.GetTotalTimeTaken()
		if timeTaken != time.Duration(0) {
			vars = append(vars, fmt.Sprintf("TIME_TAKEN_SEC=%d", int(timeTaken.Seconds())))
		}
	}
	return vars
}

func (v *NotifierUI) runNotificationScript(completionType BackupCompletionType,
	backupProgress *backup.Progress, scriptPath string) error {

	// get default shell
	shell := os.Getenv("SHELL")
	// once not found fallback to bash
	if shell == "" {
		shell = "/usr/bin/bash"
	}

	_, err := core.RunExecutableWithExtraVars(shell,
		buildEnvVars(completionType, backupProgress), "/etc/gorsync/notification.sh")
	if err != nil {
		return err
	}
	return nil
}

// reportCompletion updates backup process state and progress bar status.
func (v *NotifierUI) ReportCompletion(progress float32, err error,
	backupProgress *backup.Progress, async bool) {

	completionType := v.decodeBackupCompletionType(err, backupProgress)
	var finalMsg string
	switch completionType {
	case BackupTerminated:
		finalMsg = locale.T(MsgAppWindowBackupProgressTerminated, nil)
	case BackupFailed:
		finalMsg = locale.T(MsgAppWindowBackupProgressFailed, nil)
	case BackupCompletedWithErrors:
		finalMsg = locale.T(MsgAppWindowBackupProgressCompletedWithErrors, nil)
	case BackupSucessfullyCompleted:
		finalMsg = locale.T(MsgAppWindowBackupProgressCompleted, nil)
	}

	mp := NewMarkup(MARKUP_SIZE_LARGER, 0, 0, finalMsg, nil)
	err2 := v.UpdateBackupProgress(&progress, mp.String(), async)
	if err2 != nil {
		lg.Fatal(err2)
	}

	go func(completionType BackupCompletionType, backupProgress *backup.Progress) {
		time.Sleep(time.Millisecond * 200)
		MustIdleAdd(func() {
			err := v.ScrollView()
			if err != nil {
				lg.Fatal(err)
			}
		})

		enabled, err := v.checkDesktopNotificationEnabled()
		if err != nil {
			lg.Fatal(err)
		}
		if enabled && completionType != BackupTerminated {
			err = v.sendDesktopNotification(completionType, backupProgress)
			if err != nil {
				lg.Warn(locale.T(MsgAppWindowShowNotificationError,
					struct{ Error error }{Error: err}))
			}
		}
		scriptPath := "/etc/gorsync/notification.sh"
		enabled, err = v.checkNotificationScriptEnabled()
		if err != nil {
			lg.Fatal(err)
		}
		if enabled {
			if stat, err := os.Stat(scriptPath); err == nil {
				mode := stat.Mode()
				// check script is executable for POSIX-kind OS
				if !shell.IsLinuxMacOSFreeBSD() || mode&0111 != 0 {
					err = v.runNotificationScript(completionType,
						backupProgress, scriptPath)
					if err != nil {
						lg.Warn(locale.T(MsgAppWindowRunNotificationScriptError,
							struct{ Error error }{Error: err}))
					}
				} else {
					lg.Warn(locale.T(MsgAppWindowNotificationScriptExecutableError,
						struct{ ScriptPath string }{ScriptPath: scriptPath}))
				}
			} else {
				lg.Warn(locale.T(MsgAppWindowGetExecutableScriptInfoError,
					struct{ Error error }{Error: err}))
			}
		}
		// report about real completion via asynchronous method
		close(v.done)

	}(completionType, backupProgress)

}
