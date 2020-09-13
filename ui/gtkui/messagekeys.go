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

// ------------------------------------------------------------
// File contains message identifiers for localization purpose.
// Message identifier names is self-descriptive, so ordinary
// it's easy to understand what message is made for.
// Message ID is used to call translation functions from
// "locale" package.
// ------------------------------------------------------------

const (
	MsgAppEnvironmentTitle = "AppEnvironmentTitle"
	MsgGLIBInfo            = "GLIBInfo"
	MsgGTKInfo             = "GTKInfo"
	MsgRsyncInfo           = "RsyncInfo"
	MsgGolangInfo          = "GolangInfo"
	MsgDialogYesButton     = "DialogYesButton"
	MsgDialogNoButton      = "DialogNoButton"
	MsgActionDoesNotFound  = "ActionDoesNotFound"

	MsgAboutDlgAppFeaturesAndBenefitsTitle   = "AboutDlgAppFeaturesAndBenefitsTitle"
	MsgAboutDlgAppFeaturesAndBenefitsSection = "AboutDlgAppFeaturesAndBenefitsSection"
	MsgAboutDlgAppDescriptionSection         = "AboutDlgAppDescriptionSection"
	MsgAboutDlgReleasedUnderLicense          = "AboutDlgReleasedUnderLicense"
	MsgAboutDlgFollowMyGithubProjectTitle    = "AboutDlgFollowMyGithubProjectTitle"
	MsgAboutDlgAppCopyright                  = "AboutDlgAppCopyright"
	MsgAboutDlgAppAuthorsBlock               = "AboutDlgAppAuthorsBlock"
	MsgAboutDlgAppLearnMore                  = "AboutDlgAppLearnMore"
	MsgAboutDlgDoNotShowCaption              = "AboutDlgDoNotShowCaption"

	MsgPrefDlgGeneralUserInterfaceOptionsSecion       = "PrefDlgGeneralUserInterfaceOptionsSecion"
	MsgPrefDlgGeneralBackupSettingsSection            = "PrefDlgGeneralBackupSettingsSection"
	MsgPrefDlgAdvancedRsyncDedupSettingsSection       = "PrefDlgAdvancedRsyncDedupSettingsSection"
	MsgPrefDlgAdvansedRsyncSettingsSection            = "PrefDlgAdvansedRsyncSettingsSection"
	MsgPrefDlgAdvancedBackupSettingsSection           = "PrefDlgAdvancedBackupSettingsSection"
	MsgPrefDlgAdvancedRsyncFileTransferOptionsSection = "PrefDlgAdvancedRsyncFileTransferOptionsSection"
	MsgPrefDlgRestartPanelCaptionWithLink             = "PrefDlgRestartPanelCaptionWithLink"

	MsgPrefDlgDoNotShowAtAppStartupCaption = "PrefDlgDoNotShowAtAppStartupCaption"
	MsgPrefDlgDoNotShowAtAppStartupHint    = "PrefDlgDoNotShowAtAppStartupHint"

	MsgPrefDlgSessionLogControlFontSizeCaption = "PrefDlgSessionLogControlFontSizeCaption"
	MsgPrefDlgSessionLogControlFontSizeHint    = "PrefDlgSessionLogControlFontSizeHint"

	MsgPrefDlgSourcesCaption                  = "PrefDlgSourcesCaption"
	MsgPrefDlgSourceRsyncPathCaption          = "PrefDlgSourceRsyncPathCaption"
	MsgPrefDlgSourceRsyncPathRetryHint        = "PrefDlgSourceRsyncPathRetryHint"
	MsgPrefDlgSourceRsyncPathDescriptionHint  = "PrefDlgSourceRsyncPathDescriptionHint"
	MsgPrefDlgSourceRsyncPathNotValidatedHint = "PrefDlgSourceRsyncPathNotValidatedHint"
	MsgPrefDlgSourceRsyncPathEmptyError       = "PrefDlgSourceRsyncPathEmptyError"
	MsgPrefDlgSourceRsyncValidatingHint       = "PrefDlgSourceRsyncValidatingHint"

	MsgPrefDlgDestinationSubpathCaption          = "PrefDlgDestinationSubpathCaption"
	MsgPrefDlgDestinationSubpathHint             = "PrefDlgDestinationSubpathHint"
	MsgPrefDlgDestinationSubpathNotValidatedHint = "PrefDlgDestinationSubpathNotValidatedHint"
	MsgPrefDlgDestinationSubpathExpressionError  = "PrefDlgDestinationSubpathExpressionError"
	MsgPrefDlgDestinationSubpathNotUniqueError   = "PrefDlgDestinationSubpathNotUniqueError"

	MsgPrefDlgExtraOptionsBoxCaption      = "PrefDlgExtraOptionsBoxCaption"
	MsgPrefDlgExtraOptionsBoxHint         = "PrefDlgExtraOptionsBoxHint"
	MsgPrefDlgAuthPasswordCaption         = "PrefDlgAuthPasswordCaption"
	MsgPrefDlgAuthPasswordHint            = "PrefDlgAuthPasswordHint"
	MsgPrefDlgChangeFilePermissionCaption = "PrefDlgChangeFilePermissionCaption"
	MsgPrefDlgChangeFilePermissionHint    = "PrefDlgChangeFilePermissionHint"

	MsgPrefDlgOverrideRsyncTransferOptionsBoxCaption = "PrefDlgOverrideRsyncTransferOptionsBoxCaption"
	MsgPrefDlgOverrideRsyncTransferOptionsBoxHint    = "PrefDlgOverrideRsyncTransferOptionsBoxHint"

	MsgPrefDlgEnableBackupBlockCaption = "PrefDlgEnableBackupBlockCaption"
	MsgPrefDlgEnableBackupBlockHint    = "PrefDlgEnableBackupBlockHint"

	MsgPrefDlgDeleteBackupBlockCaption     = "PrefDlgDeleteBackupBlockCaption"
	MsgPrefDlgDeleteBackupBlockHint        = "PrefDlgDeleteBackupBlockHint"
	MsgPrefDlgDeleteBackupBlockDialogTitle = "PrefDlgDeleteBackupBlockDialogTitle"
	MsgPrefDlgDeleteBackupBlockDialogText  = "PrefDlgDeleteBackupBlockDialogText"

	MsgPrefDlgProfileNameCaption       = "PrefDlgProfileNameCaption"
	MsgPrefDlgProfileNameHint          = "PrefDlgProfileNameHint"
	MsgPrefDlgProfileNameExistsWarning = "PrefDlgProfileNameExistsWarning"
	MsgPrefDlgProfileNameEmptyWarning  = "PrefDlgProfileNameEmptyWarning"

	MsgPrefDlgDefaultDestPathCaption = "PrefDlgDefaultDestPathCaption"
	MsgPrefDlgDefaultDestPathHint    = "PrefDlgDefaultDestPathHint"

	MsgPrefDlgSkipFolderBackupFileSignatureCaption = "PrefDlgSkipFolderBackupFileSignatureCaption"
	MsgPrefDlgSkipFolderBackupFileSignatureHint    = "PrefDlgSkipFolderBackupFileSignatureHint"

	MsgPrefDlgPerformDesktopNotificationCaption = "PrefDlgPerformDesktopNotificationCaption"
	MsgPrefDlgPerformDesktopNotificationHint    = "PrefDlgPerformDesktopNotificationHint"

	MsgPrefDlgRunNotificationScriptCaption = "PrefDlgRunNotificationScriptCaption"
	MsgPrefDlgRunNotificationScriptHint    = "PrefDlgRunNotificationScriptHint"

	MsgPrefDlgAutoManageBackupBlockSizeCaption = "PrefDlgAutoManageBackupBlockSizeCaption"
	MsgPrefDlgAutoManageBackupBlockSizeHint    = "PrefDlgAutoManageBackupBlockSizeHint"

	MsgPrefDlgBackupBlockSizeCaption = "PrefDlgBackupBlockSizeCaption"
	MsgPrefDlgBackupBlockSizeHint    = "PrefDlgBackupBlockSizeHint"

	MsgPrefDlgRsyncRetryCountCaption = "PrefDlgRsyncRetryCountCaption"
	MsgPrefDlgRsyncRetryCountHint    = "PrefDlgRsyncRetryCountHint"

	MsgPrefDlgRsyncLowLevelLogCaption = "PrefDlgRsyncLowLevelLogCaption"
	MsgPrefDlgRsyncLowLevelLogHint    = "PrefDlgRsyncLowLevelLogHint"

	MsgPrefDlgRsyncIntensiveLowLevelLogCaption = "PrefDlgRsyncIntensiveLowLevelLogCaption"
	MsgPrefDlgRsyncIntensiveLowLevelLogHint    = "PrefDlgRsyncIntensiveLowLevelLogHint"

	MsgPrefDlgUsePreviousBackupForDedupCaption = "PrefDlgUsePreviousBackupForDedupCaption"
	MsgPrefDlgUsePreviousBackupForDedupHint    = "PrefDlgUsePreviousBackupForDedupHint"

	MsgPrefDlgNumberOfPreviousBackupToUseCaption = "PrefDlgNumberOfPreviousBackupToUseCaption"
	MsgPrefDlgNumberOfPreviousBackupToUseHint    = "PrefDlgNumberOfPreviousBackupToUseHint"

	MsgPrefDlgRsyncCompressFileTransferCaption = "PrefDlgRsyncCompressFileTransferCaption"
	MsgPrefDlgRsyncCompressFileTransferHint    = "PrefDlgRsyncCompressFileTransferHint"

	MsgPrefDlgRsyncTransferSourcePermissionsCaption = "PrefDlgRsyncTransferSourcePermissionsCaption"
	MsgPrefDlgRsyncTransferSourcePermissionsHint    = "PrefDlgRsyncTransferSourcePermissionsHint"

	MsgPrefDlgRsyncTransferSourceOwnerCaption = "PrefDlgRsyncTransferSourceOwnerCaption"
	MsgPrefDlgRsyncTransferSourceOwnerHint    = "PrefDlgRsyncTransferSourceOwnerHint"

	MsgPrefDlgRsyncTransferSourceGroupCaption = "PrefDlgRsyncTransferSourceGroupCaption"
	MsgPrefDlgRsyncTransferSourceGroupHint    = "PrefDlgRsyncTransferSourceGroupHint"

	MsgPrefDlgRsyncRecreateSymlinksCaption = "PrefDlgRsyncRecreateSymlinksCaption"
	MsgPrefDlgRsyncRecreateSymlinksHint    = "PrefDlgRsyncRecreateSymlinksHint"

	MsgPrefDlgRsyncTransferDeviceFilesCaption = "PrefDlgRsyncTransferDeviceFilesCaption"
	MsgPrefDlgRsyncTransferDeviceFilesHint    = "PrefDlgRsyncTransferDeviceFilesHint"

	MsgPrefDlgRsyncTransferSpecialFilesCaption = "PrefDlgRsyncTransferSpecialFilesCaption"
	MsgPrefDlgRsyncTransferSpecialFilesHint    = "PrefDlgRsyncTransferSpecialFilesHint"

	MsgPrefDlgLanguageCaption                    = "PrefDlgLanguageCaption"
	MsgPrefDlgLanguageHint                       = "PrefDlgLanguageHint"
	MsgPrefDlgDefaultLanguageEntry               = "PrefDlgDefaultLanguageEntry"
	MsgPrefDlgAddBackupBlockHint                 = "PrefDlgAddBackupBlockHint"
	MsgPrefDlgProfileConfigIssuesDetectedWarning = "PrefDlgProfileConfigIssuesDetectedWarning"
	MsgPrefDlgPreferencesDialogCaption           = "PrefDlgPreferencesDialogCaption"

	MsgPrefDlgGeneralProfileTabName = "PrefDlgGeneralProfileTabName"
	MsgPrefDlgProfileTabName        = "PrefDlgProfileTabName"
	MsgPrefDlgGeneralTabName        = "PrefDlgGeneralTabName"
	MsgPrefDlgAdvancedTabName       = "PrefDlgAdvancedTabName"

	MsgPrefDlgAddProfileHint           = "PrefDlgAddProfileHint"
	MsgPrefDlgDeleteProfileHint        = "PrefDlgDeleteProfileHint"
	MsgPrefDlgDeleteProfileDialogTitle = "PrefDlgDeleteProfileDialogTitle"
	MsgPrefDlgDeleteProfileDialogText  = "PrefDlgDeleteProfileDialogText"

	MsgSchemaConfigDlgTitle                   = "SchemaConfigDlgTitle"
	MsgSchemaConfigDlgNoSchemaFoundError      = "SchemaConfigDlgNoSchemaFoundError"
	MsgSchemaConfigDlgSchemaDoesNotFoundError = "SchemaConfigDlgSchemaDoesNotFoundError"
	MsgSchemaConfigDlgSchemaErrorAdvise       = "SchemaConfigDlgSchemaErrorAdvise"

	MsgAppWindowAboutMenuCaption       = "AppWindowAboutMenuCaption"
	MsgAppWindowHelpMenuCaption        = "AppWindowHelpMenuCaption"
	MsgAppWindowPreferencesMenuCaption = "AppWindowPreferencesMenuCaption"
	MsgAppWindowPreferencesHint        = "AppWindowPreferencesHint"
	MsgAppWindowQuitMenuCaption        = "AppWindowQuitMenuCaption"
	MsgAppWindowRunBackupHint          = "AppWindowRunBackupHint"
	MsgAppWindowStopBackupHint         = "AppWindowStopBackupHint"

	MsgAppWindowProfileCaption                      = "AppWindowProfileCaption"
	MsgAppWindowProfileHint                         = "AppWindowProfileHint"
	MsgAppWindowProfileBackupPlanInfoSourceCount    = "AppWindowProfileBackupPlanInfoSourceCount"
	MsgAppWindowProfileBackupPlanInfoTotalSize      = "AppWindowProfileBackupPlanInfoTotalSize"
	MsgAppWindowProfileBackupPlanInfoSkipSize       = "AppWindowProfileBackupPlanInfoSkipSize"
	MsgAppWindowProfileBackupPlanInfoDirectoryCount = "AppWindowProfileBackupPlanInfoDirectoryCount"
	MsgAppWindowInquiringProfileStatus              = "AppWindowInquiringProfileStatus"
	MsgAppWindowNoneProfileEntry                    = "AppWindowNoneProfileEntry"

	MsgAppWindowRsyncPathIsEmptyError      = "AppWindowRsyncPathIsEmptyError"
	MsgAppWindowDestPathCaption            = "AppWindowDestPathCaption"
	MsgAppWindowDestPathHint               = "AppWindowDestPathHint"
	MsgAppWindowDestPathIsValidStatusPart1 = "AppWindowDestPathIsValidStatusPart1"
	MsgAppWindowDestPathIsValidStatusPart2 = "AppWindowDestPathIsValidStatusPart2"
	MsgAppWindowDestPathIsEmptyError1      = "AppWindowDestPathIsEmptyError1"
	MsgAppWindowDestPathIsEmptyError2      = "AppWindowDestPathIsEmptyError2"
	MsgAppWindowDestPathIsNotExistError    = "AppWindowDestPathIsNotExistError"
	MsgAppWindowDestPathIsNotExistAdvise   = "AppWindowDestPathIsNotExistAdvise"

	MsgAppWindowBackupProgressStartMessage               = "AppWindowBackupProgressStartMessage"
	MsgAppWindowBackupProgressInquiringSourceID          = "AppWindowBackupProgressInquiringSourceID"
	MsgAppWindowBackupProgressInquiringSourceDescription = "AppWindowBackupProgressInquiringSourceDescription"
	MsgAppWindowBackupProgressTimePassedSuffix           = "AppWindowBackupProgressTimePassedSuffix"
	MsgAppWindowBackupProgressETASuffix                  = "AppWindowBackupProgressETASuffix"
	MsgAppWindowBackupProgressSizeCompletedSuffix        = "AppWindowBackupProgressSizeCompletedSuffix"
	MsgAppWindowBackupProgressSizeLeftToProcessSuffix    = "AppWindowBackupProgressSizeLeftToProcessSuffix"
	MsgAppWindowBackupProgressCompleted                  = "AppWindowBackupProgressCompleted"
	MsgAppWindowBackupProgressCompletedWithErrors        = "AppWindowBackupProgressCompletedWithErrors"
	MsgAppWindowBackupProgressTerminated                 = "AppWindowBackupProgressTerminated"
	MsgAppWindowBackupProgressFailed                     = "AppWindowBackupProgressFailed"
	MsgAppWindowOverallProgressCaption                   = "AppWindowOverallProgressCaption"
	MsgAppWindowProgressStatusCaption                    = "AppWindowProgressStatusCaption"
	MsgAppWindowSessionLogCaption                        = "AppWindowSessionLogCaption"
	MsgAppWindowCannotStartBackupProcessTitle            = "AppWindowCannotStartBackupProcessTitle"

	MsgAppWindowTerminateBackupDlgTitle = "AppWindowTerminateBackupDlgTitle"
	MsgAppWindowTerminateBackupDlgText  = "AppWindowTerminateBackupDlgText"

	MsgAppWindowOutOfSpaceDlgTitle           = "AppWindowOutOfSpaceDlgTitle"
	MsgAppWindowOutOfSpaceDlgText1           = "AppWindowOutOfSpaceDlgText1"
	MsgAppWindowOutOfSpaceDlgText2           = "AppWindowOutOfSpaceDlgText2"
	MsgAppWindowOutOfSpaceDlgIgnoreButton    = "AppWindowOutOfSpaceDlgIgnoreButton"
	MsgAppWindowOutOfSpaceDlgRetryButton     = "AppWindowOutOfSpaceDlgRetryButton"
	MsgAppWindowOutOfSpaceDlgTerminateButton = "AppWindowOutOfSpaceDlgTerminateButton"

	MsgAppWindowRsyncUtilityDlgTitle         = "AppWindowRsyncUtilityDlgTitle"
	MsgAppWindowRsyncUtilityDlgNotFoundError = "AppWindowRsyncUtilityDlgNotFoundError"

	MsgAppWindowShowNotificationError             = "AppWindowShowNotificationError"
	MsgAppWindowRunNotificationScriptError        = "AppWindowRunNotificationScriptError"
	MsgAppWindowNotificationScriptExecutableError = "AppWindowNotificationScriptExecutableError"
	MsgAppWindowGetExecutableScriptInfoError      = "AppWindowGetExecutableScriptInfoError"

	MsgLogBackupStageOutOfSpaceWarning = "LogBackupStageOutOfSpaceWarning"

	MsgGeneralHintStatusCaption      = "GeneralHintStatusCaption"
	MsgGeneralHintDescriptionCaption = "GeneralHintDescriptionCaption"

	MsgDesktopNotificationBackupSuccessfullyCompleted = "DesktopNotificationBackupSuccessfullyCompleted"
	MsgDesktopNotificationBackupCompletedWithErrors   = "DesktopNotificationBackupCompletedWithErrors"
	MsgDesktopNotificationBackupTerminated            = "DesktopNotificationBackupTerminated"
	MsgDesktopNotificationBackupFailed                = "DesktopNotificationBackupFailed"
	MsgDesktopNotificationTotalSize                   = "DesktopNotificationTotalSize"
	MsgDesktopNotificationSkippedSize                 = "DesktopNotificationSkippedSize"
	MsgDesktopNotificationFailedToBackupSize          = "DesktopNotificationFailedToBackupSize"
	MsgDesktopNotificationTimeTaken                   = "DesktopNotificationTimeTaken"
)
