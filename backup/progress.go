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

package backup

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
)

// Progress tracks all aspects of backup session progress
// including time points, volume processed and so on.
// Progress employ Notifier interface to report progress
// to plug-in interface.
type Progress struct {
	Context       context.Context
	LogFiles      *LogFiles
	Log           logger.PackageLog
	RsyncLog      *rsync.Logging
	Progress      *core.SizeProgress
	TotalProgress *core.SizeProgress

	// Notifier interface used to plug external objects to notify
	Notifier Notifier

	// Time stamps in 1st stage
	StartPlanTime time.Time
	EndPlanTime   time.Time

	// Time stamps in 2nd stage
	StartBackupTime time.Time
	EndBackupTime   time.Time

	// Previous backup sessions found to use for deduplicaton
	PreviousBackups *PreviousBackups

	RootDest     string
	BackupFolder string

	// Notify only once (theoretically it never happens)
	SizeChangedNotified bool
}

// StartPlanStage save the start time of 1st stage.
func (v *Progress) StartPlanStage() {
	v.StartPlanTime = time.Now()
}

// FinishPlanStage save the end time of 1st stage.
func (v *Progress) FinishPlanStage() {
	v.EndPlanTime = time.Now()
}

// StartBackupStage save the start time of 2nd stage.
func (v *Progress) StartBackupStage() {
	v.StartBackupTime = time.Now()
}

// FinishBackupStage save the end time of 2nd stage.
func (v *Progress) FinishBackupStage() {
	v.EndBackupTime = time.Now()
}

// GetTotalTimeTaken count up total time of backup session execution.
func (v *Progress) GetTotalTimeTaken() time.Duration {
	var timeTaken time.Duration
	now := time.Now()
	if v.EndPlanTime.After(v.StartPlanTime) {
		timeTaken += v.EndPlanTime.Sub(v.StartPlanTime)
	} else if !v.StartPlanTime.IsZero() {
		timeTaken += now.Sub(v.StartPlanTime)
	}
	if v.EndBackupTime.After(v.StartBackupTime) {
		timeTaken += v.EndBackupTime.Sub(v.StartBackupTime)
	} else if !v.StartBackupTime.IsZero() {
		timeTaken += now.Sub(v.StartBackupTime)
	}
	return timeTaken
}

// CalcTimePassedAndETA count total time passed in backup stage (2nd stage)
// and compute ETA (estimated time of arrival) - time left.
func (v *Progress) CalcTimePassedAndETA(plan *Plan) (time.Duration, *time.Duration) {
	// timePassed := time.Now().Sub(v.StartBackupTime)
	timePassed := time.Since(v.StartBackupTime)
	if v.SizeBackedUp() > 0 {
		totalTime := float32(timePassed) * float32(plan.BackupSize) /
			float32(v.SizeBackedUp())
		eta := time.Duration(totalTime) - timePassed
		// lg.Debugf("Left to backup: %v", v.LeftToBackup())
		// lg.Debugf("Total time: %v", totalTime)
		// lg.Debugf("Time pass: %v", timePassed)
		// lg.Debugf("ETA: %v", time.Duration(totalTime)-timePassed)
		return timePassed, &eta
	}
	return timePassed, nil
}

// PrintTotalStatistics print results on backup session completion. Print all statistics
// including time taken, volume processed, errors happens and so on.
func (v *Progress) PrintTotalStatistics(lg logger.PackageLog, plan *Plan) error {
	lines, err := v.getTotalStatistics(plan)
	if err != nil {
		lg.Error(err)
		return err
	}
	for _, line := range lines {
		lg.Info(line)
	}
	return nil
}

// SayGoodbye report completion to the backup session log.
func (v *Progress) SayGoodbye(lg logger.PackageLog) {
	lg.Info(locale.T(MsgLogBackupStageExitMessage, nil))
}

// SizeBackedUp return total size processed during 2nd stage.
func (v *Progress) SizeBackedUp() core.FolderSize {
	return v.TotalProgress.GetTotal()
}

// LeftToBackup return size left to process in 2nd stage.
func (v *Progress) LeftToBackup(plan *Plan) core.FolderSize {
	var left core.FolderSize
	// small protection in case when original backup size get changed
	if plan.BackupSize >= v.SizeBackedUp() {
		left = plan.BackupSize - v.SizeBackedUp()
	} else {
		if !v.SizeChangedNotified {
			v.Log.Notify(locale.T(MsgLogBackupDetectedTotalBackupSizeGetChanged, nil))
			v.SizeChangedNotified = true
		}
	}
	return left
}

// PreviousBackupsUsed save previous backup sessions found for deduplication to activate.
func (v *Progress) PreviousBackupsUsed(prevBackups *PreviousBackups) {
	v.PreviousBackups = prevBackups
}

// SetRootDestination set absolute destination path,
// where backup session will create it new subfolder and store data.
func (v *Progress) SetRootDestination(rootDestPath string) {
	v.RootDest = rootDestPath
}

