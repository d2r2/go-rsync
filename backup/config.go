package backup

import (
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/rsync"
)

type IRsyncConfigurable interface {
	GetRsyncParams(addExtraParams []string) []string
}

// Node contain information about single rsync source backup.
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
	// rsync --compress
	RsyncCompressFileTransfer *bool `toml:"rsync_compress_file_transfer"`
	// rsync --links
	RsyncRecreateSymlinks *bool `toml:"rsync_recreate_symlinks"`
	// rsync --perms
	RsyncTransferSourcePermissions *bool `toml:"rsync_transfer_source_permissions"`
	// rsync --group
	RsyncTransferSourceGroup *bool `toml:"rsync_transfer_source_group"`
	// rsync --owner
	RsyncTransferSourceOwner *bool `toml:"rsync_transfer_source_owner"`
	// rsync --devices
	RsyncTransferDeviceFiles *bool `toml:"rsync_transfer_device_files"`
	// rsync --specials
	RsyncTransferSpecialFiles *bool `toml:"rsync_transfer_special_files"`

	// BackupNode list contain all RSYNC sources to backup in one session.
	//Modules []Module `toml:"backup_module"`
}

func NewConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	LocalLog.Debug(f("%+v", config))
	return &config, nil
}

// Prepare RSYNC CLI parameters to run console RSYNC process.
func (conf *Config) GetRsyncParams(addExtraParams []string) []string {
	var params []string
	if conf.RsyncCompressFileTransfer != nil && *conf.RsyncCompressFileTransfer {
		params = append(params, "--compress")
	}
	if conf.RsyncTransferSourceOwner != nil && *conf.RsyncTransferSourceOwner {
		params = append(params, "--owner")
	}
	if conf.RsyncTransferSourceGroup != nil && *conf.RsyncTransferSourceGroup {
		params = append(params, "--group")
	}
	if conf.RsyncTransferSourcePermissions != nil && *conf.RsyncTransferSourcePermissions {
		params = append(params, "--perms")
	}
	if conf.RsyncRecreateSymlinks != nil && *conf.RsyncRecreateSymlinks {
		params = append(params, "--links")
	}
	if conf.RsyncTransferDeviceFiles != nil && *conf.RsyncTransferDeviceFiles {
		params = append(params, "--devices")
	}
	if conf.RsyncTransferSpecialFiles != nil && *conf.RsyncTransferSpecialFiles {
		params = append(params, "--specials")
	}
	params = append(params, addExtraParams...)
	return params
}

func (conf *Config) usePreviousBackupEnabled() bool {
	var usePreviousBackup bool = true
	if conf.UsePreviousBackup != nil {
		usePreviousBackup = *conf.UsePreviousBackup
	}
	return usePreviousBackup
}

func (conf *Config) numberOfPreviousBackupToUse() int {
	var numberOfPreviousBackupToUse int = 1
	if conf.NumberOfPreviousBackupToUse != nil {
		numberOfPreviousBackupToUse = *conf.NumberOfPreviousBackupToUse
	}
	return numberOfPreviousBackupToUse
}

func (conf *Config) getRsyncSettings() *rsync.Logging {
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

type Module struct {
	SourceRsync          string  `toml:"src_rsync"`
	DestSubPath          string  `toml:"dst_subpath"`
	ChangeFilePermission string  `toml:"rsync_change_file_permission"`
	AuthPassword         *string `toml:"module_auth_password"`
}

// Prepare RSYNC CLI parameters to run console RSYNC process.
func (module *Module) GetRsyncParams(addExtraParams []string) []string {
	var params []string
	if module.ChangeFilePermission != "" {
		params = append(params, fmt.Sprintf("--chmod=%s", module.ChangeFilePermission))
	}
	params = append(params, addExtraParams...)
	return params
}
