package core

type FolderBackupType int

const (
	FBT_UNKNOWN FolderBackupType = iota
	FBT_SKIP
	FBT_RECURSIVE
	FBT_CONTENT
)

func (v FolderBackupType) String() string {
	var backupStr string
	switch v {
	case FBT_SKIP:
		backupStr = "skip"
	case FBT_RECURSIVE:
		backupStr = "full folder content"
	case FBT_CONTENT:
		backupStr = "folder files"
	}
	return backupStr
}

type FolderSize int64

func NewFolderSize(size int64) FolderSize {
	v := FolderSize(size)
	return v
}

func (v FolderSize) GetByteCount() uint64 {
	return uint64(v)
}

// func (v FolderSize) GetReadableSize() string {
// 	str := humanize.Bytes(v.GetByteCount())
// 	return str
// }

func (v FolderSize) Add(value FolderSize) FolderSize {
	a := v + value
	return a
}

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

type SizeProgress struct {
	Completed *FolderSize
	Skipped   *FolderSize
	Failed    *FolderSize
}

func NewProgressCompleted(size FolderSize) SizeProgress {
	this := SizeProgress{Completed: &size}
	return this
}

func NewProgressSkipped(size FolderSize) SizeProgress {
	this := SizeProgress{Skipped: &size}
	return this
}

func NewProgressFailed(size FolderSize) SizeProgress {
	this := SizeProgress{Failed: &size}
	return this
}

func (this *SizeProgress) Add(size SizeProgress) {
	if size.Completed != nil {
		if this.Completed == nil {
			this.Completed = size.Completed
		} else {
			done := *this.Completed + *size.Completed
			this.Completed = &done
		}
	}
	if size.Skipped != nil {
		if this.Skipped == nil {
			this.Skipped = size.Skipped
		} else {
			done := *this.Skipped + *size.Skipped
			this.Skipped = &done
		}
	}
	if size.Failed != nil {
		if this.Failed == nil {
			this.Failed = size.Failed
		} else {
			done := *this.Failed + *size.Failed
			this.Failed = &done
		}
	}
}

func (this *SizeProgress) GetTotal() FolderSize {
	var total FolderSize
	if this.Completed != nil {
		total += *this.Completed
	}
	if this.Skipped != nil {
		total += *this.Skipped
	}
	if this.Failed != nil {
		total += *this.Failed
	}
	return total
}