// SetBackupFolder set newly created backup session subfolder,
// where copied data and logs will be stored.
func (v *Progress) SetBackupFolder(backupFolder string) error {
	v.BackupFolder = backupFolder
	// relocate log files from one destination path to another
	fullPath := v.GetBackupFullPath(backupFolder)
	err := v.LogFiles.ChangeRootPath(fullPath)
	return err
}

// GetBackupFullPath return absolute destination path with backup
// session subfolder concatenated.
func (v *Progress) GetBackupFullPath(backupFolder string) string {
	backupFullPath := filepath.Join(v.RootDest, backupFolder)
	return backupFullPath
}

// EventPlanStage_NodeStructureStartInquiry report about start inquiry of RSYNC source (1st stage).
func (v *Progress) EventPlanStage_NodeStructureStartInquiry(sourceID int,
	sourceRsync string) error {

	v.Log.Info(locale.T(MsgLogPlanStageInquirySource,
		struct {
			SourceID int
			Path     string
		}{SourceID: sourceID + 1, Path: sourceRsync}))

	if v.Notifier != nil {
		err := v.Notifier.NotifyPlanStage_NodeStructureStartInquiry(sourceID, sourceRsync)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventPlanStage_NodeStructureDoneInquiry report about end inquiry of RSYNC source (1st stage).
func (v *Progress) EventPlanStage_NodeStructureDoneInquiry(sourceID int,
	sourceRsync string, dir *core.Dir) error {

	folderCount := dir.GetFoldersCount()
	skipFolderCount := dir.GetFoldersIgnoreCount()
	v.Log.Infof("%s, %s, %s",
		locale.TP(MsgLogPlanStageSourceFolderCountInfo,
			struct{ FolderCount int }{FolderCount: folderCount}, folderCount),
		locale.TP(MsgLogPlanStageSourceSkipFolderCountInfo,
			struct{ SkipFolderCount int }{SkipFolderCount: skipFolderCount}, skipFolderCount),
		locale.T(MsgLogPlanStageSourceTotalSizeInfo,
			struct {
				TotalSize string
			}{TotalSize: core.GetReadableSize(dir.GetTotalSize())}))

	if v.Notifier != nil {
		err := v.Notifier.NotifyPlanStage_NodeStructureDoneInquiry(sourceID,
			sourceRsync, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventBackupStage_FolderStartBackup report about backup folder start (2nd stage).
func (v *Progress) EventBackupStage_FolderStartBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *Plan) error {

	backupFolder := v.GetBackupFullPath(v.BackupFolder)
	path, err := core.GetRelativePath(backupFolder, paths.DestPath)
	if err != nil {
		return err
	}

	timePassed, eta := v.CalcTimePassedAndETA(plan)
	leftToBackup := v.LeftToBackup(plan)

	etaStr := "*"
	if eta != nil {
		sections := 2
		etaStr = core.FormatDurationToDaysHoursMinsSecs(*eta, true, &sections)
	}

	msg := locale.T(MsgLogBackupStageProgressBackupSuccess,
		struct{ SizeLeft, TimeLeft, BackupAction, FolderPath string }{
			SizeLeft: core.GetReadableSize(leftToBackup), TimeLeft: etaStr,
			BackupAction: GetBackupTypeDescription(backupType),
			FolderPath:   path})

	if backupType == core.FBT_SKIP {
		v.Log.Notify(msg)
	} else {
		v.Log.Info(msg)
	}

	if v.Notifier != nil {
		err := v.Notifier.NotifyBackupStage_FolderStartBackup(backupFolder,
			paths, backupType, leftToBackup, timePassed, eta)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventBackupStage_FolderDoneBackup report about backup folder end (2nd stage).
func (v *Progress) EventBackupStage_FolderDoneBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *Plan,
	sizeDone core.SizeProgress, sessionErr error) error {

	v.Progress.Add(sizeDone)
	v.TotalProgress.Add(sizeDone)

	timePassed, eta := v.CalcTimePassedAndETA(plan)
	leftToBackup := v.LeftToBackup(plan)

	if v.Notifier != nil {
		backupFolder := v.GetBackupFullPath(v.BackupFolder)
		err := v.Notifier.NotifyBackupStage_FolderDoneBackup(backupFolder,
			paths, backupType, leftToBackup, sizeDone, timePassed, eta, sessionErr)
		if err != nil {
			return err
		}
	}

	return nil
}

// getTotalStatistics prepare multiline report about backup session results.
// Used to report about results in the end of backup process.
func (v *Progress) getTotalStatistics(plan *Plan) ([]string, error) {
	sections := 2
	var b bytes.Buffer
	wli := writeLineIndent
	wli(&b, 0, DoubleSplitLogLine)
	wli(&b, 0, locale.T(MsgLogStatisticsSummaryCaption, nil))
	wli(&b, 1, locale.T(MsgLogStatisticsEnvironmentCaption, nil))
	wli(&b, 2, f("%s %s", core.GetAppFullTitle(), core.GetAppVersion()))
	version, protocol, err := rsync.GetRsyncVersion()
	if err != nil {
		if rsync.IsExtractVersionAndProtocolError(err) {
			version = "?"
			protocol = version
		} else {
			return nil, err
		}
	}
	wli(&b, 2, locale.T(MsgRsyncInfo, struct{ RSYNCDetectedVer, RSYNCDetectedProtocol string }{
		RSYNCDetectedVer: version, RSYNCDetectedProtocol: protocol}))
	wli(&b, 2, locale.T(MsgGolangInfo, struct{ GolangVersion, AppArchitecture string }{
		GolangVersion:   core.GetGolangVersion(),
		AppArchitecture: core.GetAppArchitecture()}))
	wli(&b, 1, locale.T(MsgLogStatisticsResultsCaption, nil))
	wli(&b, 2, locale.T(MsgLogStatisticsStatusCaption, nil))
	if v.TotalProgress.Failed != nil {
		wli(&b, 3, locale.T(MsgLogStatisticsStatusCompletedWithErrors, nil))
	} else {
		wli(&b, 3, locale.T(MsgLogStatisticsStatusSuccessfullyCompleted, nil))
	}
	wli(&b, 2, locale.T(MsgLogStatisticsPlanStageCaption, nil))
	for i, node := range plan.Nodes {
		wli(&b, 3, locale.T(MsgLogStatisticsPlanStageSourceToBackup,
			struct {
				SeqID       int
				RsyncSource string
			}{
				SeqID: i + 1, RsyncSource: node.Module.SourceRsync}))
	}
	wli(&b, 3, locale.T(MsgLogStatisticsPlanStageTotalSize, struct{ TotalSize string }{
		TotalSize: core.GetReadableSize(plan.BackupSize)}))
	var foldersCount int
	for _, node := range plan.Nodes {
		foldersCount += node.RootDir.GetFoldersCount()
	}
	wli(&b, 3, locale.T(MsgLogStatisticsPlanStageFolderCount, struct{ FolderCount int }{
		FolderCount: foldersCount}))
	var foldersIgnoreCount int
	for _, node := range plan.Nodes {
		foldersIgnoreCount += node.RootDir.GetFoldersIgnoreCount()
	}
	wli(&b, 3, locale.T(MsgLogStatisticsPlanStageFolderSkipCount, struct{ FolderCount int }{
		FolderCount: foldersIgnoreCount}))
	timeTaken := v.EndPlanTime.Sub(v.StartPlanTime)
	wli(&b, 3, locale.T(MsgLogStatisticsPlanStageTimeTaken, struct{ TimeTaken string }{
		TimeTaken: core.FormatDurationToDaysHoursMinsSecs(timeTaken, true, &sections)}))
	wli(&b, 2, locale.T(MsgLogStatisticsBackupStageCaption, nil))
	backupFolder := v.GetBackupFullPath(v.BackupFolder)
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageDestinationPath, struct{ Path string }{
		Path: backupFolder}))

	if len(v.PreviousBackups.Backups) > 0 && plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(v.RootDest, v.PreviousBackups.GetDirPaths())
		if err != nil {
			return nil, err
		}
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStagePreviousBackupFound, struct{ Path string }{
			Path: v.RootDest}))
		for _, path := range paths {
			wli(&b, 4, path)
		}
	} else if len(v.PreviousBackups.Backups) > 0 && !plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(v.RootDest, v.PreviousBackups.GetDirPaths())
		if err != nil {
			return nil, err
		}
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStagePreviousBackupFoundButDisabled, struct{ Path string }{
			Path: v.RootDest}))
		for _, path := range paths {
			wli(&b, 4, path)
		}
	} else {
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStageNoValidPreviousBackupFound, nil))
	}

	var size core.FolderSize
	if v.TotalProgress.Completed != nil {
		size = *v.TotalProgress.Completed
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageTotalSize, struct{ TotalSize string }{
		TotalSize: core.GetReadableSize(size)}))
	size = 0
	if v.TotalProgress.Skipped != nil {
		size = *v.TotalProgress.Skipped
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageSkippedSize, struct{ SkippedSize string }{
		SkippedSize: core.GetReadableSize(size)}))
	size = 0
	if v.TotalProgress.Failed != nil {
		size = *v.TotalProgress.Failed
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageFailedToBackupSize, struct{ FailedToBackupSize string }{
		FailedToBackupSize: core.GetReadableSize(size)}))
	timeTaken = v.EndBackupTime.Sub(v.StartBackupTime)
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageTimeTaken, struct{ TimeTaken string }{
		TimeTaken: core.FormatDurationToDaysHoursMinsSecs(timeTaken, true, &sections)}))
	wli(&b, 0, DoubleSplitLogLine)
	return splitToLines(&b)
}

// Close release any resources occupied.
func (v *Progress) Close() error {
	if v.LogFiles != nil {
		return v.LogFiles.Close()
	}
	return nil
}
