package backup

import (
	"context"
	"math"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/rsync"
)

// =============================================================================================
//
// This is most scientific part of application :)
//
// Contains heuristic algorithm for searching of optimal backup RSYNC source
// traverse path during backup process.
//
// Prerequisites:
//		- RSYNC source directory tree structure to backup. This structure downloaded
//			to temporary file space to build directory tree in memory.
//		- Backup block size. Define size of backup to do at once (at one call of RSYNC utility).
//		- RSYNC utility might backup in 2 modes:
//			* "full backup" when folder content backup recursively.
//			* "local content backup", when backup only flat content, which
//				include only files from folder, do not include any nested folders.
//
// Task:
//		Find optimal (or close to optimal) traverse path of directory tree to limit
//		number of calls to RSYNC. Limiting number of calls to RSYNC might significantly
//		decrease backup time in case when we have folders with a lot of child folders
//		(deep nested structure) with small content. In such cases we are trying to backup
//		such folders recursively, which might significantly speed up planing and backup stages.
//
// =============================================================================================

const MaxUint = ^uint(0)
const MaxInt = int(MaxUint >> 1)
const MinInt = -MaxInt - 1

func markMesuredAll(dir *core.Dir) {
	dir.Metrics.Measured = true
	for _, item := range dir.Childs {
		markMesuredAll(item)
	}
}

func findNonMeasuredIgnoreDir(dir *core.Dir) *core.Dir {
	if !dir.Metrics.Measured && dir.Metrics.IgnoreToBackup {
		return dir
	}
	for _, item := range dir.Childs {
		if ignore := findNonMeasuredIgnoreDir(item); ignore != nil {
			return ignore
		}
	}
	return nil
}

func getNonMeasuredDir(dir *core.Dir) *core.Dir {
	if ignore := findNonMeasuredIgnoreDir(dir); ignore != nil {
		return ignore
	}
	if !dir.Metrics.Measured {
		return dir
	}
	for _, item := range dir.Childs {
		child := getNonMeasuredDir(item)
		if child != nil {
			return child
		}
	}
	return nil
}

// measureLocalUpToRoot calculate "local size" metric for chain of parent folders
// up to root, if not yet defined.
func measureLocalUpToRoot(ctx context.Context, dir *core.Dir, retryCount *int, log *rsync.Logging) error {
	item := dir
	for {
		item = item.Parent
		if item == nil {
			break
		}
		var err error
		size := item.Metrics.Size
		if size == nil {
			size, err = rsync.ObtainDirLocalSize(ctx, item, retryCount, log)
			if err != nil {
				return err
			}
		}
		item.Metrics.Measured = true
		item.Metrics.BackupType = core.FBT_CONTENT
		item.Metrics.Size = size
	}
	return nil
}

func findDownNonMeasuredDirByWeight(dir *core.Dir, weight int) *core.Dir {
	if !dir.Metrics.Measured && dir.Metrics.ChildrenCount <= weight {
		return dir
	} else if !dir.Metrics.Measured && dir.Metrics.ChildrenCount > weight && len(dir.Childs) == 0 {
		return dir
	} else {
		var found *core.Dir
		minDiff := MaxInt
		for _, item := range dir.Childs {
			child := findDownNonMeasuredDirByWeight(item, weight)
			if child != nil {
				if int(math.Abs(float64(child.Metrics.ChildrenCount-weight))) < minDiff {
					found = child
					minDiff = int(math.Abs(float64(child.Metrics.ChildrenCount - weight)))
				}
			}
		}
		return found
	}
}

func findUpNonMeasuredDirByWeight(dir *core.Dir, weight int) *core.Dir {
	item := dir
	parent := item.Parent
	for {
		if parent == nil || parent.Metrics.Measured {
			//LocalLog.Debugf("From %v up to %v", dir.Paths.RsyncSourcePath, item.Paths.RsyncSourcePath)
			return item
		}
		if parent.Metrics.ChildrenCount > weight {
			//LocalLog.Debugf("From %v up to %v", dir.Paths.RsyncSourcePath, item.Paths.RsyncSourcePath)
			return parent
		}
		item = parent
		parent = item.Parent
	}
}

