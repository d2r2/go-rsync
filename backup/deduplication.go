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
)

type NodeSignature struct {
	SourceRsyncCipher string
	DestSubPath       string
}

func GetSignature(node BackupNode) NodeSignature {
	sha := ChipherStr(normalizeRsync(node.SourceRsync))
	signature := NodeSignature{SourceRsyncCipher: sha, DestSubPath: node.DestSubPath}
	// lg.Debug(sha)
	return signature
}

// ChipherStr encode str with SHA256.
func ChipherStr(str string) string {
	hasher := sha256.New()
	var b bytes.Buffer
	b.WriteString(str)
	hasher.Write(b.Bytes())
	sha := base64.URLEncoding.EncodeToString(hasher.Sum(nil))
	return sha
}

// normalizeRsync remove excess divider in rsync path.
func normalizeRsync(source string) string {
	chars := []rune(source)
	for i := len(chars) - 1; i >= 0; i-- {
		if chars[i] != '/' {
			newChars := chars[:i+1]
			return string(newChars)
		}
	}
	return ""
}

type NodeSignatures struct {
	Signatures []NodeSignature
}

func GetNodeSignatures(config *Config) NodeSignatures {
	signatures := make([]NodeSignature, len(config.BackupNodes))
	for i, item := range config.BackupNodes {
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

// PrevBackup describe found previous backup, which contain same Rsync source.
// Such previous backups used for Rsync utility deduplication, which
// significantly decrease size and time for repeated backup sessions.
type PrevBackup struct {
	// Full path to signature file name
	SignatureFileName string
	Signature         NodeSignature
}

func (v PrevBackup) GetDirPath() string {
	backupPath := path.Join(path.Dir(v.SignatureFileName), v.Signature.DestSubPath)
	return backupPath
}

// PrevBackups keeps list of previous backup found. See description of PrevBackup.
type PrevBackups struct {
	Backups []PrevBackup
}

func (v *PrevBackups) GetDirPaths() []string {
	var paths []string
	for _, b := range v.Backups {
		paths = append(paths, b.GetDirPath())
	}
	return paths
}

func FindPrevBackupPathsByNodeSignatures(lg logger.PackageLog, destPath string,
	signs NodeSignatures, lastN int) (*PrevBackups, error) {

	items, err := ioutil.ReadDir(destPath)
	if err != nil {
		return nil, err
	}
	candidates := make(map[string][]struct {
		time   time.Time
		backup PrevBackup
	})

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
						candidates[item1.SourceRsyncCipher] = append(candidates[item1.SourceRsyncCipher], struct {
							time   time.Time
							backup PrevBackup
						}{time: stat.ModTime(), backup: backup})
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, err
			}
		}
	}

	candidates2 := make(map[string][]struct {
		time   time.Time
		backup PrevBackup
	})
	for k, v := range candidates {
		sorted := FilesSortedByDate{Files: v}
		sort.Sort(sorted)
		candidates2[k] = sorted.Files
	}

	var backups []PrevBackup
	for i := 0; i < lastN; i++ {
		for _, v := range candidates2 {
			if i < len(v) {
				backups = append(backups, v[i].backup)
			}
		}
	}

	backups2 := &PrevBackups{Backups: backups}
	return backups2, nil
}

type FilesSortedByDate struct {
	Files []struct {
		time   time.Time
		backup PrevBackup
	}
}

func (s FilesSortedByDate) Len() int {
	return len(s.Files)
}

func (s FilesSortedByDate) Less(i, j int) bool {
	return s.Files[i].time.After(s.Files[j].time)
}

func (s FilesSortedByDate) Swap(i, j int) {
	node := s.Files[i]
	s.Files[i] = s.Files[j]
	s.Files[j] = node
}

func CreateMetadataSignatureFile(config *Config, destPath string) error {
	signs := GetNodeSignatures(config)
	err := os.MkdirAll(destPath, 0777)
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
