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
	PrevBackups *PrevBackups

	RootDest     string
	BackupFolder string

	// Notify only once (theoretically it never happens)
	SizeChangedNotified bool
}

// StartPlanStage save the start time of 1st stage.
func (this *Progress) StartPlanStage() {
	this.StartPlanTime = time.Now()
}

// FinishPlanStage save the end time of 1st stage.
func (this *Progress) FinishPlanStage() {
	this.EndPlanTime = time.Now()
}

// Save the start time of 2nd stage.
func (this *Progress) StartBackupStage() {
	this.StartBackupTime = time.Now()
}

// Save the end time of 2nd stage.
func (this *Progress) FinishBackupStage() {
	this.EndBackupTime = time.Now()
}

// GetTotalTimeTaken count up total time of backup session execution.
func (this *Progress) GetTotalTimeTaken() time.Duration {
	var timeTaken time.Duration
	now := time.Now()
	if this.EndPlanTime.After(this.StartPlanTime) {
		timeTaken += this.EndPlanTime.Sub(this.StartPlanTime)
	} else if !this.StartPlanTime.IsZero() {
		timeTaken += now.Sub(this.StartPlanTime)
	}
	if this.EndBackupTime.After(this.StartBackupTime) {
		timeTaken += this.EndBackupTime.Sub(this.StartBackupTime)
	} else if !this.StartBackupTime.IsZero() {
		timeTaken += now.Sub(this.StartBackupTime)
	}
	return timeTaken
}

// CalcTimePassedAndETA count total time passed in backup stage (2nd stage)
// and compute ETA (estimated time of arrival) - time left.
func (this *Progress) CalcTimePassedAndETA(plan *Plan) (time.Duration, *time.Duration) {
	timePassed := time.Now().Sub(this.StartBackupTime)
	if this.SizeBackedUp() > 0 {
		totalTime := float32(timePassed) * float32(plan.BackupSize) /
			float32(this.SizeBackedUp())
		eta := time.Duration(totalTime) - timePassed
		// lg.Debugf("Left to backup: %v", this.LeftToBackup())
		// lg.Debugf("Total time: %v", totalTime)
		// lg.Debugf("Time pass: %v", timePassed)
		// lg.Debugf("ETA: %v", time.Duration(totalTime)-timePassed)
		return timePassed, &eta
	}
	return timePassed, nil
}

func (this *Progress) PrevBackupsUsed(prevBackups *PrevBackups) {
	this.PrevBackups = prevBackups
}