func MeasureDir(ctx context.Context, dir *core.Dir, retryCount *int, log *rsync.Logging,
	blockSize *backupBlockSizeSettings) (int, error) {

	totalCount := 0
	for {
		found, count, err := searchDownOptimalDir(ctx, dir, retryCount, log, blockSize)
		if err != nil {
			return 0, err
		}
		totalCount += count
		if found == nil {
			break
		}

		if found.Metrics.IgnoreToBackup {
			LocalLog.Debugf("Selected for skip (count=%v): %v", count, found.Paths.RsyncSourcePath)
			found.Metrics.BackupType = core.FBT_SKIP
		} else {
			LocalLog.Debugf("Selected for full backup (count=%v): %v", count, found.Paths.RsyncSourcePath)
			found.Metrics.BackupType = core.FBT_RECURSIVE
		}

		markMesuredAll(found)
		err = measureLocalUpToRoot(ctx, found, retryCount, log)
		if err != nil {
			return 0, err
		}
	}
	return totalCount, nil
}

func getFullSizesUpToRoot(dir *core.Dir) ([]uint64, []int) {
	sizes := []uint64{}
	depths := []int{}
	item := dir
	for {
		if item.Metrics.FullSize != nil {
			sizes = append([]uint64{item.Metrics.FullSize.GetByteCount()}, sizes...)
			depths = append([]int{item.Metrics.Depth}, depths...)
		}
		item = item.Parent
		if item == nil {
			break
		}
	}
	return sizes, depths
}

func getRoot(dir *core.Dir) *core.Dir {
	root := dir
	for {
		if root.Parent != nil {
			root = root.Parent
		} else {
			break
		}
	}
	return root
}

// calcFullSizesWithRoot calc "full size" metric for current folder and root, if not defined yet.
func calcFullSizesWithRoot(ctx context.Context, dir *core.Dir, retryCount *int, log *rsync.Logging) (int, error) {
	count := 0
	root := getRoot(dir)
	if root.Metrics.FullSize == nil {
		fullSize, err := rsync.ObtainDirFullSize(ctx, root, retryCount, log)
		if err != nil {
			return 0, err
		}
		root.Metrics.FullSize = fullSize
		count++
	}
	if dir.Metrics.FullSize == nil {
		fullSize, err := rsync.ObtainDirFullSize(ctx, dir, retryCount, log)
		if err != nil {
			return 0, err
		}
		dir.Metrics.FullSize = fullSize
		count++
	}
	return count, nil
}

// findDownNonMeasuredDirByDepth find folder have not measured yet
// from specific "depth".
func findDownNonMeasuredDirByDepth(dir *core.Dir, depth int) *core.Dir {
	if !dir.Metrics.Measured && dir.Metrics.Depth >= depth {
		return dir
	} else if !dir.Metrics.Measured && dir.Metrics.Depth < depth && len(dir.Childs) == 0 {
		return dir
	} else {
		var found *core.Dir
		maxWeight := MinInt
		for _, item := range dir.Childs {
			child := findDownNonMeasuredDirByDepth(item, depth)
			if child != nil {
				if child.Metrics.ChildrenCount > maxWeight {
					found = child
					maxWeight = child.Metrics.ChildrenCount
				}
			}
		}
		return found
	}
}

// Round returns the nearest integer, rounding ties away from zero.
func round(x float64) float64 {
	t := math.Trunc(x)
	if math.Abs(x-t) >= 0.5 {
		return t + math.Copysign(1, x)
	}
	return t
}

// interpolLagrange implement prediction of next f(x), when set of fi(xi) provided.
func interpolLagrange(sizes []uint64, depths []int, searchSize uint64) int {
	result := 0.0
	for i := 0; i < len(sizes); i++ {
		term := float64(depths[i])
		for j := 0; j < len(sizes); j++ {
			if j != i {
				term = term * float64(searchSize-sizes[j]) / float64(sizes[i]-sizes[j])
			}
		}
		result += term
	}
	return int(round(result))
}

