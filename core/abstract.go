package core

// FolderBackupType define how
// to backup folder content.
type FolderBackupType int

const (
	// Undefined backup approach.
	FBT_UNKNOWN FolderBackupType = iota
	// Skip to backup folder content (including subfolders).
	FBT_SKIP
	// Backup full folder content including all subfolders.
	FBT_RECURSIVE
	// Backup only files located directly in the folder.
	// Do not backup subfolders.
	FBT_CONTENT
)

// String implement Stringer interface.
func (v FolderBackupType) String() string {
	var backupStr string
	switch v {
	case FBT_SKIP:
		backupStr = "skip"
	case FBT_RECURSIVE:
		backupStr = "full folder content"
	case FBT_CONTENT:
		backupStr = "folder files"
	case FBT_UNKNOWN:
		backupStr = "<undefined>"
	}
	return backupStr
}

// FolderSize used to signify size of backup objects.
type FolderSize int64

func NewFolderSize(size int64) FolderSize {
	v := FolderSize(size)
	return v
}

// GetByteCount returns size of FolderSize in bytes.
func (v FolderSize) GetByteCount() uint64 {
	return uint64(v)
}

// Add combines sizes of two FolderSize objects.
func (v FolderSize) Add(value FolderSize) FolderSize {
	a := v + value
	return a
}

// AddSizeProgress accumulate all sizes from SizeProgress with instance size.
func (v FolderSize) AddSizeProgress(value SizeProgress) FolderSize {
	var totalDone FolderSize
	if value.Completed != nil {
		totalDone += *value.Completed
	}
	if value.Failed != nil {
		totalDone += *value.Failed
	}
	if value.Skipped != nil {
		totalDone += *value.Skipped
	}
	a := v + totalDone
	return a
}

// SizeProgress keeps all sizes which may arise during backup process.
type SizeProgress struct {
	// Completed signify successfully backed up size.
	Completed *FolderSize
	// Skipped signify size that was skipped during backup process.
	Skipped *FolderSize
	// Failed signify size that was not backed up due to some issues.
	Failed *FolderSize
}

// NewProgressCompleted creates the SizeProgress object
// with the size that was successfully backed up.
func NewProgressCompleted(size FolderSize) SizeProgress {
	this := SizeProgress{Completed: &size}
	return this
}

// NewProgressSkipped creates the SizeProgress object
// with the size that was skipped.
func NewProgressSkipped(size FolderSize) SizeProgress {
	this := SizeProgress{Skipped: &size}
	return this
}

// NewProgressSkipped creates the SizeProgress object
// with the size that was not backed up due to some issues.
func NewProgressFailed(size FolderSize) SizeProgress {
	this := SizeProgress{Failed: &size}
	return this
}

// Add combines sizes of two SizeProgress objects.
func (v *SizeProgress) Add(size SizeProgress) {
	if size.Completed != nil {
		if v.Completed == nil {
			v.Completed = size.Completed
		} else {
			done := *v.Completed + *size.Completed
			v.Completed = &done
		}
	}
	if size.Skipped != nil {
		if v.Skipped == nil {
			v.Skipped = size.Skipped
		} else {
			done := *v.Skipped + *size.Skipped
			v.Skipped = &done
		}
	}
	if size.Failed != nil {
		if v.Failed == nil {
			v.Failed = size.Failed
		} else {
			done := *v.Failed + *size.Failed
			v.Failed = &done
		}
	}
}

// GetTotal gets total SizeProgress size.
func (v *SizeProgress) GetTotal() FolderSize {
	var total FolderSize
	return total.AddSizeProgress(*v)
}
