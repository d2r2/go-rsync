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
	DoubleSplitLogLine string = strings.Repeat("=", 100)
	SingleSplitLogLine string = strings.Repeat("-", 100)
)

// BuildBackupPlan perform 1st stage (plan stage) to measure RSYNC source volume
// to backup and find optimal traverse path of source directory tree.
// Use plan built in 1st stage later in 2nd stage.
func BuildBackupPlan(ctx context.Context, lg logger.PackageLog, config *Config,
	modules []Module, notifier Notifier) (*Plan, *Progress, error) {

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

	// create specific RSYNC log file (might be activated in
	// backup session preference for debug purpose)
	rsyncLog := config.getRsyncLoggingSettings()
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

	progress.Log.Info(DoubleSplitLogLine)
	progress.Log.Info(locale.T(MsgLogPlanStageStarting, nil))
	progress.Log.Info(locale.T(MsgLogPlanStageStartTime,
		struct{ Time string }{Time: progress.StartPlanTime.Format("2006 Jan 2 15:04:05")}))

	list := []Node{}
	var totalBackupSize core.FolderSize
	progress.Log.Info(locale.TP(MsgLogPlanStartIterateViaNSources,
		struct{ SourceCount int }{SourceCount: len(modules)},
		len(modules)))

	for i, item := range modules {
		progress.Log.Info(SingleSplitLogLine)
		err := progress.EventPlanStage_NodeStructureStartInquiry(i, item.SourceRsync)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}

		dr, backupSize, err := estimateNode(ctx, item.AuthPassword, item, progress, config)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}
		if backupSize != nil {
			totalBackupSize += *backupSize
		}

		err = progress.EventPlanStage_NodeStructureDoneInquiry(i, item.SourceRsync, dr)
		if err != nil {
			progress.Log.Error(err)
			return nil, nil, err
		}

		node := Node{Module: item, RootDir: dr}
		list = append(list, node)
	}
	progress.Log.Info(SingleSplitLogLine)
	progress.FinishPlanStage()
	//	progress.Log.Debugf("Plan: %+v", list)
	progress.Log.Info(locale.T(MsgLogPlanStageEndTime,
		struct{ Time string }{Time: progress.EndPlanTime.Format("2006 Jan 2 15:04:05")}))
	backup := &Plan{Config: config, Nodes: list, BackupSize: totalBackupSize}
	//progress.Log.Debugf("Plan: %+v", backup)
	return backup, progress, nil
}

func estimateNode(ctx context.Context, password *string, module Module, progress *Progress,
	config *Config) (*core.Dir, *core.FolderSize, error) {

	tempDir, err := ioutil.TempDir("", "backup_dir_tree_")
	if err != nil {
		return nil, nil, err
	}
	defer os.RemoveAll(tempDir)

	progress.Log.Info(locale.T(MsgLogPlanStageUseTemporaryFolder,
		struct{ Path string }{Path: tempDir}))

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(module.SourceRsync, ""),
		DestPath:        filepath.Join(tempDir, module.DestSubPath),
	}

	err = createDirAll(paths.DestPath)
	if err != nil {
		err = errors.New(f("%s: %v", locale.T(MsgLogPlanStageUseTemporaryFolder,
			struct{ Path string }{Path: tempDir}), err))
		return nil, nil, err
	}

	// RSYNC settings to copy only folder's structure and some specific files
	options := rsync.NewOptions(rsync.WithDefaultParams([]string{"--recursive"})).
		AddParams(f("--include=%s", "*"+"/")).
		AddParams(f("--include=%s", config.SigFileIgnoreBackup)).
		AddParams(f("--exclude=%s", "*")).
		SetRetryCount(config.RsyncRetryCount).
		SetAuthPassword(password)
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
	count, err := MeasureDir(ctx, password, dir, config.RsyncRetryCount, progress.RsyncLog, blockSize)
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

