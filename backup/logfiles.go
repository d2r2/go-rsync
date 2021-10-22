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
	"io/ioutil"
	"os"
	"path"

	shell "github.com/d2r2/go-shell"
)

// LogFiles track log files during backup session.
// It has functionality to relocate log files from
// one storage to another: used when log files moved
// from /tmp partition to permanent destination location.
type LogFiles struct {
	rootPath string
	logs     map[string]*os.File
}

// NewLogFiles create new LogFiles instance.
func NewLogFiles() *LogFiles {
	v := &LogFiles{logs: make(map[string]*os.File)}
	return v
}

// CreateOrGetLogFile return os.File by file name identifier.
// This allow to control and operate multiple log files in one place.
func (v *LogFiles) CreateOrGetLogFile(suffixPath string) (*os.File, error) {
	err := v.assignRootPathByDefault()
	if err != nil {
		return nil, err
	}
	file := v.logs[suffixPath]
	if file == nil {
		file, err = os.OpenFile(v.getFullPath(suffixPath), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
		v.logs[suffixPath] = file
	}
	return file, nil
}

func (v *LogFiles) getFullPath(suffixPath string) string {
	return path.Join(v.rootPath, suffixPath)
}

// Close will close all os.File instances found in the object.
func (v *LogFiles) Close() error {
	for suffixPath, val := range v.logs {
		if val != nil {
			err := val.Close()
			if err != nil {
				return err
			}
			v.logs[suffixPath] = nil
		}
	}
	return nil
}

// ChangeRootPath relocate log files from one storage to another.
// Used to move from 1st backup stage (plan stage) to 2nd (backup stage).
// In 1st backup stage we keep log files in /tmp partition, in 2nd stage
// we relocate and save them in destination location.
func (v *LogFiles) ChangeRootPath(newRootPath string) error {
	err := v.Close()
	if err != nil {
		return err
	}
	if _, err = os.Stat(v.rootPath); !os.IsNotExist(err) {
		for suffixPath := range v.logs {
			oldpath := v.getFullPath(suffixPath)
			newpath := path.Join(newRootPath, suffixPath)
			_, err = shell.CopyFile(oldpath, newpath)
			if err != nil {
				return err
			}
		}
	}
	v.rootPath = newRootPath
	return nil
}

func (v *LogFiles) assignRootPathByDefault() error {
	if v.rootPath == "" {
		dir, err := ioutil.TempDir("", "gorsync_logs_")
		if err != nil {
			return err
		}
		v.rootPath = dir
	}
	return nil
}
