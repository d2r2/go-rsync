package rsync

import (
	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
)

type RsyncProcessTerminatedError struct {
}

func (v *RsyncProcessTerminatedError) Error() string {
	return locale.T(MsgRsyncProcessTerminatedError, nil)
}

func IsRsyncProcessTerminatedError(err error) bool {
	if err != nil {
		_, ok := err.(*RsyncProcessTerminatedError)
		return ok
	}
	return false
}

type RsyncCallFailedError struct {
	ExitCode    int
	Description string
}

func NewRsyncCallFailedError(exitCode int) *RsyncCallFailedError {
	v := &RsyncCallFailedError{
		ExitCode:    exitCode,
		Description: getRsyncExitCodeDesc(exitCode),
	}
	return v
}

func (v *RsyncCallFailedError) Error() string {
	return locale.T(MsgRsyncCallFailedError,
		struct {
			Description string
			ExitCode    int
		}{Description: v.Description, ExitCode: v.ExitCode})
}

func IsRsyncCallFailedError(err error) bool {
	if err != nil {
		_, ok := err.(*RsyncCallFailedError)
		return ok
	}
	return false
}

// GetRsyncExitCodeDesc return rsync exit code descriptions
// taken from here: http://wpkg.org/Rsync_exit_codes
func getRsyncExitCodeDesc(exitCode int) string {
	codes := map[int]string{
		0: "success",
		1: "syntax or usage error",
		2: "protocol incompatibility",
		3: "errors selecting input/output files, dirs",
		4: "requested action not supported: an attempt was made to manipulate " +
			"64-bit files on a platform that cannot support them; or an option was " +
			"specified that is supported by the client and not by the server",
		5:   "error starting client-server protocol",
		6:   "daemon unable to append to log-file",
		10:  "error in socket I/O",
		11:  "error in file I/O",
		12:  "error in rsync protocol data stream",
		13:  "errors with program diagnostics",
		14:  "error in IPC code",
		20:  "received SIGUSR1 or SIGINT",
		21:  "some error returned by waitpid()",
		22:  "error allocating core memory buffers",
		23:  "partial transfer due to error",
		24:  "partial transfer due to vanished source files",
		25:  "the --max-delete limit stopped deletions",
		30:  "timeout in data send/receive",
		35:  "timeout waiting for daemon connection",
		255: "unexplained error",
	}
	if v, ok := codes[exitCode]; ok {
		return v
	} else {
		return f("Unknown rsync exit code: %d", exitCode)
	}
}

type ErrorHook func(err error, paths core.SrcDstPath, predictedSize *core.FolderSize,
	repeated int, retryLeft int) (newRetryLeft int, criticalError error)
