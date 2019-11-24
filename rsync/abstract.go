package rsync

import (
	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
)

// Logging keep settings whether we need to log RSYNC utility functioning.
// We can log only RSYNC calls, but also STDOUT output for intensive log.
type Logging struct {
	EnableLog          bool
	EnableIntensiveLog bool
	Log                logger.PackageLog
}

// ErrorHookCall is a delegate used to work around RSYNC issues
// caused by out of disk space case.
type ErrorHookCall func(err error, paths core.SrcDstPath, predictedSize *core.FolderSize,
	repeated int, retryLeft int) (newRetryLeft int, criticalError error)

// ErrorHook contains call and predicted size to work around RSYNC issues
// caused by out of disk space case.
type ErrorHook struct {
	Call          ErrorHookCall
	PredictedSize *core.FolderSize
}

func NewErrorHook(call ErrorHookCall, predictedSize core.FolderSize) *ErrorHook {
	v := &ErrorHook{Call: call, PredictedSize: &predictedSize}
	return v
}

// Options keep settings for RSYNC call.
// Settings include: retry count, parameters, ErrorHook object
// for recover attempt if issue thrown.
type Options struct {
	RetryCount int
	Params     []string
	ErrorHook  *ErrorHook
	Password   *string
}

func NewOptions(params []string) *Options {
	options := &Options{Params: params}
	return options
}

// AddParams add RSYNC command line options.
func (v *Options) AddParams(params ...string) *Options {
	v.Params = append(v.Params, params...)
	return v
}

// SetRetryCount set retry count for repeated call in case
// of error return results (exit code <> 0).
func (v *Options) SetRetryCount(retryCount *int) *Options {
	if retryCount != nil {
		if *retryCount >= 0 {
			// limit number of retry count to 5 maximum
			if *retryCount < 6 {
				v.RetryCount = *retryCount
			} else {
				v.RetryCount = 5
			}
		}
	}
	return v
}

// SetAuthPassword set password to use in RSYNC call to
// get data from authenticated (password protected) RSYNC module.
// Read option "secrets file" at https://linux.die.net/man/5/rsyncd.conf,
// which describe how to protect RSYNC data source with password.
func (v *Options) SetAuthPassword(password *string) *Options {
	v.Password = password
	return v
}

// SetErrorHook define callback function to run, if RESYNC
// utility exited with error code <> 0.
// Such callback might suggest issue source and make recommendation
// to user via UI to resolve the issue before following retry.
func (v *Options) SetErrorHook(errorHook *ErrorHook) *Options {
	v.ErrorHook = errorHook
	return v
}

// WithDefaultParams return list of obligatory options
// for each run of RSYNC utility.
func WithDefaultParams(params []string) []string {
	defParams := []string{"--progress", "--verbose"}
	params2 := append(defParams, params...)
	return params2
}
