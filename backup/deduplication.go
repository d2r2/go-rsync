package backup

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"time"

	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/go-rsync/rsync"
)

// NodeSignature keep RSYNC source path
// crypted with hash function and destination subpath.
// RSYNC source path crypted with hash function
// is used as "source identifier" to search repeated
// backup sessions to use for deduplication.
// Content of this object is serialized to the file
// stored in backup session root folder.
type NodeSignature struct {
	SourceRsyncCipher string
	DestSubPath       string
}

// GetSignature builds NodeSignature object on the basis of BackupNodePath data.
func GetSignature(module Module) NodeSignature {
	signature := NodeSignature{SourceRsyncCipher: GenerateSourceID(module.SourceRsync),
		DestSubPath: module.DestSubPath}
	return signature
}

// GenerateSourceID convert RSYNC source URL to unique identifier.
func GenerateSourceID(rsyncSource string) string {
	return chipherStr(rsync.NormalizeRsyncURL(rsyncSource))
}

// chipherStr encode str with SHA256 hash function.
// Used to encode RSYNC source path before file serialization.
func chipherStr(str string) string {
	hasher := sha256.New()
	var b bytes.Buffer
	b.WriteString(str)
	hasher.Write(b.Bytes())
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// NodeSignatures keeps list of RSYNC source to backup in one session.
type NodeSignatures struct {
	Signatures []NodeSignature
}

func GetNodeSignatures(modules []Module) NodeSignatures {
	signatures := make([]NodeSignature, len(modules))
	for i, item := range modules {
		signatures[i] = GetSignature(item)
	}
	s := NodeSignatures{Signatures: signatures}
	return s
}

func (v NodeSignatures) FindFirstSignature(signature string) *NodeSignature {
	for _, item := range v.Signatures {
		if item.SourceRsyncCipher == signature {
			return &item
		}
	}
	return nil
}

// PrevBackup describe previous backup found, which contain same RSYNC source.
// Such previous backups used for RSYNC utility deduplication, which
// significantly decrease size and time for new backup session.
type PrevBackup struct {
	// Full path to signature file name
	SignatureFileName string
	Signature         NodeSignature
}

// GetDirPath returns full path to data copied in previous successful backup session.
func (v PrevBackup) GetDirPath() string {
	backupPath := path.Join(path.Dir(v.SignatureFileName), v.Signature.DestSubPath)
	return backupPath
}

// PrevBackups keeps list of previous backup found. See description of PrevBackup.
type PrevBackups struct {
	Backups []PrevBackup
}

// GetDirPaths provide file system paths to previous backup sessions found.
func (v *PrevBackups) GetDirPaths() []string {
	paths := make([]string, len(v.Backups))
	for i, b := range v.Backups {
		paths[i] = b.GetDirPath()
	}
	return paths
}

// FilterBySourceID choose backup sessions which contains same source
// as specified by sourceID.
func (v *PrevBackups) FilterBySourceID(sourceID string) *PrevBackups {
	var newPrevBackups []PrevBackup
	for _, v := range v.Backups {
		if sourceID == v.Signature.SourceRsyncCipher {
			newPrevBackups = append(newPrevBackups, v)
		}
	}
	return &PrevBackups{Backups: newPrevBackups}
}

type prevBackupEntry struct {
	time   time.Time
	backup PrevBackup
}

// FindPrevBackupPathsByNodeSignatures search for previous backup sessions which
// might significantly decrease backup size and speed up process.
// In the end it should return list of previous backup sessions sorted by date/time
// in descending order (recent go first).
func FindPrevBackupPathsByNodeSignatures(lg logger.PackageLog, destPath string,
	signs NodeSignatures, lastN int) (*PrevBackups, error) {

	// select all child items from root backup destination path
	items, err := ioutil.ReadDir(destPath)
	if err != nil {
		return nil, err
	}

	candidates := make(map[string][]prevBackupEntry)

	// loop through child folders to identify them as a previous backup sessions
	for _, item := range items {
		if item.IsDir() {
			fileName := filepath.Join(destPath, item.Name(), GetMetadataSignatureFileName())
			stat, err := os.Stat(fileName)
			if err != nil {
				if !os.IsNotExist(err) {
					if os.IsPermission(err) {
						lg.Notify(locale.T(MsgLogBackupStagePreviousBackupDiscoveryPermissionError,
							struct{ Path string }{Path: item.Name()}))
					} else {
						lg.Notify(locale.T(MsgLogBackupStagePreviousBackupDiscoveryOtherError,
							struct {
								Path  string
								Error error
							}{Path: item.Name(), Error: err}))
					}
				}
				continue
			}

			file, err := os.Open(fileName)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				signs2, err := DecodeSignatures(scanner.Text())
				if err != nil {
					break
				}
				for _, item1 := range signs.Signatures {
					if candidate := signs2.FindFirstSignature(item1.SourceRsyncCipher); candidate != nil {
						backup := PrevBackup{SignatureFileName: fileName, Signature: *candidate}
						candidates[item1.SourceRsyncCipher] = append(candidates[item1.SourceRsyncCipher],
							prevBackupEntry{time: stat.ModTime(), backup: backup})
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, err
			}
		}
	}

	// sort all candidates found by creation/modification time, to select the most resent previous backup sessions
	candidates2 := make(map[string][]prevBackupEntry)
	for k, v := range candidates {
		// sort previous backup sessions in descending order (the most recent come first)
		sorted := filesSortedByDate{Files: v}
		sort.Sort(sorted)
		maxPrevSessions := lastN
		// extra protection: according to limitation which exist in RSYNC,
		// no more than 20 --link-dest options could be provided with CLI, otherwise
		// RSYNC call failed (syntax or usage error, code 1) thrown;
		// maximum number of --link-dest option in single RSYNC call (detected experimentally)
		const maxLinkDest = 20
		// if still exceed, cut down
		if maxPrevSessions > maxLinkDest {
			maxPrevSessions = maxLinkDest
		}
		if len(sorted.Files) > maxPrevSessions {
			// cut to maxPrevSessions maximum
			sorted.Files = sorted.Files[:maxPrevSessions]
		}
		candidates2[k] = sorted.Files
	}

	var backups []PrevBackup
	for _, v := range candidates2 {
		for _, v2 := range v {
			backups = append(backups, v2.backup)
		}
	}

	backups2 := &PrevBackups{Backups: backups}
	return backups2, nil
}

// Temporary object used to sort found previous backup sessions by creation/modification date
// in descending order (the most recent come first).
type filesSortedByDate struct {
	Files []prevBackupEntry
}

func (s filesSortedByDate) Len() int {
	return len(s.Files)
}

func (s filesSortedByDate) Less(i, j int) bool {
	return s.Files[i].time.After(s.Files[j].time)
}

func (s filesSortedByDate) Swap(i, j int) {
	node := s.Files[i]
	s.Files[i] = s.Files[j]
	s.Files[j] = node
}

// CreateMetadataSignatureFile serialize RSYNC sources plus destination subpaths
// to the special "backup session signature" file.
func CreateMetadataSignatureFile(modules []Module, destPath string) error {
	signs := GetNodeSignatures(modules)
	err := createDirAll(destPath)
	if err != nil {
		return err
	}
	destPath = filepath.Join(destPath, GetMetadataSignatureFileName())
	file, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer file.Close()
	v, err := EncodeSignatures(signs)
	if err != nil {
		return err
	}
	_, err = file.WriteString(v)
	if err != nil {
		return err
	}
	return nil
}

// EncodeSignatures encode NodeSignatures object to self-describing binary format.
func EncodeSignatures(signs NodeSignatures) (string, error) {
	b := bytes.Buffer{}
	e := gob.NewEncoder(&b)
	err := e.Encode(signs)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// DecodeSignatures decode NodeSignatures object from self-describing binary format.
func DecodeSignatures(str string) (*NodeSignatures, error) {
	m := &NodeSignatures{}
	by, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	b := bytes.Buffer{}
	b.Write(by)
	d := gob.NewDecoder(&b)
	err = d.Decode(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}
