package backup

import (
	"time"

	"github.com/d2r2/go-rsync/core"
)

type BackupNode struct {
	SourceRsync string `toml:"src_rsync"`
	DestSubPath string `toml:"dst_subpath"`
}

// BackupNodePlan contain information about single rsync source backup.
type BackupNodePlan struct {
	BackupNode BackupNode
	RootDir    *core.Dir
}

// BackupPlan keep all necessary information obtained from
// preferences and 1st backup pass to start backup process.
type BackupPlan struct {
	Config     *Config
	Nodes      []BackupNodePlan
	BackupSize core.FolderSize
}

type Notifier interface {
	NotifyPlanStage_NodeStructureStartInquiry(sourceID int,
		sourceRsync string) error
	NotifyPlanStage_NodeStructureDoneInquiry(sourceID int,
		sourceRsync string, dir *core.Dir) error
	NotifyBackupStage_FolderStartBackup(rootDest string,
		paths core.SrcDstPath, backupType core.FolderBackupType,
		leftToBackup core.FolderSize,
		timePassed time.Duration, eta *time.Duration,
	) error
	NotifyBackupStage_FolderDoneBackup(rootDest string,
		paths core.SrcDstPath, backupType core.FolderBackupType,
		leftToBackup core.FolderSize, sizeDone core.SizeProgress,
		timePassed time.Duration, eta *time.Duration,
		sessionErr error) error
}
