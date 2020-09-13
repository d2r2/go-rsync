//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2020 Denis Dyakov <denis.dyakov@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package backup

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
)

// TAB_RUNE keep tab character.
const TAB_RUNE = '\t'

func createDirAll(path string) error {
	err := os.MkdirAll(path, 0777)
	return err
}

func createDirInBackupStage(path string) error {
	err := createDirAll(path)
	if err != nil {
		err = errors.New(locale.T(MsgLogBackupStageFailedToCreateFolder,
			struct {
				Path  string
				Error error
			}{Path: path, Error: err}))
		return err
	}
	return nil
}

func splitToLines(buf *bytes.Buffer) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(buf)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func writeLineIndent(buf *bytes.Buffer, tabNumber int, text string) {
	for i := 0; i < tabNumber; i++ {
		buf.WriteRune(TAB_RUNE)
	}
	buf.WriteString(fmt.Sprintln(text))
}

// GetBackupTypeDescription return localized description of how
// application will backup specific directory described by core.Dir object.
// It could be 3 options:
// 1) full backup: backup full content include all nested folders;
// 2) flat backup: backup only direct files in folder, ignore nested folders;
// 3) skip backup: skip folder backup (this happens when specific signature file found).
func GetBackupTypeDescription(backupType core.FolderBackupType) string {
	var backupStr string
	switch backupType {
	case core.FBT_SKIP:
		backupStr = locale.T(MsgFolderBackupTypeSkipDescription, nil)
	case core.FBT_RECURSIVE:
		backupStr = locale.T(MsgFolderBackupTypeRecursiveDescription, nil)
	case core.FBT_CONTENT:
		backupStr = locale.T(MsgFolderBackupTypeContentDescription, nil)
	}
	return backupStr
}

// GetBackupFolderName return new folder name for ongoing backup process.
func GetBackupFolderName(incomplete bool, date *time.Time) string {
	prefixPath := "~rsync_backup"
	if incomplete {
		prefixPath += "_(incomplete)"
	}
	var dt time.Time = time.Now()
	if date != nil {
		dt = *date
	}
	prefixPath += dt.Format("~20060102-150405~")
	return prefixPath
}

// GetMetadataSignatureFileName return the name of specific file
// which describe all sources used in backup process.
func GetMetadataSignatureFileName() string {
	return "~backup_nodes~.signatures"
}

// GetLogFileName return the name of general backup process log.
func GetLogFileName() string {
	return "~backup_log~.log"
}

// GetRsyncLogFileName return the name of specific low-level RSYNC utility log.
func GetRsyncLogFileName() string {
	return "~rsync_log~.log"
}