// interpolLinear implement prediction of next f(x), when dots f1(x1), f2(x2) provided.
func interpolLinear(size1, size2 uint64, depth1, depth2 int, searchSize uint64) int {
	depth3 := float64(depth1) + float64(searchSize-size1)*float64(depth2-depth1)/float64(size2-size1)
	return int(round(depth3))
}

// selectChildByWeight choose child folder, where "weight" metric is maximized.
func selectChildByWeight(dir *core.Dir) *core.Dir {
	var found *core.Dir
	maxWeight := MinInt
	for _, item := range dir.Childs {
		if maxWeight < item.Metrics.ChildrenCount {
			found = item
			maxWeight = item.Metrics.ChildrenCount
		}
	}
	return found
}

// backupBlockSizeSettings provide default "backup block size at once"
// taken from preference menu.
type backupBlockSizeSettings struct {
	AutoManageBackupBlockSize bool
	BackupBlockSize           uint64
}

// calcOptimalBackupBlockSize contains simple formula to
// gives backup block size low/high limits obtained from
// total backup size.
func calcOptimalBackupBlockSize(dir *core.Dir) uint64 {
	const splitTo = 50
	root := getRoot(dir)
	bs := root.Metrics.FullSize.GetByteCount() / splitTo
	if bs > core.MegabytesToBytes(5000) {
		bs = core.MegabytesToBytes(5000)
	} else if bs < core.MegabytesToBytes(300) {
		bs = core.MegabytesToBytes(300)
	}
	return bs
}

// searchDownOptimalDir is a main recurrent function to find optimal (or close to optimal)
// traverse path to backup source minimizing number of RSYNC utility calls.
func searchDownOptimalDir(ctx context.Context, dir *core.Dir, retryCount *int, log *rsync.Logging,
	blockSize *backupBlockSizeSettings) (*core.Dir, int, error) {

	LocalLog.Debugf("Start searching optimal folder from root %v",
		dir.Paths.RsyncSourcePath)

	found := getNonMeasuredDir(dir)

	if found != nil {
		LocalLog.Debugf("Get non-measured candidate %v to test",
			found.Paths.RsyncSourcePath)
	}

	totalFullSizeCount := 0
	if found != nil {
		count, err := calcFullSizesWithRoot(ctx, found, retryCount, log)
		if err != nil {
			return nil, 0, err
		}
		totalFullSizeCount += count

		if blockSize.AutoManageBackupBlockSize {
			bs := calcOptimalBackupBlockSize(found)
			if blockSize.BackupBlockSize != bs {
				blockSize.BackupBlockSize = bs
			}
		}

		if found.Metrics.IgnoreToBackup {
			return found, totalFullSizeCount, nil
		}

		// Extract "full size" metrics from current folder up to root (with inversed order in output)
		// to use it later for interpolation.
		sizes, depths := getFullSizesUpToRoot(found)

		// If from the start [found] directory get fit in size defined by blockSize.BackupBlockSize,
		// then stop searching anymore and return first candidate found.
		if sizes[0] <= blockSize.BackupBlockSize {
			return found, totalFullSizeCount, nil
		}

		// Case of first iteration, when only root measured for "full size".
		// We can't interpolate here, because for interpolation we need as minimum 2 points.
		// So, we use other approach: we employ "bisection method" to select next
		// candidate for optimal "block size".
		if len(sizes) == 1 {
			root := getRoot(found)
			next := findDownNonMeasuredDirByWeight(found, root.Metrics.ChildrenCount/2)
			if next == found {
				return next, totalFullSizeCount, nil
			} else {
				count, err := calcFullSizesWithRoot(ctx, next, retryCount, log)
				if err != nil {
					return nil, 0, err
				}
				totalFullSizeCount += count

				if next.Metrics.FullSize.GetByteCount() > blockSize.BackupBlockSize {
					next, count, err = searchDownOptimalDir(ctx, next, retryCount, log, blockSize)
					if err != nil {
						return nil, 0, err
					}
					totalFullSizeCount += count
				}
				return next, totalFullSizeCount, nil
			}
			// In case, when we have more than 1 folder with "full size" metric,
			// we can interpolate and predict to find next folder with appropriate "block size".
		} else {
			//depth := interpolLagrange(sizes, depths, blockSize.BackupBlockSize)
			depth := interpolLinear(sizes[0], sizes[len(sizes)-1],
				depths[0], depths[len(depths)-1], blockSize.BackupBlockSize)
			LocalLog.Debugf("Found depth %v from [sizes=%v, depths=%v] for size %v",
				depth, sizes, depths, blockSize.BackupBlockSize)

			LocalLog.Debugf("Get dir by depth %v starting from %q", depth, found.Paths.RsyncSourcePath)

			next := findDownNonMeasuredDirByDepth(found, depth)
			count, err := calcFullSizesWithRoot(ctx, next, retryCount, log)
			if err != nil {
				return nil, 0, err
			}
			totalFullSizeCount += count
			if next.Metrics.FullSize.GetByteCount() > blockSize.BackupBlockSize && len(next.Childs) > 0 {
				next = selectChildByWeight(next)
				next, count, err = searchDownOptimalDir(ctx, next, retryCount, log, blockSize)
				if err != nil {
					return nil, 0, err
				}
				totalFullSizeCount += count

			}
			return next, totalFullSizeCount, nil
		}
	}
	return nil, 0, nil
}

