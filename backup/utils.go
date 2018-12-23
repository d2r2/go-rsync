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

const (
	TAB_RUNE = '\t'
)

func createDirAll(path string) error {
	err := os.MkdirAll(path, 0777)
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

// func writeLog(log bytes.Buffer, destPath, fileName string) error {
// 	err := os.MkdirAll(destPath, 0777)
// 	if err != nil {
// 		return err
// 	}
// 	destPath = filepath.Join(destPath, fileName)
// 	file, err := os.Create(destPath)
// 	if err != nil {
// 		return err
// 	}
// 	defer file.Close()
// 	_, err = file.Write(log.Bytes())
// 	return err
// }

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

// func TranslateRsyncError(err error) error {
// 	if err2, ok := err.(*rsync.RsyncError); ok {
// 		// translate RsyncError to local language
// 		err = errors.New(locale.T(MsgRsyncCallFailedError,
// 			struct {
// 				Description string
// 				ExitCode    int
// 			}{Description: err2.Description, ExitCode: err2.ExitCode}))
// 	}
// 	return err
// }

// GetBackupFolderName return new folder name for ongoing backup process.
func GetBackupFolderName( /*backupType io.BackupType,*/
	incomplete bool, date *time.Time) string {

	var prefixPath string = "~rsync_backup"
	/*
		if backupType == io.BT_FULL {
			//		prefixPath = "~rsync_backup_full"
			prefixPath = "~rsync_backup"
		} else if backupType == io.BT_DIFF {
			//		prefixPath = "~rsync_backup_snap"
			prefixPath = "~rsync_backup"
		}
	*/
	if incomplete {
		prefixPath += "_(incomplete)"
	}
	var t time.Time = time.Now()
	if date != nil {
		t = *date
	}
	prefixPath += t.Format("~20060102-150405~")
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

// GetRsyncLogFileName return the name of specific low-level Rsync utility log.
func GetRsyncLogFileName() string {
	return "~rsync_log~.log"
}
