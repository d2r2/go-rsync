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

type Progress struct {
	Context  context.Context
	LogFiles *LogFiles
	Log      logger.PackageLog
	// LogBuffer      bytes.Buffer
	RsyncLog *rsync.Logging
	// RsyncLogBuffer bytes.Buffer
	Progress      *core.SizeProgress
	TotalProgress *core.SizeProgress
	// GlobalErrorCount int
	Notifier Notifier

	StartPlanTime   time.Time
	EndPlanTime     time.Time
	StartBackupTime time.Time
	EndBackupTime   time.Time
	PrevBackups     *PrevBackups

	RootDest     string
	BackupFolder string

	// Notify only once
	SizeChangedNotified bool
}

func (this *Progress) StartPlanStage() {
	this.StartPlanTime = time.Now()
}

func (this *Progress) FinishPlanStage() {
	this.EndPlanTime = time.Now()
}

func (this *Progress) StartBackupStage() {
	this.StartBackupTime = time.Now()
}

func (this *Progress) FinishBackupStage() {
	this.EndBackupTime = time.Now()
}

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

func (this *Progress) CalcTimePassedAndETA(plan *BackupPlan) (time.Duration, *time.Duration) {
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

func (this *Progress) PrintTotalStatistics(lg logger.PackageLog, plan *BackupPlan) error {
	lines, err := this.GetTotalStatistics(plan)
	if err != nil {
		lg.Error(err)
		return err
	}
	for _, line := range lines {
		lg.Info(line)
	}
	return nil
}

func (this *Progress) SayGoodbye(lg logger.PackageLog) {
	lg.Info(locale.T(MsgLogBackupStageExitMessage, nil))
}

func (this *Progress) SizeBackedUp() core.FolderSize {
	return this.TotalProgress.GetTotal()
}

func (this *Progress) LeftToBackup(plan *BackupPlan) core.FolderSize {
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
	path := this.GetBackupFullPath(backupFolder)
	return this.LogFiles.ChangeRootPath(path)
}

func (this *Progress) GetBackupFullPath(backupFolder string) string {
	backupFullPath := filepath.Join(this.RootDest, backupFolder)
	return backupFullPath
}

func (this *Progress) EventPlanStage_NodeStructureStartInquiry(sourceId int,
	sourceRsync string) error {

	this.Log.Info(locale.T(MsgLogPlanStageInquirySource,
		struct {
			SourceID int
			Path     string
		}{SourceID: sourceId + 1, Path: sourceRsync}))

	if this.Notifier != nil {
		err := this.Notifier.NotifyPlanStage_NodeStructureStartInquiry(sourceId, sourceRsync)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Progress) EventPlanStage_NodeStructureDoneInquiry(sourceId int,
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
		err := this.Notifier.NotifyPlanStage_NodeStructureDoneInquiry(sourceId,
			sourceRsync, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

/*
func (this *Progress) EventPlanStage_FolderStartInquiry(path core.SrcDstPath,
	plan *BackupPlan) error {
	return nil
}

func (this *Progress) EventPlanStage_FolderDoneInquiry(path core.SrcDstPath,
	plan *BackupPlan, folderSize io.FolderSize, skipFlag bool) error {
	return nil
}
*/

func (this *Progress) EventBackupStage_FolderStartBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *BackupPlan) error {

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

func (this *Progress) EventBackupStage_FolderDoneBackup(paths core.SrcDstPath,
	backupType core.FolderBackupType, plan *BackupPlan,
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

func (this *Progress) GetTotalStatistics(plan *BackupPlan) ([]string, error) {
	sections := 2
	var b bytes.Buffer
	wli := writeLineIndent
	wli(&b, 0, DoubleSplitLine)
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
				SeqID: i + 1, RsyncSource: node.BackupNode.SourceRsync}))
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
	wli(&b, 0, DoubleSplitLine)
	return splitToLines(&b)
}

func (this *Progress) Close() error {
	if this.LogFiles != nil {
		return this.LogFiles.Close()
	}
	return nil
}

// func (this *Progress) WriteLog(log bytes.Buffer) error {
// 	backupFolder := this.GetBackupFullPath()
// 	fileName := GetLogFileName()
// 	err := writeLog(log, backupFolder, fileName)
// 	if err != nil {
// 		this.Log.Warn(f("Error writing file %q: %v", fileName, err))
// 	}
// 	return err
// }

// func (this *Progress) WriteRsyncLog(log bytes.Buffer) error {
// 	backupFolder := this.GetBackupFullPath()
// 	fileName := GetRsyncLogFileName()
// 	err := writeLog(log, backupFolder, fileName)
// 	if err != nil {
// 		this.Log.Warn(f("Error writing file %q: %v", fileName, err))
// 	}
// 	return err
// }
