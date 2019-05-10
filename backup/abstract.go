package backup

import (
	"time"

	"github.com/d2r2/go-rsync/core"
)

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
