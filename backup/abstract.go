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
	"time"

	"github.com/d2r2/go-rsync/core"
)

// Notifier interface is used as a contract to provide
// event-driven mechanism, to map backup process steps with,
// for instance, user interface.
type Notifier interface {

	// Pair of calls to report about 1st pass start and completion.
	NotifyPlanStage_NodeStructureStartInquiry(sourceID int,
		sourceRsync string) error
	NotifyPlanStage_NodeStructureDoneInquiry(sourceID int,
		sourceRsync string, dir *core.Dir) error

	// Pair of calls to report about 2nd pass start and completion.
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
