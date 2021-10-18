//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2022 Denis Dyakov <denis.dyakov@gma**.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

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

const (
	// MaxUint keep maximum UINT value.
	MaxUint = ^uint(0)
	// MaxInt keep maximum poistive INT value.
	MaxInt = int(MaxUint >> 1)
	// MinInt keep minimum negative INT value.
	MinInt = -MaxInt - 1
)

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
// up to root, if not yet defined. Additionally mark all folder's chain up to root
// with core.FBT_CONTENT attribute.
func measureLocalUpToRoot(ctx context.Context, password *string, dir *core.Dir, retryCount *int,
	rsyncProtocol string, log *rsync.Logging) error {

	item := dir
	for {
		item = item.Parent
		if item == nil {
			break
		}
		var err error
		size := item.Metrics.Size
		if size == nil {
			size, err = rsync.ObtainDirLocalSize(ctx, password, item, retryCount, rsyncProtocol, log)
			if err != nil {
				return err
			}
		}
		item.Metrics.Measured = true
		// Mark up folder as "content backup" type, when during backup stage will be
		// backed up only files, but not nested folders.
		item.Metrics.BackupType = core.FBT_CONTENT
		item.Metrics.Size = size
	}
	return nil
}

// findDownNonMeasuredDirByWeight find children node (folder), which weight metric
// (weight here - number of all children folders) as close as possible to weight parameter
// passed to the function.
func findDownNonMeasuredDirByWeight(dir *core.Dir, weight int) *core.Dir {
	if !dir.Metrics.Measured &&
		(dir.Metrics.ChildrenCount <= weight ||
			dir.Metrics.ChildrenCount > weight && len(dir.Childs) == 0) {
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

// MeasureDir is a main heuristic function for planing backup process. MeasureDir walks through
// directory tree taken from dir variable, and find size which match size criteria optimal for splitting
// backup process to the pieces. As a result folders marked with corresponding type of processing,
// like core.FBT_RECURSIVE, core.FBT_CONTENT or core.FBT_SKIP, which lately used in backup stage
// as a direct instruction what to do. Returning totalCount contains statistics how many times
// application call RSYNC utility to measure folder size on remote server (with all content).
func MeasureDir(ctx context.Context, password *string, dir *core.Dir, retryCount *int,
	rsyncProtocol string, log *rsync.Logging, blockSize *backupBlockSizeSettings) (int, error) {

	totalCount := 0
	for {
		found, count, err := searchDownOptimalDir(ctx, password, dir, retryCount, rsyncProtocol, log, blockSize)
		if err != nil {
			return 0, err
		}
		totalCount += count
		if found == nil {
			break
		}

		if found.Metrics.IgnoreToBackup {
			LocalLog.Debugf("Selected for skip (count=%v): %v", count, found.Paths.RsyncSourcePath)
			// Mark this folder as "skip to backup" (because it contains special signature file).
			found.Metrics.BackupType = core.FBT_SKIP
		} else {
			LocalLog.Debugf("Selected for full backup (count=%v): %v", count, found.Paths.RsyncSourcePath)
			// Mark this folder as "recursive backup", when this folder and all included content and subfoders
			// are backing up in single RSYNC call.
			found.Metrics.BackupType = core.FBT_RECURSIVE
		}

		markMesuredAll(found)
		err = measureLocalUpToRoot(ctx, password, found, retryCount, rsyncProtocol, log)
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
func calcFullSizesWithRoot(ctx context.Context, password *string, dir *core.Dir,
	retryCount *int, rsyncProtocol string, log *rsync.Logging) (int, error) {

	count := 0
	root := getRoot(dir)
	if root.Metrics.FullSize == nil {
		fullSize, err := rsync.ObtainDirFullSize(ctx, password, root, retryCount, rsyncProtocol, log)
		if err != nil {
			return 0, err
		}
		root.Metrics.FullSize = fullSize
		count++
	}
	if dir.Metrics.FullSize == nil {
		fullSize, err := rsync.ObtainDirFullSize(ctx, password, dir, retryCount, rsyncProtocol, log)
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
	if bs > 5*core.GB {
		bs = 5 * core.GB
	} else if bs < 300*core.MB {
		bs = 300 * core.MB
	}
	return bs
}

// searchDownOptimalDir is a main recurrent function to find optimal (or close to optimal)
// walk path of backup source directory tree minimizing number of RSYNC utility calls.
func searchDownOptimalDir(ctx context.Context, password *string, dir *core.Dir, retryCount *int,
	rsyncProtocol string, log *rsync.Logging, blockSize *backupBlockSizeSettings) (*core.Dir, int, error) {

	LocalLog.Debugf("Start searching optimal folder from root %v",
		dir.Paths.RsyncSourcePath)

	found := getNonMeasuredDir(dir)

	if found != nil {
		LocalLog.Debugf("Get non-measured candidate %v to test",
			found.Paths.RsyncSourcePath)
	}

	totalFullSizeCount := 0
	if found != nil {
		count, err := calcFullSizesWithRoot(ctx, password, found, retryCount, rsyncProtocol, log)
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
			// employ "bisection method" to select next candidate
			weight := root.Metrics.ChildrenCount / 2
			next := findDownNonMeasuredDirByWeight(found, weight)
			if next == found {
				return next, totalFullSizeCount, nil
			} else {
				count, err := calcFullSizesWithRoot(ctx, password, next, retryCount,
					rsyncProtocol, log)
				if err != nil {
					return nil, 0, err
				}
				totalFullSizeCount += count

				if next.Metrics.FullSize.GetByteCount() > blockSize.BackupBlockSize {
					next, count, err = searchDownOptimalDir(ctx, password, next, retryCount,
						rsyncProtocol, log, blockSize)
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

			LocalLog.Debugf("Get dir by depth %v starting from %q", depth,
				found.Paths.RsyncSourcePath)

			next := findDownNonMeasuredDirByDepth(found, depth)
			count, err := calcFullSizesWithRoot(ctx, password, next, retryCount, rsyncProtocol, log)
			if err != nil {
				return nil, 0, err
			}
			totalFullSizeCount += count
			if next.Metrics.FullSize.GetByteCount() > blockSize.BackupBlockSize && len(next.Childs) > 0 {
				next = selectChildByWeight(next)
				next, count, err = searchDownOptimalDir(ctx, password, next, retryCount, rsyncProtocol,
					log, blockSize)
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