/*
func MeasureDir2(dir *core.Dir, ctx context.Context, retryCount *int, log *rsync.Logging,
	mbSize uint64) (int, error) {
	totalCount := 0
	for {
		found, count, err := searchDownOptimalDir2(dir, ctx, retryCount, log, mbSize)
		if err != nil {
			return 0, err
		}
		totalCount += count
		if found == nil {
			break
		}
		if found.Metrics.IgnoreToBackup {
			LocalLog.Debugf("Selected for skip (count=%v): %v", count, found.Paths.RsyncSourcePath)
		} else {
			LocalLog.Debugf("Selected for full backup (count=%v): %v", count, found.Paths.RsyncSourcePath)
		}
		markMesuredAll(found)
		err = measureLocalUpToRoot(found, ctx, retryCount, log)
		if err != nil {
			return 0, err
		}
	}
	return totalCount, nil
}

func compareDirVs(dir *core.Dir, ctx context.Context, retryCount *int, log *rsync.Logging,
	mbSize uint64) (cmp string, x, y int64, count int, err error) {

	mbs := mbSize * 1024 * 1024
	fullSize := dir.Metrics.FullSize
	if fullSize == nil {
		fullSize, err = obtainFullSize(dir, ctx, retryCount, log)
		if err != nil {
			return "", 0, 0, 0, err
		}
		count = 1
		LocalLog.Debugf("Get %q full size (weight=%v): %v", dir.Paths.RsyncSourcePath,
			dir.Metrics.FullCount, humanize.Bytes(fullSize.GetByteCount()))
		dir.Metrics.FullSize = fullSize
	}
	if dir.Metrics.IgnoreToBackup ||
		fullSize.GetByteCount() == mbs ||
		fullSize.GetByteCount() > mbs && len(dir.Childs) == 0 {
		return "=", int64(dir.Metrics.FullCount), int64(fullSize.GetByteCount()), count, nil
	}
	if fullSize.GetByteCount() > mbs {
		return ">", int64(dir.Metrics.FullCount), int64(fullSize.GetByteCount()), count, nil
	} else {
		return "<", int64(dir.Metrics.FullCount), int64(fullSize.GetByteCount()), count, nil
	}
}

func interpolLinear2(x1, x2, y1, y2, y3 int64) (x3 *int64) {
	if y2-y1 == 0 {
		return nil
	} else {
		var y float64
		//if y3 > y1 {
		y = float64(y3 - y1)
		//} else {
		//y = float64(y1 - y3)
		//}
		x31 := float64(x2-x1)*y/float64(y2-y1) + float64(x1)
		x32 := int64(x31)
		return &x32
	}
}

type DirSizeSorted struct {
	Dirs []*core.Dir
}

func (v *core.DirSizeSorted) Len() int {
	return len(v.Dirs)
}

func (v *core.DirSizeSorted) Swap(i, j int) {
	v.Dirs[i], v.Dirs[j] = v.Dirs[j], v.Dirs[i]
}

func (v *core.DirSizeSorted) Less(i, j int) bool {
	if v.Dirs[i].Metrics.FullSize == v.Dirs[j].Metrics.FullSize {
		return v.Dirs[i].Metrics.FullCount > v.Dirs[j].Metrics.FullCount
	} else {
		return v.Dirs[i].Metrics.FullSize.GetByteCount() >
			v.Dirs[j].Metrics.FullSize.GetByteCount()
	}
}

func (v *core.DirSizeSorted) Add(dir *core.Dir) {
	v.Dirs = append(v.Dirs, dir)
}

func (v *core.DirSizeSorted) Clear() {
	v.Dirs = nil
}

func (v *core.DirSizeSorted) Count() int {
	return len(v.Dirs)
}

func searchDownOptimalDir2(dir *core.Dir, ctx context.Context, retryCount *int, log *rsync.Logging,
	mbSize uint64) (*core.Dir, int, error) {

	div := 2
	var moveUp bool
	//var prevMoveUp bool
	count := 0
	found := getNonMeasuredDir(dir)
	first := true
	//var lastX, lastY int64
	//var x, y int64
	if found != nil {
		item := found
		weight := found.Metrics.FullCount
		//var prev *core.Dir
		candidates := new(DirSizeSorted)
		for {
			//LocalLog.Debugf("Found non measured dir: %v", item.Paths.RsyncSourcePath)
			var cmp string
			var err error
			var cnt int
			//lastX, lastY = x, y
			cmp, _, _, //, x, y,
				cnt, err = compareDirVs(item, ctx, retryCount, log, mbSize)
			if err != nil {
				return nil, 0, err
			}
			count += cnt
			if cmp == "=" || cmp == "<" && first {
				if item.Metrics.IgnoreToBackup {
					item.Metrics.BackupType = io.FBT_SKIP
				} else {
					item.Metrics.BackupType = io.FBT_RECURSIVE
				}
				return item, count, nil
			} else if cmp == ">" {
				LocalLog.Debugf("Move down from: %v", item.Paths.RsyncSourcePath)
				//LocalLog.Debugf("Candidate count %v", candidates.Count())
				if candidates.Count() > 0 {
					sort.Sort(candidates)
					selected := candidates.Dirs[0]
					selected.Metrics.BackupType = io.FBT_RECURSIVE
					return selected, count, nil
				}
				//prevMoveUp = moveUp
				moveUp = false

				if found.Metrics.FullCount/div > 0 {
					weight -= found.Metrics.FullCount / div
				} else {
					weight--
				}

				//x := interpolLinear2(x, lastX, y, lastY, int64(mbSize*1024*1024))
				//if x == nil {
				//	weight--
				//} else {
				//	weight = int(*x)
				//}

				//LocalLog.Debugf("New weight=%v, interpol=%v", weight, newWeight)
				//weight = int(newWeight)
			} else if cmp == "<" {

				candidates.Add(item)
				//LocalLog.Debugf("Add candidate %v", item.Paths.RsyncSourcePath)
				LocalLog.Debugf("Move up from: %v", item.Paths.RsyncSourcePath)
				//prev = item
				//prevMoveUp = moveUp
				moveUp = true

				if found.Metrics.FullCount/div > 0 {
					weight += found.Metrics.FullCount / div
				} else {
					weight++
				}

				//x := interpolLinear2(x, lastX, y, lastY, int64(mbSize*1024*1024))
				//if x == nil {
				//	weight++
				//} else {
				//	weight = int(*x)
				//}

				//LocalLog.Debugf("New weight=%v, interpol=%v", weight, newWeight)
				//weight = int(newWeight)
			}
			if moveUp {
				item = findUpNonMeasuredDirByWeight(item, weight)
			} else {
				item = findDownNonMeasuredDirByWeight(found, weight)
			}

			if found.Metrics.FullCount/div > 0 {
				div *= 2
			}
			//count++
			first = false
		}
	}
	return nil, count, nil
}
*/
