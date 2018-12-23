package backup

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
)

var (
	DoubleSplitLine string = strings.Repeat("=", 100)
	SingleSplitLine string = strings.Repeat("-", 100)
)

func BuildBackupPlan(ctx context.Context, lg logger.PackageLog, config *Config,
	notifier Notifier) (*BackupPlan, *Progress, error) {

	progress := &Progress{Context: ctx, Notifier: notifier}

	progress.LogFiles = NewLogFiles()

	// create main log file
	log := core.NewProxyLog(lg, "backup", 6, "2006-01-02T15:04:05",
		func(line string) error {
			writer, err := progress.LogFiles.GetAppendFile(GetLogFileName())
			if err != nil {
				return err
			}
			// ignore error
			io.WriteString(writer, line)
			return nil
		}, logger.InfoLevel)
	progress.Log = log

	// create specific RSYNC log file
	rsyncLog := config.getRsyncSettings()
	if rsyncLog.EnableLog {
		log = core.NewProxyLog(nil, "rsync", 5, "2006-01-02T15:04:05",
			func(line string) error {
				writer, err := progress.LogFiles.GetAppendFile(GetRsyncLogFileName())
				if err != nil {
					return err
				}
				// ignore error
				io.WriteString(writer, line)
				return nil
			}, logger.InfoLevel)
		rsyncLog.Log = log
		progress.RsyncLog = rsyncLog
	}

	progress.StartPlanStage()

	progress.Log.Info(DoubleSplitLine)
	progress.Log.Info(locale.T(MsgLogPlanStageStarting, nil))
	progress.Log.Info(locale.T(MsgLogPlanStageStartTime,
		struct{ Time string }{Time: progress.StartPlanTime.Format("2006 Jan 2 15:04:05")}))

	list := []BackupNodePlan{}
	var totalBackupSize core.FolderSize
	progress.Log.Info(locale.TP(MsgLogPlanStartIterateViaNSources,
		struct{ SourceCount int }{SourceCount: len(config.BackupNodes)},
		len(config.BackupNodes)))

	for i, node := range config.BackupNodes {
		progress.Log.Info(SingleSplitLine)
		err := progress.EventPlanStage_NodeStructureStartInquiry(i, node.SourceRsync)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}

		dr, backupSize, err := estimateNode(ctx, node, progress, config)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}
		if backupSize != nil {
			totalBackupSize += *backupSize
		}

		err = progress.EventPlanStage_NodeStructureDoneInquiry(i, node.SourceRsync, dr)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}

		plan := BackupNodePlan{BackupNode: node, RootDir: dr}
		list = append(list, plan)
	}
	progress.Log.Info(SingleSplitLine)
	progress.FinishPlanStage()
	//	progress.Log.Debugf("Plan: %+v", list)
	progress.Log.Info(locale.T(MsgLogPlanStageEndTime,
		struct{ Time string }{Time: progress.EndPlanTime.Format("2006 Jan 2 15:04:05")}))
	backup := &BackupPlan{Config: config, Nodes: list, BackupSize: totalBackupSize}
	//progress.Log.Debugf("BackupPlan: %+v", backup)
	return backup, progress, nil
}

func estimateNode(ctx context.Context, node BackupNode, progress *Progress, config *Config) (*core.Dir, *core.FolderSize, error) {
	tempDir, err := ioutil.TempDir("", "backup_dir_tree_")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(tempDir)

	progress.Log.Info(locale.T(MsgLogPlanStageUseTemporaryFolder,
		struct{ Path string }{Path: tempDir}))

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(node.SourceRsync, ""),
		DestPath:        filepath.Join(tempDir, node.DestSubPath),
	}

	err = os.MkdirAll(paths.DestPath, 0777)
	if err != nil {
		err = errors.New(f("%s: %v", locale.T(MsgLogPlanStageUseTemporaryFolder,
			struct{ Path string }{Path: tempDir}), err))
		return nil, nil, err
	}

	// RSYNC settings to copy only folder's structure and some specific files
	options := rsync.NewOptions(rsync.WithDefaultParams("--recursive")).
		AddParams(f("--include=%s", "*"+"/")).
		AddParams(f("--include=%s", config.SigFileIgnoreBackup)).
		AddParams(f("--exclude=%s", "*")).
		SetRetryCount(config.RsyncRetryCount)
	sessionErr, _, _ := rsync.RunRsyncWithRetry(ctx, options, progress.RsyncLog, nil, paths)
	if sessionErr != nil {
		return nil, nil, sessionErr
	}
	dir, err := core.BuildDirTree(paths, config.SigFileIgnoreBackup)
	if err != nil {
		return nil, nil, err
	}

	progress.Log.Debug("---------------------------------")
	progress.Log.Debug("Start heuristic search")
	progress.Log.Debug("---------------------------------")

	blockSize := config.getBackupBlockSizeSettings()
	count, err := MeasureDir(ctx, dir, config.RsyncRetryCount, progress.RsyncLog, blockSize)
	if err != nil {
		return nil, nil, err
	}
	progress.Log.Debugf("Total \"full size\" cycle factor %v, full backup %v, content backup %v", count,
		core.GetReadableSize(dir.GetFullBackupSize()),
		core.GetReadableSize(dir.GetContentBackupSize()))
	progress.Log.Debug("---------------------------------")
	progress.Log.Debug("End heuristic search")
	progress.Log.Debug("---------------------------------")
	backupSize2 := dir.GetTotalSize()

	return dir, &backupSize2, nil
}

