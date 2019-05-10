package core

import (
	"fmt"
	"io/ioutil"
	"os"
)

// DirMetrics keeps metrics defined in 1st pass of folders tree.
// Metrics used lately in heuristic algorithm to find optimal folder tree traverse.
type DirMetrics struct {
	// Define depth from root folder. Root folder has Depth = 0.
	Depth int
	// Total count of all child folders.
	ChildrenCount int
	// "Size" metric defines summary size of all local files,
	// do not include any child folders.
	Size *FolderSize
	// "Full size" metric, which include all files and
	// child folders with their content.
	FullSize *FolderSize
	// Flag which means, that folder contain special file
	// which serves as tag to do not backup this folder.
	IgnoreToBackup bool
	// Flag which means, that this folder already marked
	// as "measured" in traverse path search.
	Measured bool
	// Type of backup for current folder defined
	// as a result of traverse path search.
	BackupType FolderBackupType
}

// Dir is a "tree data structure" to describe folder's tree
// received from the source in 1st pass of backup process to measure
// counts/sizes and to predict time necessary for backup process (ETA).
// https://en.wikipedia.org/wiki/Tree_(data_structure)
type Dir struct {
	Paths   SrcDstPath
	Name    string
	Parent  *Dir
	Childs  []*Dir
	Metrics DirMetrics
}

// BuildDirTree scans and creates Dir object which reflects
// real recursive directory structure defined by file system path
// in paths argument.
func BuildDirTree(paths SrcDstPath, ignoreBackupFileSigName string) (*Dir, error) {
	info, err := os.Stat(paths.DestPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		// does not translate this message, since it is very unlikely
		return nil, fmt.Errorf("path %q should be a folder", paths.DestPath)
	}
	root := &Dir{Name: info.Name(), Paths: paths, Metrics: DirMetrics{Depth: 0}}
	_, err = createOffsprings(root, paths, ignoreBackupFileSigName, 1)
	if err != nil {
		return nil, err
	}
	return root, nil
}

func (v *Dir) GetTotalSize() FolderSize {
	// use nested call since it's recursive
	return getTotalSize(v)
}

func (v *Dir) GetIgnoreSize() FolderSize {
	// use nested call since it's recursive
	return getIgnoreSize(v)
}

func (v *Dir) GetFullBackupSize() FolderSize {
	// use nested call since it's recursive
	return getFullBackupSize(v)
}

func (v *Dir) GetContentBackupSize() FolderSize {
	// use nested call since it's recursive
	return getContentBackupSize(v)
}

func (v *Dir) GetFoldersCount() int {
	// use nested call since it's recursive
	return getFoldersCount(v)
}

func (v *Dir) GetFoldersIgnoreCount() int {
	// use nested call since it's recursive
	return getFoldersIgnoreCount(v)
}

func containsMeasuredDir(dir *Dir) bool {
	if dir.Metrics.Measured {
		return true
	}
	for _, item := range dir.Childs {
		if containsMeasuredDir(item) {
			return true
		}
	}
	return false
}

func containsNonMeasuredDir(dir *Dir) bool {
	if !dir.Metrics.Measured {
		return true
	}
	for _, item := range dir.Childs {
		if containsNonMeasuredDir(item) {
			return true
		}
	}
	return false
}

func getTotalSize(dir *Dir) FolderSize {
	var size FolderSize
	if dir.Metrics.BackupType == FBT_CONTENT {
		size = *dir.Metrics.Size
	} else if dir.Metrics.BackupType == FBT_RECURSIVE {
		size = *dir.Metrics.FullSize
	} else if dir.Metrics.BackupType == FBT_SKIP {
		size = *dir.Metrics.FullSize
	}
	for _, item := range dir.Childs {
		size += getTotalSize(item)
	}
	return size
}

func getFullBackupSize(dir *Dir) FolderSize {
	var size FolderSize
	if dir.Metrics.BackupType == FBT_RECURSIVE {
		size = *dir.Metrics.FullSize
	}
	for _, item := range dir.Childs {
		size += getFullBackupSize(item)
	}
	return size
}

func getContentBackupSize(dir *Dir) FolderSize {
	var size FolderSize
	if dir.Metrics.BackupType == FBT_CONTENT {
		size = *dir.Metrics.Size
	}
	for _, item := range dir.Childs {
		size += getContentBackupSize(item)
	}
	return size
}

func getIgnoreSize(dir *Dir) FolderSize {
	var size FolderSize
	if dir.Metrics.BackupType == FBT_SKIP {
		size = *dir.Metrics.FullSize
	}
	for _, item := range dir.Childs {
		size += getIgnoreSize(item)
	}
	return size
}

func getFoldersIgnoreCount(dir *Dir) int {
	count := 0
	if dir.Metrics.BackupType == FBT_SKIP {
		count++
	}
	for _, item := range dir.Childs {
		count += getFoldersIgnoreCount(item)
	}
	return count
}

func getFoldersCount(dir *Dir) int {
	count := len(dir.Childs)
	for _, item := range dir.Childs {
		count += getFoldersCount(item)
	}
	return count
}

func createOffsprings(parent *Dir, paths SrcDstPath,
	sigFileIgnoreBackup string, depth int) (int, error) {

	// lg.Debug(f("Iterate path: %q", path))
	items, err := ioutil.ReadDir(paths.DestPath)
	if err != nil {
		return 0, err
	}
	if sigFileIgnoreBackupFound(items, sigFileIgnoreBackup) {
		parent.Metrics.IgnoreToBackup = true
		parent.Metrics.ChildrenCount = 1
		return 1, nil
	}
	totalCount := 1
	for _, item := range items {
		if item.IsDir() {
			name := item.Name()
			paths2 := paths.Join(name)
			dir := &Dir{Parent: parent, Name: name, Paths: paths2,
				Metrics: DirMetrics{Depth: depth}}
			count, err := createOffsprings(dir, paths2,
				sigFileIgnoreBackup, depth+1)
			if err != nil {
				return 0, err
			}
			parent.Childs = append(parent.Childs, dir)
			totalCount += count
		}
	}
	parent.Metrics.ChildrenCount = totalCount
	return totalCount, nil
}

func sigFileIgnoreBackupFound(items []os.FileInfo, sigFileIgnoreBackup string) bool {
	for _, item := range items {
		if !item.IsDir() && item.Name() == sigFileIgnoreBackup {
			return true
		}
	}
	return false
}
