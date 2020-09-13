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

package gtkui

const (
	APP_SCHEMA_ID              = "org.d2r2.gorsync"
	SETTINGS_SCHEMA_ID         = APP_SCHEMA_ID + "." + "Settings"
	SETTINGS_SCHEMA_PATH       = "/org/d2r2/gorsync/"
	PROFILE_SCHEMA_SUFFIX_ID   = "Profile"
	PROFILE_SCHEMA_SUFFIX_PATH = "profiles/%s"
	SOURCE_SCHEMA_SUFFIX_ID    = "Source"
	SOURCE_SCHEMA_SUFFIX_PATH  = "sources/%s"
)

const (
	CFG_IGNORE_FILE_SIGNATURE                          = "ignore-file-signature"
	CFG_RSYNC_RETRY_COUNT                              = "rsync-retry-count"
	CFG_MANAGE_AUTO_BACKUP_BLOCK_SIZE                  = "manage-automatically-backup-block-size"
	CFG_MAX_BACKUP_BLOCK_SIZE_MB                       = "max-backup-block-size-mb"
	CFG_ENABLE_USE_OF_PREVIOUS_BACKUP                  = "enable-use-of-previous-backup"
	CFG_NUMBER_OF_PREVIOUS_BACKUP_TO_USE               = "number-of-previous-backup-to-use"
	CFG_ENABLE_LOW_LEVEL_LOG_OF_RSYNC                  = "enable-low-level-log-for-rsync"
	CFG_ENABLE_INTENSIVE_LOW_LEVEL_LOG_OF_RSYNC        = "enable-intensive-low-level-log-for-rsync"
	CFG_RSYNC_TRANSFER_SOURCE_GROUP_INCONSISTENT       = "rsync-transfer-source-group-inconsistent"
	CFG_RSYNC_TRANSFER_SOURCE_GROUP                    = "rsync-transfer-source-group"
	CFG_RSYNC_TRANSFER_SOURCE_OWNER_INCONSISTENT       = "rsync-transfer-source-owner-inconsistent"
	CFG_RSYNC_TRANSFER_SOURCE_OWNER                    = "rsync-transfer-source-owner"
	CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS_INCONSISTENT = "rsync-transfer-source-permissions-inconsistent"
	CFG_RSYNC_TRANSFER_SOURCE_PERMISSIONS              = "rsync-transfer-source-permissions"
	CFG_RSYNC_RECREATE_SYMLINKS_INCONSISTENT           = "rsync-recreate-symlinks-inconsistent"
	CFG_RSYNC_RECREATE_SYMLINKS                        = "rsync-recreate-symlinks"
	CFG_RSYNC_TRANSFER_DEVICE_FILES_INCONSISTENT       = "rsync-transfer-device-files-inconsistent"
	CFG_RSYNC_TRANSFER_DEVICE_FILES                    = "rsync-transfer-device-files"
	CFG_RSYNC_TRANSFER_SPECIAL_FILES_INCONSISTENT      = "rsync-transfer-special-files-inconsistent"
	CFG_RSYNC_TRANSFER_SPECIAL_FILES                   = "rsync-transfer-special-files"
	CFG_RSYNC_COMPRESS_FILE_TRANSFER                   = "rsync-compress-file-transfer"
	CFG_BACKUP_LIST                                    = "profile-list"
	CFG_SOURCE_LIST                                    = "source-list"
	CFG_DONT_SHOW_ABOUT_ON_STARTUP                     = "dont-show-about-dialog-on-startup"
	CFG_UI_LANGUAGE                                    = "ui-language"
	CFG_SESSION_LOG_WIDGET_FONT_SIZE                   = "session-log-widget-font-size"
	CFG_PROFILE_NAME                                   = "profile-name"
	CFG_PROFILE_DEST_ROOT_PATH                         = "destination-root-path"
	CFG_MODULE_RSYNC_SOURCE_PATH                       = "rsync-source-path"
	CFG_MODULE_DEST_SUBPATH                            = "dest-subpath"
	CFG_MODULE_CHANGE_FILE_PERMISSION                  = "change-file-permission"
	CFG_MODULE_AUTH_PASSWORD                           = "auth-password"
	CFG_MODULE_ENABLED                                 = "source-dest-block-enabled"
	CFG_PERFORM_DESKTOP_NOTIFICATION                   = "perform-backup-completion-desktop-notification"
	CFG_RUN_NOTIFICATION_SCRIPT                        = "run-backup-completion-notification-script"
)