func (this *BackupPlan) RunBackup(progress *Progress, destPath string, errorHook rsync.ErrorHook) error {

	// Execute backup stage
	err := runBackup(this, progress, destPath, errorHook)
	if err != nil {
		progress.Log.Error(locale.T(MsgLogBackupStageCriticalError,
			struct{ Error error }{Error: err}))
	}

	// Next lines should be executed even if backup failed and err variable is not empty,
	// to store log files to backup folder.

	if progress.RsyncLog != nil {
		// _ = progress.WriteRsyncLog(progress.RsyncLogBuffer)
		rsyncLogFileName := path.Join(progress.GetBackupFullPath(progress.BackupFolder), GetRsyncLogFileName())
		progress.Log.Info(locale.T(MsgLogBackupStageSaveRsyncExtraLogTo,
			struct{ Path string }{Path: rsyncLogFileName}))
	}

	logFileName := path.Join(progress.GetBackupFullPath(progress.BackupFolder), GetLogFileName())
	progress.Log.Info(locale.T(MsgLogBackupStageSaveLogTo,
		struct{ Path string }{Path: logFileName}))
	// _ = progress.WriteLog(progress.LogBuffer)

	progress.SayGoodbye(progress.Log)

	return err
}

func runBackup(plan *BackupPlan, progress *Progress, destPath string, errorHook rsync.ErrorHook) error {

	//var backupType io.BackupType = io.BT_DIFF

	progress.TotalProgress = &core.SizeProgress{}
	progress.Progress = &core.SizeProgress{}
	progress.StartBackupStage()

	progress.Log.Info(DoubleSplitLine)
	progress.Log.Info(locale.T(MsgLogBackupStageStarting, nil))
	progress.Log.Info(locale.T(MsgLogBackupStageStartTime,
		struct{ Time string }{Time: progress.StartBackupTime.Format("2006 Jan 2 15:04:05")}))

	err := createDirAll(destPath)
	if err != nil {
		return err
	}
	progress.SetRootDestination(destPath)

	backupFolder := GetBackupFolderName(true, &progress.StartBackupTime)
	path := progress.GetBackupFullPath(backupFolder)
	err = createDirAll(path)
	if err != nil {
		return err
	}
	err = progress.SetBackupFolder(backupFolder)
	if err != nil {
		return err
	}
	destPath2 := progress.GetBackupFullPath(progress.BackupFolder)
	progress.Log.Info(locale.T(MsgLogBackupStageBackupToDestination,
		struct{ Path string }{Path: destPath2}))

	progress.Log.Info(locale.T(MsgLogBackupStageDiscoveringPreviousBackups, nil))
	prevBackups, err := FindPrevBackupPathsByNodeSignatures(progress.Log, destPath,
		GetNodeSignatures(plan.Config), plan.Config.numberOfPreviousBackupToUse())
	if err != nil {
		return err
	}
	LocalLog.Debugf("End searching for previous backups")
	progress.PrevBackupsUsed(prevBackups)
	if len(prevBackups.Backups) > 0 && plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(destPath, prevBackups.GetDirPaths())
		if err != nil {
			return err
		}
		progress.Log.Info(locale.T(MsgLogBackupStagePreviousBackupFoundAndWillBeUsed,
			struct{ Path string }{Path: destPath}))
		for _, path := range paths {
			progress.Log.Info(string(TAB_RUNE) + path)
		}
	} else if len(prevBackups.Backups) > 0 && !plan.Config.usePreviousBackupEnabled() {
		paths, err := core.GetRelativePaths(destPath, prevBackups.GetDirPaths())
		if err != nil {
			return err
		}
		progress.Log.Info(locale.T(MsgLogBackupStagePreviousBackupFoundButDisabled,
			struct{ Path string }{Path: destPath}))
		for _, path := range paths {
			progress.Log.Info(string(TAB_RUNE) + path)
		}
	} else {
		progress.Log.Notify(locale.T(MsgLogBackupStagePreviousBackupNotFound, nil))
	}

	for i, node := range plan.Nodes {
		progress.Log.Info(SingleSplitLine)
		progress.Log.Info(locale.T(MsgLogBackupStageStartToBackupFromSource,
			struct {
				SeqID       int
				RsyncSource string
			}{SeqID: i + 1, RsyncSource: node.BackupNode.SourceRsync}))

		err := runBackupNode(plan, node, plan.Config.SigFileIgnoreBackup,
			plan.Config.RsyncRetryCount, plan.Config.MaxBackupBlockSizeMb,
			progress, destPath2, errorHook, prevBackups)
		if err != nil {
			return err
		}
	}
	progress.Log.Info(SingleSplitLine)
	newBackupFolder := GetBackupFolderName(false, &progress.StartBackupTime)
	destPath3 := progress.GetBackupFullPath(newBackupFolder)
	err = os.Rename(destPath2, destPath3)
	if err != nil {
		return err
	}
	err = progress.SetBackupFolder(newBackupFolder)
	if err != nil {
		return err
	}

	LocalLog.Debugf("BACKUP FINAL: total progress %+v", progress.TotalProgress)
	LocalLog.Debugf("BACKUP FINAL: left to backup %+v", progress.LeftToBackup(plan))

	progress.Log.Info(locale.T(MsgLogBackupStageRenameDestination,
		struct{ Path string }{Path: destPath3}))

	err = CreateMetadataSignatureFile(plan.Config, destPath3)
	if err != nil {
		return err
	}

	progress.FinishBackupStage()
	progress.Log.Info(locale.T(MsgLogBackupStageEndTime,
		struct{ Time string }{Time: progress.EndBackupTime.Format("2006 Jan 2 15:04:05")}))

	err = progress.PrintTotalStatistics(progress.Log, plan)
	if err != nil {
		return err
	}

	return nil
}

