package backup

import (
	"io/ioutil"
	"os"
	"path"

	shell "github.com/d2r2/go-shell"
)

type LogFiles struct {
	rootPath string
	logs     map[string]*os.File
}

func NewLogFiles() *LogFiles {
	v := &LogFiles{logs: make(map[string]*os.File)}
	return v
}

func (v *LogFiles) GetAppendFile(suffixPath string) (*os.File, error) {
	err := v.assignRootPathByDefault()
	if err != nil {
		return nil, err
	}
	file := v.logs[suffixPath]
	if file == nil {
		file, err = os.OpenFile(v.getFullPath(suffixPath), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
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
