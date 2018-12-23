package backup

import (
	"github.com/BurntSushi/toml"
	"github.com/d2r2/go-rsync/rsync"
)

type Config struct {
	SigFileIgnoreBackup                string `toml:"sig_file_ignore_backup"`
	RsyncRetryCount                    *int   `toml:"retry_count"`
	AutoManageBackupBlockSize          *bool  `auto_manage_backup_block_size`
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

	BackupNodes []BackupNode `toml:"backup_node"`
}

func NewConfig(filePath string) (*Config, error) {
	var config Config
	if _, err := toml.DecodeFile(filePath, &config); err != nil {
		return nil, err
	}
	LocalLog.Debug(f("%+v", config))
	return &config, nil
}

func (conf *Config) getRsyncParams(addExtraParams ...string) []string {
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

type sortConfig struct {
	BackupNodes []BackupNode
}

func (s sortConfig) Len() int {
	return len(s.BackupNodes)
}

func (s sortConfig) Less(i, j int) bool {
	if s.BackupNodes[i].SourceRsync < s.BackupNodes[j].SourceRsync &&
		s.BackupNodes[i].DestSubPath < s.BackupNodes[j].DestSubPath {
		return true
	} else {
		return false
	}
}

func (s sortConfig) Swap(i, j int) {
	node := s.BackupNodes[i]
	s.BackupNodes[i] = s.BackupNodes[j]
	s.BackupNodes[j] = node
}
