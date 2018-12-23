package rsync

import (
	logger "github.com/d2r2/go-logger"
	"github.com/d2r2/go-rsync/core"
)

// // Keep application exit state including exit code
// // with information about error if took place.
// type ErrorSpec struct {
// 	Error     error
// 	ErrorCode int
// 	RetryLeft int
// }

// // GetError build error object from internal failure state.
// func (v *ErrorSpec) GetError() error {
// 	var err error
// 	if v != nil {
// 		if v.Error == ErrRsyncProcessTerminated {
// 			err = v.Error
// 		} else {
// 			err = errors.New(f("RSYNC call failed (%s, code %d)",
// 				v.Error, v.ErrorCode))
// 		}
// 	}
// 	return err
// }

type Logging struct {
	EnableLog          bool
	EnableIntensiveLog bool
	Log                logger.PackageLog
}

type Options struct {
	RetryCount    int
	Params        []string
	ErrorHook     ErrorHook
	PredictedSize *core.FolderSize
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
			if *retryCount < 6 {
				v.RetryCount = *retryCount
			} else {
				v.RetryCount = 5
			}
		}
	}
	return v
}

func (v *Options) SetErrorHook(errorHook ErrorHook) *Options {
	v.ErrorHook = errorHook
	return v
}

func (v *Options) SetPredictedSize(predictedSize core.FolderSize) *Options {
	v.PredictedSize = &predictedSize
	return v
}

func WithDefaultParams(params ...string) []string {
	defParams := []string{"--progress", "--verbose"}
	params2 := append(defParams, params...)
	return params2
}

func Params(params ...string) []string {
	return params
}