func runBackupNode(plan *BackupPlan, nodePlan BackupNodePlan, ignoreBackupSigFileName string,
	retryCount *int, rsyncMaxBlockSizeMb *int, progress *Progress, destRootPath string,
	errorHook rsync.ErrorHook, prevBackups *PrevBackups) error {

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(nodePlan.BackupNode.SourceRsync, ""),
		DestPath:        filepath.Join(destRootPath, nodePlan.BackupNode.DestSubPath),
	}

	progress.Progress = &core.SizeProgress{}
	err := backupDir(nodePlan.RootDir, ignoreBackupSigFileName, retryCount, rsyncMaxBlockSizeMb,
		plan, progress, paths, errorHook, prevBackups.GetDirPaths())
	return err
}

// func suppressRsyncError(err *rsync.ErrorSpec) *rsync.ErrorSpec {
// 	if err != nil {
// 		if err.ErrorCode == 23 {
// 			return nil
// 		} else {
// 			return err
// 		}
// 	}
// 	return nil
// }

// func IsCriticalError(err *rsync.ErrorSpec) bool {
// 	if err != nil {
// 		if err.Error == rsync.ErrRsyncProcessTerminated {
// 			return true
// 		}
// 	}
// 	return false
// }

func formatError(sessionErr error, skipped bool, rootDest string,
	paths core.SrcDstPath, dirSize core.FolderSize) (string, error) {

	destPath, err := core.GetRelativePath(rootDest, paths.DestPath)
	if err != nil {
		return "", err
	}

	if skipped {
		str := locale.T(MsgLogBackupStageProgressSkipBackupError,
			struct {
				Error                         error
				Size, RsyncSource, FolderPath string
			}{
				Error: sessionErr, Size: core.GetReadableSize(dirSize),
				RsyncSource: paths.RsyncSourcePath, FolderPath: destPath})
		return str, nil
	}
	str := locale.T(MsgLogBackupStageProgressBackupError,
		struct {
			Error                         error
			Size, RsyncSource, FolderPath string
		}{
			Error: sessionErr, Size: core.GetReadableSize(dirSize),
			RsyncSource: paths.RsyncSourcePath, FolderPath: destPath})
	return str, nil
}

