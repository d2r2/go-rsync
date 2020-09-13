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

package core

import (
	"bytes"
	"path"
	"path/filepath"
	"strings"
)

// SrcDstPath link to each other RSYNC source URL
// with destination extra path added to backup folder.
type SrcDstPath struct {
	RsyncSourcePath string
	DestPath        string
}

// Join fork SrcDstPath with new variant, where
// new "folder" appended to the end of the path.
func (v SrcDstPath) Join(item string) SrcDstPath {
	newPathTwin := SrcDstPath{
		RsyncSourcePath: RsyncPathJoin(v.RsyncSourcePath, item),
		DestPath:        filepath.Join(v.DestPath, item)}
	return newPathTwin
}

// RsyncPathJoin used to join path elements in RSYNC url.
func RsyncPathJoin(elements ...string) string {
	// use standard URL separator
	const separator = '/'
	var buf bytes.Buffer
	for _, item := range elements {
		buf.WriteString(item)
		if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != separator {
			buf.WriteByte(separator)
		}
	}
	return buf.String()
}

// GetRelativePath cut off root prefix from destPath (if found).
func GetRelativePath(rootDest, destPath string) (string, error) {
	rel, err := filepath.Rel(rootDest, destPath)
	if err != nil {
		return "", err
	}
	rel = "." + strings.Trim((path.Join(" ", rel, " ")), " ")
	return rel, nil
}

// GetRelativePaths cut off root prefix from multiple paths (if found).
func GetRelativePaths(rootDest string, paths []string) ([]string, error) {
	var newPaths []string
	for _, p := range paths {
		np, err := GetRelativePath(rootDest, p)
		if err != nil {
			return nil, err
		}
		newPaths = append(newPaths, np)
	}
	return newPaths, nil
}
