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

func (v *Options) AddParams(params ...string) *Options {
	v.Params = append(v.Params, params...)
	return v
}

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

func (v *Options) SetAuthPassword(password *string) *Options {
	v.Password = password
	return v
}

func (v *Options) SetErrorHook(errorHook *ErrorHook) *Options {
	v.ErrorHook = errorHook
	return v
}

func WithDefaultParams(params []string) []string {
	defParams := []string{"--progress", "--verbose"}
	params2 := append(defParams, params...)
	return params2
}
