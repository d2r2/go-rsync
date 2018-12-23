package rsync

import (
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
)

func ObtainDirLocalSize(ctx context.Context, dir *core.Dir, retryCount *int, log *Logging) (*core.FolderSize, error) {
	// rsync "dry run" to get total size of backup
	var stdOut bytes.Buffer
	options := NewOptions(WithDefaultParams("--dry-run", "--compress")).
		AddParams("--dirs").
		SetRetryCount(retryCount)
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, log, &stdOut, dir.Paths)
	if sessionErr != nil {
		return nil, sessionErr
	}
	backupSize, err := extractBackupSize(&stdOut)
	if err != nil {
		return nil, err
	}
	if backupSize != nil {
		lg.Debugf("Get rsync %q size: %v", dir.Paths.RsyncSourcePath,
			core.GetReadableSize(*backupSize))
	}
	return backupSize, nil
}

func ObtainDirFullSize(ctx context.Context, dir *core.Dir, retryCount *int, log *Logging) (*core.FolderSize, error) {
	// rsync "dry run" to get total size of backup
	var stdOut bytes.Buffer
	options := NewOptions(WithDefaultParams("--dry-run", "--compress")).
		AddParams("--recursive", f("--include=%s", "*"+"/")).
		SetRetryCount(retryCount)
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, log, &stdOut, dir.Paths)
	if sessionErr != nil {
		return nil, sessionErr
	}
	backupSize, err := extractBackupSize(&stdOut)
	if err != nil {
		return nil, err
	}
	return backupSize, nil
}

// extractBackupSize parse and decode Rsync console output
// to obtain folder content size.
func extractBackupSize(stdOut *bytes.Buffer) (*core.FolderSize, error) {
	// Parse the line: "total size is 2,227,810,354  speedup is 507,127.33 (DRY RUN)"
	// to extract "total size" value.
	re := regexp.MustCompile(`total\s+size\s+is\s+(?P<Number>((\d+)\,?)+)`)
	str := stdOut.String()
	m := core.FindStringSubmatchIndexes(re, str)
	if a, ok := m["Number"]; ok {
		str2 := strings.Replace(str[a[0]:a[1]], ",", "", -1)
		// lg.Debugf("%v", str2)
		i, err := strconv.ParseInt(str2, 10, 64)
		if err != nil {
			return nil, errors.New(locale.T(MsgRsyncCannotParseFolderSizeOutputError,
				struct{ Text string }{Text: str2}))
		}
		i2 := core.FolderSize(i)
		return &i2, nil
	} else {
		return nil, errors.New(locale.T(MsgRsyncCannotFindFolderSizeOutputError, nil))
	}
}

func GetPathStatus(ctx context.Context, sourceRSync string, recursive bool) error {
	tempDir, err := ioutil.TempDir("", "backup_dir_status_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(sourceRSync, ""),
		DestPath:        tempDir,
	}
	// rsync "dry run" to get total size of backup
	var stdOut bytes.Buffer
	options := NewOptions(WithDefaultParams("--include=*/", "--dry-run"))
	if recursive {
		options.AddParams("--recursive")
	}
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, nil, &stdOut, paths)
	if sessionErr != nil {
		return sessionErr
	}
	return nil
}