func reportProgress(sessionErr, retryErr error, size core.FolderSize,
	plan *BackupPlan, progress *Progress, paths core.SrcDstPath,
	backupType core.FolderBackupType, skipped bool) error {

	if retryErr != nil {
		progress.Log.Info(locale.T(MsgLogBackupStageRecoveredFromError,
			struct{ Error error }{Error: retryErr}))
	}

	if sessionErr != nil {
		str, err := formatError(sessionErr, skipped,
			progress.RootDest, paths, size)
		if err != nil {
			return err
		}
		progress.Log.Warn(str)
		err = progress.EventBackupStage_FolderDoneBackup(paths, backupType, plan,
			core.NewProgressFailed(size), sessionErr)
		if err != nil {
			return err
		}
	} else {
		var sizeProgress core.SizeProgress
		if skipped {
			sizeProgress = core.NewProgressSkipped(size)
		} else {
			sizeProgress = core.NewProgressCompleted(size)
		}
		err := progress.EventBackupStage_FolderDoneBackup(paths, backupType, plan,
			sizeProgress, nil)
		if err != nil {
			return err
		}
	}
	LocalLog.Debugf("TotalProgress = %v, Progress = %v", progress.TotalProgress, progress.Progress)
	//LocalLog.Debugf("BACKUP: skipped size: %v", size)
	return nil
}

func backupDir(dir *core.Dir, ignoreBackupSigFileName string,
	retryCount *int, rsyncMaxBlockSizeMb *int,
	plan *BackupPlan, progress *Progress,
	paths core.SrcDstPath, errorHook rsync.ErrorHook,
	prevBackupPaths []string) error {

	var err error
	var backupType core.FolderBackupType

	err = createDirAll(paths.DestPath)
	if err != nil {
		return err
	}

	if dir.Metrics.BackupType == core.FBT_SKIP {
		backupType = core.FBT_SKIP
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run backup in "skip mode"
		options := rsync.NewOptions(rsync.WithDefaultParams("--times")).
			AddParams("--delete", "--dirs").
			AddParams(f("--include=%s", ignoreBackupSigFileName), "--exclude=*").
			SetRetryCount(retryCount).
			SetErrorHook(errorHook).
			// minimum size for empty signature file
			SetPredictedSize(core.NewFolderSize(1000))

		sessionErr, retryErr, criticalErr := rsync.RunRsyncWithRetry(progress.Context,
			options, progress.RsyncLog, nil, paths)
		if criticalErr != nil {
			return criticalErr
		}

		err = reportProgress(sessionErr, retryErr, *dir.Metrics.FullSize, plan, progress, paths, backupType, true)
		if err != nil {
			return err
		}
	} else if dir.Metrics.BackupType == core.FBT_RECURSIVE {
		backupType = core.FBT_RECURSIVE
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run full backup including content with recursion
		options := rsync.NewOptions(rsync.WithDefaultParams("--times")).
			AddParams("--delete", "--recursive").
			SetRetryCount(retryCount).
			SetErrorHook(errorHook).
			SetPredictedSize(*dir.Metrics.FullSize)

		if plan.Config.usePreviousBackupEnabled() {
			//options = append(options, "--fuzzy", "--fuzzy")
			for _, path := range prevBackupPaths {
				options.AddParams(f("--link-dest=%s", path))
			}
		}

		sessionErr, retryErr, criticalErr := rsync.RunRsyncWithRetry(progress.Context,
			options, progress.RsyncLog, nil, paths)
		if criticalErr != nil {
			return criticalErr
		}

		err = reportProgress(sessionErr, retryErr, *dir.Metrics.FullSize, plan, progress, paths, backupType, false)
		if err != nil {
			return err
		}
	} else if dir.Metrics.BackupType == core.FBT_CONTENT {
		backupType = core.FBT_CONTENT
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run backup only folder content without recursion (flat mode)
		options := rsync.NewOptions(rsync.WithDefaultParams("--times")).
			AddParams("--delete", "--dirs").
			SetRetryCount(retryCount).
			SetErrorHook(errorHook).
			SetPredictedSize(*dir.Metrics.Size)

		if plan.Config.usePreviousBackupEnabled() {
			//options = append(options, "--fuzzy", "--fuzzy")
			for _, path := range prevBackupPaths {
				options.AddParams(f("--link-dest=%s", path))
			}
		}

		sessionErr, retryErr, criticalErr := rsync.RunRsyncWithRetry(progress.Context,
			options, progress.RsyncLog, nil, paths)
		if criticalErr != nil {
			return criticalErr
		}

		err = reportProgress(sessionErr, retryErr, *dir.Metrics.Size, plan, progress, paths, backupType, false)
		if err != nil {
			return err
		}

		for _, item := range dir.Childs {
			prevBackupPaths2 := append([]string(nil), prevBackupPaths...)
			for i, path := range prevBackupPaths2 {
				prevBackupPaths2[i] = filepath.Join(path, item.Name)
			}
			err = backupDir(item, ignoreBackupSigFileName, retryCount, rsyncMaxBlockSizeMb,
				plan, progress, paths.Join(item.Name), errorHook, prevBackupPaths2)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