// RunBackup perform whole 2nd stage (backup stage) here, then save and
// report about completion to session logs.
func (plan *Plan) RunBackup(progress *Progress, destPath string,
	errorHookCall rsync.ErrorHookCall) error {

	// Execute backup stage
	err := runBackup(plan, progress, destPath, errorHookCall)
	if err != nil {
		progress.Log.Error(locale.T(MsgLogBackupStageCriticalError,
			struct{ Error error }{Error: err}))
	}

	// Next lines should be executed even if backup failed and err variable is not empty,
	// to store log files in backup destination folder.

	if progress.RsyncLog != nil {
		rsyncLogFileName := path.Join(progress.GetBackupFullPath(progress.BackupFolder), GetRsyncLogFileName())
		progress.Log.Info(locale.T(MsgLogBackupStageSaveRsyncExtraLogTo,
			struct{ Path string }{Path: rsyncLogFileName}))
	}

	logFileName := path.Join(progress.GetBackupFullPath(progress.BackupFolder), GetLogFileName())
	progress.Log.Info(locale.T(MsgLogBackupStageSaveLogTo,
		struct{ Path string }{Path: logFileName}))

	progress.SayGoodbye(progress.Log)

	return err
}

// Perform whole 2nd stage (backup stage) here.
func runBackup(plan *Plan, progress *Progress, destPath string, errorHookCall rsync.ErrorHookCall) error {

	progress.TotalProgress = &core.SizeProgress{}
	progress.Progress = &core.SizeProgress{}
	progress.StartBackupStage()

	progress.Log.Info(DoubleSplitLogLine)
	progress.Log.Info(locale.T(MsgLogBackupStageStarting, nil))
	progress.Log.Info(locale.T(MsgLogBackupStageStartTime,
		struct{ Time string }{Time: progress.StartBackupTime.Format("2006 Jan 2 15:04:05")}))

	// create new folder with date/time stamp for new backup session
	err := createDirInBackupStage(destPath)
	if err != nil {
		return err
	}
	progress.SetRootDestination(destPath)
	backupFolder := GetBackupFolderName(true, &progress.StartBackupTime)
	path := progress.GetBackupFullPath(backupFolder)
	err = createDirInBackupStage(path)
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

	// search for previous backup sessions: this might activate deduplication capabilities
	progress.Log.Info(locale.T(MsgLogBackupStageDiscoveringPreviousBackups, nil))
	prevBackups, err := FindPrevBackupPathsByNodeSignatures(progress.Log, destPath,
		GetNodeSignatures(plan.GetModules()), plan.Config.numberOfPreviousBackupToUse())
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

	// loop through all RSYNC source to backup
	for i, node := range plan.Nodes {
		progress.Log.Info(SingleSplitLogLine)
		progress.Log.Info(locale.T(MsgLogBackupStageStartToBackupFromSource,
			struct {
				SeqID       int
				RsyncSource string
			}{SeqID: i + 1, RsyncSource: node.Module.SourceRsync}))

		// select previous backup sessions to use for deduplication
		sourceID := GenerateSourceID(node.Module.SourceRsync)
		prevBackups2 := prevBackups.FilterBySourceID(sourceID)
		// run specific RSYNC source to backup
		err := runBackupNode(plan, node, /*plan.Config,*/
			progress, destPath2, errorHookCall, prevBackups2)
		if err != nil {
			return err
		}
	}

	// debug
	LocalLog.Debugf("BACKUP FINAL: total progress %+v", progress.TotalProgress)
	LocalLog.Debugf("BACKUP FINAL: left to backup %+v", progress.LeftToBackup(plan))

	// rename backup session folder, since backup process is completed
	progress.Log.Info(SingleSplitLogLine)
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
	progress.Log.Info(locale.T(MsgLogBackupStageRenameDestination,
		struct{ Path string }{Path: destPath3}))

	// create signature auxiliary file: used to search for previous backup sessions
	// in order to activate deduplication capabilities
	err = CreateMetadataSignatureFile(plan.GetModules(), destPath3)
	if err != nil {
		return err
	}

	progress.FinishBackupStage()
	progress.Log.Info(locale.T(MsgLogBackupStageEndTime,
		struct{ Time string }{Time: progress.EndBackupTime.Format("2006 Jan 2 15:04:05")}))

	// print statistics
	err = progress.PrintTotalStatistics(progress.Log, plan)
	if err != nil {
		return err
	}

	return nil
}

