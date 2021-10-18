//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2022 Denis Dyakov <denis.dyakov@gma**.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package rsync

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/d2r2/go-rsync/core"
	"github.com/d2r2/go-rsync/locale"
)

// ProcessTerminatedError denote a situation with termination pending.
type ProcessTerminatedError struct {
}

func (v *ProcessTerminatedError) Error() string {
	return locale.T(MsgRsyncProcessTerminatedError, nil)
}

// IsProcessTerminatedError check that error able to cast
// to ProcessTerminatedError.
func IsProcessTerminatedError(err error) bool {
	if err != nil {
		_, ok := err.(*ProcessTerminatedError)
		return ok
	}
	return false
}

// CallFailedError denote a situation when RSYNC execution
// completed with non-zero exit code.
type CallFailedError struct {
	ExitCode    int
	Description string
}

// extractError used to extract textual description of error
// which improve understanding of error root cause.
func extractError(stdErr *bytes.Buffer) string {
	var descr string
	buf := stdErr.String()
	re := regexp.MustCompile(`(?m:^@ERROR:(?P<error>.*)$)`)
	m := core.FindStringSubmatchIndexes(re, buf)
	if len(m) > 0 {
		grErr := "error"
		if _, ok := m[grErr]; ok {
			start := m[grErr][0]
			end := m[grErr][1]
			descr = strings.TrimSpace(buf[start:end])
		}
	}
	return descr
}

// NewCallFailedError creates error object based on ExitCode from RSYNC.
// Use STDERR variable to extract more human readable error description.
func NewCallFailedError(exitCode int, stdErr *bytes.Buffer) *CallFailedError {
	descr := extractError(stdErr)
	if descr != "" {
		descr += ", " + getRsyncExitCodeDesc(exitCode)
	} else {
		descr = getRsyncExitCodeDesc(exitCode)
	}

	v := &CallFailedError{
		ExitCode:    exitCode,
		Description: descr,
	}
	return v
}

func (v *CallFailedError) Error() string {
	return locale.T(MsgRsyncCallFailedError,
		struct {
			Description string
			ExitCode    int
		}{Description: v.Description, ExitCode: v.ExitCode})
}

// IsCallFailedError check that error able to cast
// to CallFailedError.
func IsCallFailedError(err error) bool {
	if err != nil {
		_, ok := err.(*CallFailedError)
		return ok
	}
	return false
}

// GetRsyncExitCodeDesc return RSYNC exit code descriptions
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
		return f("Undefined rsync exit code: %d", exitCode)
	}
}

// ExtractVersionAndProtocolError denote a situation when attempt
// to extract rsync version/protocol has failed, and
// version and protocol are undefined.
// This error is not critical (in the main) and should not lead to app failure.
type ExtractVersionAndProtocolError struct {
}

func (v *ExtractVersionAndProtocolError) Error() string {
	return locale.T(MsgRsyncExtractVersionAndProtocolError, nil)
}

// IsExtractVersionAndProtocolError check that error able to cast
// to ExtractVersionAndProtocolError.
func IsExtractVersionAndProtocolError(err error) bool {
	if err != nil {
		_, ok := err.(*ExtractVersionAndProtocolError)
		return ok
	}
	return false
}
