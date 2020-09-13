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

package rsync

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
)

// ObtainDirLocalSize parse STDOUT from RSYNC dry-run execution to extract local size of directory without nested folders.
func ObtainDirLocalSize(ctx context.Context, password *string, dir *core.Dir,
	retryCount *int, rsyncProtocol string, log *Logging) (*core.FolderSize, error) {

	// RSYNC "dry run" to get total size of backup
	var stdOut bytes.Buffer
	options := NewOptions(WithDefaultParams([]string{"--dry-run", "--compress"})).
		AddParams("--dirs").
		SetRetryCount(retryCount).
		SetAuthPassword(password)
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, log, &stdOut, dir.Paths)
	if sessionErr != nil {
		return nil, sessionErr
	}
	backupSize, err := extractBackupSize(&stdOut, rsyncProtocol)
	if err != nil {
		return nil, err
	}
	if backupSize != nil {
		lg.Debugf("Get rsync %q size: %v", dir.Paths.RsyncSourcePath,
			core.GetReadableSize(*backupSize))
	}
	return backupSize, nil
}

// ObtainDirLocalSize parse STDOUT from RSYNC dry-run execution to extract full size of directory.
func ObtainDirFullSize(ctx context.Context, password *string, dir *core.Dir,
	retryCount *int, rsyncProtocol string, log *Logging) (*core.FolderSize, error) {

	// RSYNC "dry run" to get total size of backup
	var stdOut bytes.Buffer
	options := NewOptions(WithDefaultParams([]string{"--dry-run", "--compress"})).
		AddParams("--recursive", "--include=*/").
		SetRetryCount(retryCount).
		SetAuthPassword(password)
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, log, &stdOut, dir.Paths)
	if sessionErr != nil {
		return nil, sessionErr
	}
	backupSize, err := extractBackupSize(&stdOut, rsyncProtocol)
	if err != nil {
		return nil, err
	}
	return backupSize, nil
}

// extractBackupSize parse and decode RSYNC STDOUT output to obtain folder content size.
func extractBackupSize(stdOut *bytes.Buffer, rsyncProtocol string) (*core.FolderSize, error) {
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

// GetPathStatus verify that RSYNC source path is valid.
// For this RSYNC is launched, than exit status is evaluated.
func GetPathStatus(ctx context.Context, password *string,
	sourceRSync string, recursive bool) error {

	tempDir, err := ioutil.TempDir("", "backup_dir_status_")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	paths := core.SrcDstPath{
		RsyncSourcePath: core.RsyncPathJoin(sourceRSync, ""),
		DestPath:        tempDir,
	}
	options := NewOptions(WithDefaultParams([]string{"--include=*/", "--dry-run"})).
		SetAuthPassword(password)
	if recursive {
		options.AddParams("--recursive")
	}
	sessionErr, _, _ := RunRsyncWithRetry(ctx, options, nil, nil, paths)
	if sessionErr != nil {
		return sessionErr
	}
	return nil
}

// NormalizeRsyncURL normalize RSYNC URL by:
// 1) remove user specification (if found).
// 2) remove excess '/' chars in path following host.
func NormalizeRsyncURL(rsyncURL string) string {
	_, host, path := parseRsyncURL(strings.TrimSpace(rsyncURL))
	path = removeExcessSlashChars(path)
	// assemble RSYNC URL path back, but without user specification
	newRsyncURL := fmt.Sprintf("rsync://%s%s", host, path)
	// lg.Debugf("Original RSYNC URL: %s", rsyncURL)
	// lg.Debugf("Modified RSYNC URL: %s", newRsyncURL)
	return newRsyncURL
}

// parseRsyncURL disassemble RSYNC URL to the parts.
// This parts include: rsync prefix, user (if specified), host and path.
func parseRsyncURL(rsyncURL string) (user, host, path string) {
	re := regexp.MustCompile(`(?i:^rsync://(?P<user>[^@]*@)?(?P<host>[^/]*)(?P<path>.*)$)`)
	m := core.FindStringSubmatchIndexes(re, rsyncURL)
	if len(m) > 0 {
		grUser := "user"
		if _, ok := m[grUser]; ok {
			start := m[grUser][0]
			end := m[grUser][1]
			user = rsyncURL[start:end]
		}
		grHost := "host"
		if _, ok := m[grHost]; ok {
			start := m[grHost][0]
			end := m[grHost][1]
			host = rsyncURL[start:end]
		}
		grPath := "path"
		if _, ok := m[grPath]; ok {
			start := m[grPath][0]
			end := m[grPath][1]
			path = rsyncURL[start:end]
		}
	}
	return
}

// removeExcessSlashChars remove excess path divider in RSYNC path.
func removeExcessSlashChars(path string) string {
	var buf bytes.Buffer
	lastCharIsSlash := false
	for _, ch := range path {
		if ch == '/' {
			if lastCharIsSlash {
				continue
			}
			lastCharIsSlash = true
		} else {
			lastCharIsSlash = false
		}
		buf.WriteRune(ch)
	}

	path = buf.String()
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	return path
}