// Perform backup of one source defined in backup session preferences.
func runBackupNode(plan *Plan, node Node, progress *Progress, destRootPath string,
	errorHookCall rsync.ErrorHookCall, prevBackups *PrevBackups) error {

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(node.Module.SourceRsync, ""),
		DestPath:        filepath.Join(destRootPath, node.Module.DestSubPath),
	}

	progress.Progress = &core.SizeProgress{}
	err := backupDir(node.RootDir, &node.Module,
		plan, progress, paths, errorHookCall, prevBackups.GetDirPaths())
	return err
}

// Reformat and localize error message here, if possible.
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

// Report backup progress on each backup step made.
// Report here not only successfully performed steps, but anything
// including steps ended with errors.
func reportProgress(sessionErr, retryErr error, size core.FolderSize,
	plan *Plan, progress *Progress, paths core.SrcDstPath,
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

// Major function to make all necessary RSYNC calls to execute backup process step by step.
func backupDir(dir *core.Dir, module *Module, plan *Plan, progress *Progress,
	paths core.SrcDstPath, errorHookCall rsync.ErrorHookCall, prevBackupPaths []string) error {

	var err error
	var backupType core.FolderBackupType
	defParams := []string{"--times"}

	err = createDirInBackupStage(paths.DestPath)
	if err != nil {
		return err
	}
	// subtree marked as "skipped" due to file signature found in the folder
	if dir.Metrics.BackupType == core.FBT_SKIP {
		backupType = core.FBT_SKIP
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run backup in "skip mode"
		options := rsync.NewOptions(rsync.WithDefaultParams(
			GetRsyncParams(plan.Config, module, defParams))).AddParams("--delete", "--dirs").
			// AddParams("--super").
			// AddParams("--fake-super").
			AddParams(f("--include=%s", plan.Config.SigFileIgnoreBackup), "--exclude=*").
			SetRetryCount(plan.Config.RsyncRetryCount).
			SetAuthPassword(module.AuthPassword).
			// minimum size for empty signature file
			SetErrorHook(rsync.NewErrorHook(errorHookCall, core.NewFolderSize(1*core.KB)))

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
		// subtree processed at once without splitting to the peaces
		backupType = core.FBT_RECURSIVE
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run full backup including content with recursion
		options := rsync.NewOptions(rsync.WithDefaultParams(
			GetRsyncParams(plan.Config, module, defParams))).AddParams("--delete", "--recursive").
			// AddParams("--super").
			// AddParams("--fake-super").
			SetRetryCount(plan.Config.RsyncRetryCount).
			SetAuthPassword(module.AuthPassword).
			SetErrorHook(rsync.NewErrorHook(errorHookCall, *dir.Metrics.FullSize))

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
		// process only current folder, then go deep to process subfolders recursively
		backupType = core.FBT_CONTENT
		err = progress.EventBackupStage_FolderStartBackup(paths, backupType, plan)
		if err != nil {
			return err
		}
		// run backup only folder content without nested folders (flat mode)
		options := rsync.NewOptions(rsync.WithDefaultParams(
			GetRsyncParams(plan.Config, module, defParams))).AddParams("--delete", "--dirs").
			// AddParams("--super").
			// AddParams("--fake-super").
			SetRetryCount(plan.Config.RsyncRetryCount).
			SetAuthPassword(module.AuthPassword).
			SetErrorHook(rsync.NewErrorHook(errorHookCall, *dir.Metrics.Size))

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

		// process subfolders recursively
		for _, item := range dir.Childs {
			prevBackupPaths2 := append([]string(nil), prevBackupPaths...)
			for i, path := range prevBackupPaths2 {
				prevBackupPaths2[i] = filepath.Join(path, item.Name)
			}
			err = backupDir(item, module,
				plan, progress, paths.Join(item.Name), errorHookCall, prevBackupPaths2)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
