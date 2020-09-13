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
	"fmt"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/rsync"
)

/*
type IRsyncConfigurable interface {
	GetRsyncParams(addExtraParams []string) []string
}
*/

// Node contain information about single RSYNC source backup.
type Node struct {
	Module  Module
	RootDir *core.Dir
}

// Plan keep all necessary information obtained from
// preferences and 1st backup pass to start backup process.
type Plan struct {
	Config     *Config
	Nodes      []Node
	BackupSize core.FolderSize
}

// GetModules returns all RSYNC source/destination blocks
// defined in single (specific) backup profile.
func (v *Plan) GetModules() []Module {
	modules := []Module{}
	for _, item := range v.Nodes {
		modules = append(modules, item.Module)
	}
	return modules
}

// Config keeps backup session configuration.
// Config instance is initialized mainly from
// GLIB GSettings in ui/gtkui package.
type Config struct {
	SigFileIgnoreBackup                string `toml:"sig_file_ignore_backup"`
	RsyncRetryCount                    *int   `toml:"retry_count"`
	AutoManageBackupBlockSize          *bool  `toml:"auto_manage_backup_block_size"`
	MaxBackupBlockSizeMb               *int   `toml:"max_backup_block_size_mb"`
	UsePreviousBackup                  *bool  `toml:"use_previous_backup"`
	NumberOfPreviousBackupToUse        *int   `toml:"number_of_previous_backup_to_use"`
	EnableLowLevelLogForRsync          *bool  `toml:"enable_low_level_log_rsync"`
	EnableIntensiveLowLevelLogForRsync *bool  `toml:"enable_intensive_low_level_log_rsync"`

	RsyncTransferSourceOwner       *bool `toml:"rsync_transfer_source_owner"`       // rsync --owner
	RsyncTransferSourceGroup       *bool `toml:"rsync_transfer_source_group"`       // rsync --group
	RsyncTransferSourcePermissions *bool `toml:"rsync_transfer_source_permissions"` // rsync --perms
	RsyncRecreateSymlinks          *bool `toml:"rsync_recreate_symlinks"`           // rsync --links
	RsyncTransferDeviceFiles       *bool `toml:"rsync_transfer_device_files"`       // rsync --devices
	RsyncTransferSpecialFiles      *bool `toml:"rsync_transfer_special_files"`      // rsync --specials
	RsyncCompressFileTransfer      *bool `toml:"rsync_compress_file_transfer"`      // rsync --compress

	// BackupNode list contain all RSYNC sources to backup in one session.
	//Modules []Module `toml:"backup_module"`
}

/*
func NewConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	LocalLog.Debug(f("%+v", config))
	return &config, nil
}
*/

func (conf *Config) usePreviousBackupEnabled() bool {
	var usePreviousBackup = true
	if conf.UsePreviousBackup != nil {
		usePreviousBackup = *conf.UsePreviousBackup
	}
	return usePreviousBackup
}

func (conf *Config) numberOfPreviousBackupToUse() int {
	var numberOfPreviousBackupToUse = 1
	if conf.NumberOfPreviousBackupToUse != nil {
		numberOfPreviousBackupToUse = *conf.NumberOfPreviousBackupToUse
	}
	return numberOfPreviousBackupToUse
}

func (conf *Config) getRsyncLoggingSettings() *rsync.Logging {
	logging := &rsync.Logging{}
	if conf.EnableLowLevelLogForRsync != nil {
		logging.EnableLog = *conf.EnableLowLevelLogForRsync
	}
	if conf.EnableIntensiveLowLevelLogForRsync != nil {
		logging.EnableIntensiveLog = *conf.EnableIntensiveLowLevelLogForRsync
	}
	return logging
}

func (conf *Config) getBackupBlockSizeSettings() *backupBlockSizeSettings {
	blockSize := &backupBlockSizeSettings{AutoManageBackupBlockSize: true, BackupBlockSize: 500}
	if conf.AutoManageBackupBlockSize != nil {
		blockSize.AutoManageBackupBlockSize = *conf.AutoManageBackupBlockSize
	}
	if conf.MaxBackupBlockSizeMb != nil {
		blockSize.BackupBlockSize = uint64(*conf.MaxBackupBlockSizeMb * 1024 * 1024)
	}
	return blockSize
}

// Module signify RSYNC source/destination block, with
// source/destination URLs and other auxiliary options.
// Used as configuration data in the backup session code.
type Module struct {
	SourceRsync string `toml:"src_rsync"`
	DestSubPath string `toml:"dst_subpath"`

	ChangeFilePermission string  `toml:"rsync_change_file_permission"`
	AuthPassword         *string `toml:"module_auth_password"`

	RsyncTransferSourceOwner       *bool `toml:"rsync_transfer_source_owner"`       // rsync --owner
	RsyncTransferSourceGroup       *bool `toml:"rsync_transfer_source_group"`       // rsync --group
	RsyncTransferSourcePermissions *bool `toml:"rsync_transfer_source_permissions"` // rsync --perms
	RsyncRecreateSymlinks          *bool `toml:"rsync_recreate_symlinks"`           // rsync --links
	RsyncTransferDeviceFiles       *bool `toml:"rsync_transfer_device_files"`       // rsync --devices
	RsyncTransferSpecialFiles      *bool `toml:"rsync_transfer_special_files"`      // rsync --specials
}

// GetRsyncParams prepare RSYNC CLI parameters to run console RSYNC process.
func GetRsyncParams(conf *Config, module *Module, addExtraParams []string) []string {
	var params []string
	if module.RsyncTransferSourceOwner != nil && *module.RsyncTransferSourceOwner ||
		module.RsyncTransferSourceOwner == nil && conf.RsyncTransferSourceOwner != nil &&
			*conf.RsyncTransferSourceOwner {
		params = append(params, "--owner")
	}
	if module.RsyncTransferSourceGroup != nil && *module.RsyncTransferSourceGroup ||
		module.RsyncTransferSourceGroup == nil && conf.RsyncTransferSourceGroup != nil &&
			*conf.RsyncTransferSourceGroup {
		params = append(params, "--group")
	}
	if module.RsyncTransferSourcePermissions != nil && *module.RsyncTransferSourcePermissions ||
		module.RsyncTransferSourcePermissions == nil && conf.RsyncTransferSourcePermissions != nil &&
			*conf.RsyncTransferSourcePermissions {
		params = append(params, "--perms")
	}
	if module.RsyncRecreateSymlinks != nil && *module.RsyncRecreateSymlinks ||
		module.RsyncRecreateSymlinks == nil && conf.RsyncRecreateSymlinks != nil &&
			*conf.RsyncRecreateSymlinks {
		params = append(params, "--links")
	}
	if module.RsyncTransferDeviceFiles != nil && *module.RsyncTransferDeviceFiles ||
		module.RsyncTransferDeviceFiles == nil && conf.RsyncTransferDeviceFiles != nil &&
			*conf.RsyncTransferDeviceFiles {
		params = append(params, "--devices")
	}
	if module.RsyncTransferSpecialFiles != nil && *module.RsyncTransferSpecialFiles ||
		module.RsyncTransferSpecialFiles == nil && conf.RsyncTransferSpecialFiles != nil &&
			*conf.RsyncTransferSpecialFiles {
		params = append(params, "--specials")
	}
	if conf.RsyncCompressFileTransfer != nil && *conf.RsyncCompressFileTransfer {
		params = append(params, "--compress")
	}
	if module.ChangeFilePermission != "" {
		params = append(params, fmt.Sprintf("--chmod=%s", module.ChangeFilePermission))
	}

	params = append(params, addExtraParams...)
	return params
}
