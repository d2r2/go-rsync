package backup

// ------------------------------------------------------------
// File contains message identifiers for localization purpose.
// Message identifier name's is self-descriptive, so ordinary
// it's easy to understand what message is made for.
// Message ID is used to call translation functions from
// "locale" package.
// ------------------------------------------------------------

const (
	MsgRsyncInfo  = "RsyncInfo"
	MsgGolangInfo = "GolangInfo"

	MsgFolderBackupTypeSkipDescription      = "FolderBackupTypeSkipDescription"
	MsgFolderBackupTypeRecursiveDescription = "FolderBackupTypeRecursiveDescription"
	MsgFolderBackupTypeContentDescription   = "FolderBackupTypeContentDescription"

	MsgLogPlanStageStarting                  = "LogPlanStageStarting"
	MsgLogPlanStageStartTime                 = "LogPlanStageStartTime"
	MsgLogPlanStageEndTime                   = "LogPlanStageEndTime"
	MsgLogPlanStartIterateViaNSources        = "LogPlanStartIterateViaNSources"
	MsgLogPlanStageInquirySource             = "LogPlanStageInquirySource"
	MsgLogPlanStageSourceFolderCountInfo     = "LogPlanStageSourceFolderCountInfo"
	MsgLogPlanStageSourceSkipFolderCountInfo = "LogPlanStageSourceSkipFolderCountInfo"
	MsgLogPlanStageSourceTotalSizeInfo       = "LogPlanStageSourceTotalSizeInfo"
	MsgLogPlanStageUseTemporaryFolder        = "LogPlanStageUseTemporaryFolder"
	MsgLogPlanStageBuildFolderError          = "LogPlanStageBuildFolderError"

	MsgLogBackupStageStarting                               = "LogBackupStageStarting"
	MsgLogBackupStageStartTime                              = "LogBackupStageStartTime"
	MsgLogBackupStageEndTime                                = "LogBackupStageEndTime"
	MsgLogBackupStageBackupToDestination                    = "LogBackupStageBackupToDestination"
	MsgLogBackupStagePreviousBackupDiscoveryPermissionError = "LogBackupStagePreviousBackupDiscoveryPermissionError"
	MsgLogBackupStagePreviousBackupDiscoveryOtherError      = "LogBackupStagePreviousBackupDiscoveryOtherError"
	MsgLogBackupStagePreviousBackupFoundAndWillBeUsed       = "LogBackupStagePreviousBackupFoundAndWillBeUsed"
	MsgLogBackupStagePreviousBackupFoundButDisabled         = "LogBackupStagePreviousBackupFoundButDisabled"
	MsgLogBackupStagePreviousBackupNotFound                 = "LogBackupStagePreviousBackupNotFound"
	MsgLogBackupStageStartToBackupFromSource                = "LogBackupStageStartToBackupFromSource"
	MsgLogBackupStageRenameDestination                      = "LogBackupStageRenameDestination"
	MsgLogBackupStageFailedToCreateFolder                   = "LogBackupStageFailedToCreateFolder"
	MsgLogBackupDetectedTotalBackupSizeGetChanged           = "LogBackupDetectedTotalBackupSizeGetChanged"
	MsgLogBackupStageProgressBackupSuccess                  = "LogBackupStageProgressBackupSuccess"
	MsgLogBackupStageProgressBackupError                    = "LogBackupStageProgressBackupError"
	MsgLogBackupStageProgressSkipBackupError                = "LogBackupStageProgressSkipBackupError"
	MsgLogBackupStageCriticalError                          = "LogBackupStageCriticalError"
	MsgLogBackupStageDiscoveringPreviousBackups             = "LogBackupStageDiscoveringPreviousBackups"
	MsgLogBackupStageRecoveredFromError                     = "LogBackupStageRecoveredFromError"
	MsgLogBackupStageSaveRsyncExtraLogTo                    = "LogBackupStageSaveRsyncExtraLogTo"
	MsgLogBackupStageSaveLogTo                              = "LogBackupStageSaveLogTo"
	MsgLogBackupStageExitMessage                            = "LogBackupStageExitMessage"

	MsgLogStatisticsSummaryCaption                            = "LogStatisticsSummaryCaption"
	MsgLogStatisticsEnvironmentCaption                        = "LogStatisticsEnvironmentCaption"
	MsgLogStatisticsResultsCaption                            = "LogStatisticsResultsCaption"
	MsgLogStatisticsStatusCaption                             = "LogStatisticsStatusCaption"
	MsgLogStatisticsStatusSuccessfullyCompleted               = "LogStatisticsStatusSuccessfullyCompleted"
	MsgLogStatisticsStatusCompletedWithErrors                 = "LogStatisticsStatusCompletedWithErrors"
	MsgLogStatisticsPlanStageCaption                          = "LogStatisticsPlanStageCaption"
	MsgLogStatisticsPlanStageSourceToBackup                   = "LogStatisticsPlanStageSourceToBackup"
	MsgLogStatisticsPlanStageTotalSize                        = "LogStatisticsPlanStageTotalSize"
	MsgLogStatisticsPlanStageFolderCount                      = "LogStatisticsPlanStageFolderCount"
	MsgLogStatisticsPlanStageFolderSkipCount                  = "LogStatisticsPlanStageFolderSkipCount"
	MsgLogStatisticsPlanStageTimeTaken                        = "LogStatisticsPlanStageTimeTaken"
	MsgLogStatisticsBackupStageCaption                        = "LogStatisticsBackupStageCaption"
	MsgLogStatisticsBackupStageDestinationPath                = "LogStatisticsBackupStageDestinationPath"
	MsgLogStatisticsBackupStagePreviousBackupFound            = "LogStatisticsBackupStagePreviousBackupFound"
	MsgLogStatisticsBackupStagePreviousBackupFoundButDisabled = "LogStatisticsBackupStagePreviousBackupFoundButDisabled"
	MsgLogStatisticsBackupStageNoValidPreviousBackupFound     = "LogStatisticsBackupStageNoValidPreviousBackupFound"
	MsgLogStatisticsBackupStageTotalSize                      = "LogStatisticsBackupStageTotalSize"
	MsgLogStatisticsBackupStageSkippedSize                    = "LogStatisticsBackupStageSkippedSize"
	MsgLogStatisticsBackupStageFailedToBackupSize             = "LogStatisticsBackupStageFailedToBackupSize"
	MsgLogStatisticsBackupStageTimeTaken                      = "LogStatisticsBackupStageTimeTaken"
)