// PrintTotalStatistics print results on backup session completion. Print all statistics
// including time taken, volume processed, errors happens and so on.
func (this *Progress) PrintTotalStatistics(lg logger.PackageLog, plan *Plan) error {
	lines, err := this.getTotalStatistics(plan)
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
func (this *Progress) SayGoodbye(lg logger.PackageLog) {
	lg.Info(locale.T(MsgLogBackupStageExitMessage, nil))
}

// SizeBackedUp return total size processed during 2nd stage.
func (this *Progress) SizeBackedUp() core.FolderSize {
	return this.TotalProgress.GetTotal()
}

// LeftToBackup return size left to process in 2nd stage.
func (this *Progress) LeftToBackup(plan *Plan) core.FolderSize {
	var left core.FolderSize
	// small protection in case when original backup size get changed
	if plan.BackupSize >= this.SizeBackedUp() {
		left = plan.BackupSize - this.SizeBackedUp()
	} else {
		if !this.SizeChangedNotified {
			this.Log.Notify(locale.T(MsgLogBackupDetectedTotalBackupSizeGetChanged, nil))
			this.SizeChangedNotified = true
		}
	}
	return left
}

func (this *Progress) SetRootDestination(rootDestPath string) {
	this.RootDest = rootDestPath
}

func (this *Progress) SetBackupFolder(backupFolder string) error {
	this.BackupFolder = backupFolder
	// relocate log files from one destination path to another
	fullPath := this.GetBackupFullPath(backupFolder)
	err := this.LogFiles.ChangeRootPath(fullPath)
	return err
}

func (this *Progress) GetBackupFullPath(backupFolder string) string {
	backupFullPath := filepath.Join(this.RootDest, backupFolder)
	return backupFullPath
}

// EventPlanStage_NodeStructureStartInquiry report about start inquiry of RSYNC source (1st stage).
func (this *Progress) EventPlanStage_NodeStructureStartInquiry(sourceID int,
	sourceRsync string) error {

	this.Log.Info(locale.T(MsgLogPlanStageInquirySource,
		struct {
			SourceID int
			Path     string
		}{SourceID: sourceID + 1, Path: sourceRsync}))

	if this.Notifier != nil {
		err := this.Notifier.NotifyPlanStage_NodeStructureStartInquiry(sourceID, sourceRsync)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventPlanStage_NodeStructureDoneInquiry report about end inquiry of RSYNC source (1st stage).
func (this *Progress) EventPlanStage_NodeStructureDoneInquiry(sourceID int,
	sourceRsync string, dir *core.Dir) error {

	folderCount := dir.GetFoldersCount()
	skipFolderCount := dir.GetFoldersIgnoreCount()
	this.Log.Infof("%s, %s, %s",
		locale.TP(MsgLogPlanStageSourceFolderCountInfo,
			struct{ FolderCount int }{FolderCount: folderCount}, folderCount),
		locale.TP(MsgLogPlanStageSourceSkipFolderCountInfo,
			struct{ SkipFolderCount int }{SkipFolderCount: skipFolderCount}, skipFolderCount),
		locale.T(MsgLogPlanStageSourceTotalSizeInfo,
			struct {
				TotalSize string
			}{TotalSize: core.GetReadableSize(dir.GetTotalSize())}))

	if this.Notifier != nil {
		err := this.Notifier.NotifyPlanStage_NodeStructureDoneInquiry(sourceID,
			sourceRsync, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventBackupStage_FolderStartBackup report about backup folder start (2nd stage).
func (this *Progress) EventBackupStage_FolderStartBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *Plan) error {

	backupFolder := this.GetBackupFullPath(this.BackupFolder)
	path, err := core.GetRelativePath(backupFolder, paths.DestPath)
	if err != nil {
		return err
	}

	timePassed, eta := this.CalcTimePassedAndETA(plan)
	leftToBackup := this.LeftToBackup(plan)

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
		this.Log.Notify(msg)
	} else {
		this.Log.Info(msg)
	}

	if this.Notifier != nil {
		err := this.Notifier.NotifyBackupStage_FolderStartBackup(backupFolder,
			paths, backupType, leftToBackup, timePassed, eta)
		if err != nil {
			return err
		}
	}

	return nil
}

// EventBackupStage_FolderStartBackup report about backup folder end (2nd stage).
func (this *Progress) EventBackupStage_FolderDoneBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *Plan,
	sizeDone core.SizeProgress, sessionErr error) error {

	this.Progress.Add(sizeDone)
	this.TotalProgress.Add(sizeDone)

	timePassed, eta := this.CalcTimePassedAndETA(plan)
	leftToBackup := this.LeftToBackup(plan)

	if this.Notifier != nil {
		backupFolder := this.GetBackupFullPath(this.BackupFolder)
		err := this.Notifier.NotifyBackupStage_FolderDoneBackup(backupFolder,
			paths, backupType, leftToBackup, sizeDone, timePassed, eta, sessionErr)
		if err != nil {
			return err
		}
	}

	return nil
}

// getTotalStatistics prepare multiline report about backup session results.
// Used to report about results in the end of backup process.
func (this *Progress) getTotalStatistics(plan *Plan) ([]string, error) {
	sections := 2
	var b bytes.Buffer
	wli := writeLineIndent
	wli(&b, 0, DoubleSplitLogLine)
	wli(&b, 0, locale.T(MsgLogStatisticsSummaryCaption, nil))
	wli(&b, 1, locale.T(MsgLogStatisticsEnvironmentCaption, nil))
	wli(&b, 2, f("%s %s", core.GetAppFullTitle(), core.GetAppVersion()))
	version, protocol, err := rsync.GetRsyncVersion()
	if err != nil {
		return nil, err
	}
	wli(&b, 2, locale.T(MsgRsyncInfo, struct{ RSYNCDetectedVer, RSYNCDetectedProtocol string }{
		RSYNCDetectedVer: version, RSYNCDetectedProtocol: protocol}))
	wli(&b, 2, locale.T(MsgGolangInfo, struct{ GolangVersion, AppArchitecture string }{
		GolangVersion:   core.GetGolangVersion(),
		AppArchitecture: core.GetAppArchitecture()}))
	wli(&b, 1, locale.T(MsgLogStatisticsResultsCaption, nil))
	wli(&b, 2, locale.T(MsgLogStatisticsStatusCaption, nil))
	if this.TotalProgress.Failed != nil {
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
	timeTaken := this.EndPlanTime.Sub(this.StartPlanTime)
	wli(&b, 3, locale.T(MsgLogStatisticsPlanStageTimeTaken, struct{ TimeTaken string }{
		TimeTaken: core.FormatDurationToDaysHoursMinsSecs(timeTaken, true, &sections)}))
	wli(&b, 2, locale.T(MsgLogStatisticsBackupStageCaption, nil))
	backupFolder := this.GetBackupFullPath(this.BackupFolder)
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageDestinationPath, struct{ Path string }{
		Path: backupFolder}))

	if len(this.PrevBackups.Backups) > 0 && plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(this.RootDest, this.PrevBackups.GetDirPaths())
		if err != nil {
			return nil, err
		}
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStagePreviousBackupFound, struct{ Path string }{
			Path: this.RootDest}))
		for _, path := range paths {
			wli(&b, 4, path)
		}
	} else if len(this.PrevBackups.Backups) > 0 && !plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(this.RootDest, this.PrevBackups.GetDirPaths())
		if err != nil {
			return nil, err
		}
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStagePreviousBackupFoundButDisabled, struct{ Path string }{
			Path: this.RootDest}))
		for _, path := range paths {
			wli(&b, 4, path)
		}
	} else {
		wli(&b, 3, locale.T(MsgLogStatisticsBackupStageNoValidPreviousBackupFound, nil))
	}

	var size core.FolderSize
	if this.TotalProgress.Completed != nil {
		size = *this.TotalProgress.Completed
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageTotalSize, struct{ TotalSize string }{
		TotalSize: core.GetReadableSize(size)}))
	size = 0
	if this.TotalProgress.Skipped != nil {
		size = *this.TotalProgress.Skipped
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageSkippedSize, struct{ SkippedSize string }{
		SkippedSize: core.GetReadableSize(size)}))
	size = 0
	if this.TotalProgress.Failed != nil {
		size = *this.TotalProgress.Failed
	}
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageFailedToBackupSize, struct{ FailedToBackupSize string }{
		FailedToBackupSize: core.GetReadableSize(size)}))
	timeTaken = this.EndBackupTime.Sub(this.StartBackupTime)
	wli(&b, 3, locale.T(MsgLogStatisticsBackupStageTimeTaken, struct{ TimeTaken string }{
		TimeTaken: core.FormatDurationToDaysHoursMinsSecs(timeTaken, true, &sections)}))
	wli(&b, 0, DoubleSplitLogLine)
	return splitToLines(&b)
}

// Close release any resources occupied.
func (this *Progress) Close() error {
	if this.LogFiles != nil {
		return this.LogFiles.Close()
	}
	return nil
}
