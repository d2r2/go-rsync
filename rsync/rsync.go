package rsync

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/d2r2/go-rsync/core"
	shell "github.com/d2r2/go-shell"
)

// RSYNC_CMD contains RSYNC console utility name to run.
const RSYNC_CMD = "rsync"

// RunRsyncWithRetry run RSYNC utility with retry attempts.
func RunRsyncWithRetry(ctx context.Context, options *Options, log *Logging, stdOut *bytes.Buffer,
	paths core.SrcDstPath) (sessionErr, retryErr, criticalErr error) {

	retryCount := 0
	if options != nil {
		retryCount = options.RetryCount
	}
	index := 0
	for {
		err := runSystemRsync(ctx, options.Password,
			options.Params, log, stdOut,
			paths.RsyncSourcePath, paths.DestPath)

		if err == nil {
			return
		} else if IsRsyncProcessTerminatedError(err) {
			sessionErr = err
			criticalErr = err
			return
		}

		if err != nil {
			retryErr = err
		}

		// in case of error we are trying to recover from
		// fail state via call to ErrorHook call-back function
		if options != nil && options.ErrorHook != nil {
			var newRetryLeft int
			newRetryLeft, criticalErr = options.ErrorHook.Call(err, paths,
				options.ErrorHook.PredictedSize, index, retryCount)
			if criticalErr != nil {
				break
			}
			retryCount = newRetryLeft
		}

		retryCount--
		if retryCount < 0 {
			break
		}
		index++
	}
	if criticalErr == nil && retryErr != nil {
		sessionErr = retryErr
		retryErr = nil
	}
	return
}

// IsInstalled verify, that RSYNC application present in the system.
func IsInstalled() error {
	app := shell.NewApp(RSYNC_CMD)
	return app.CheckIsInstalled()
}

// GetRsyncVersion run RSYNC to get version and protocol.
func GetRsyncVersion() (version string, protocol string, err error) {
	app := shell.NewApp(RSYNC_CMD, "--version")
	var stdOut, stdErr bytes.Buffer
	exitCode := app.Run(&stdOut, &stdErr)
	if exitCode.Error != nil {
		return "", "", exitCode.Error
	}
	scanner := bufio.NewScanner(&stdOut)
	scanner.Split(bufio.ScanLines)

	// Expression should parse a line: rsync  version 3.1.3  protocol version 31
	re := regexp.MustCompile(`version\s+(?P<version>\d+\.\d+(\.\d+)?)(\s+protocol\s+version\s+(?P<protocol>\d+))?`)
	for scanner.Scan() {
		line := scanner.Text()
		m := core.FindStringSubmatchIndexes(re, line)
		if len(m) > 0 {
			grName := "version"
			if _, ok := m[grName]; ok {
				start := m[grName][0]
				end := m[grName][1]
				version = line[start:end]
			}
			grName = "protocol"
			if _, ok := m[grName]; ok {
				start := m[grName][0]
				end := m[grName][1]
				protocol = line[start:end]
			}
			break
		}
	}
	return version, protocol, nil
}

// runSystemRsync run RSYNC utility.
// Parameters:
//	- Save console output to stdOut variable.
func runSystemRsync(ctx context.Context, password *string,
	params []string, log *Logging, stdOut *bytes.Buffer,
	source, dest string) error {

	var args []string
	if params != nil {
		args = params
	}
	args = append(args, source, dest)
	stdOut2 := stdOut
	stdErr := bytes.NewBuffer(nil)

	var logBuf bytes.Buffer
	logEnabled := false
	if log != nil && log.EnableLog && log.Log != nil {
		logEnabled = true
		if stdOut2 == nil {
			stdOut2 = bytes.NewBuffer(nil)
		}
	}

	app := shell.NewApp(RSYNC_CMD, args...)
	/*
		if user != nil {
			app.AddEnvironments([]string{fmt.Sprintf("USER=%s", *user)})
			lg.Debugf("USER: %v", *user)
		}
	*/
	var passwd string
	if password != nil {
		passwd = *password
	}
	// always add password, even empty one, for protection for
	// RSYNC module raise password input via stdin
	app.AddEnvironments([]string{fmt.Sprintf("RSYNC_PASSWORD=%s", passwd)})
	if passwd != "" {
		lg.Debugf("PASSWD: %v", passwd)
	}
	lg.Debugf("Args: %v", args)
	waitCh, err := app.Start(stdOut2, stdErr)
	if err != nil {
		return err
	}

	// stdOutLastLen := -1
	// stdErrLastLen := -1
	// for {
	select {
	case <-ctx.Done():
		lg.Debugf("Killing rsync: %v", args)
		err := app.Kill()
		if err != nil {
			return err
		}
		return &RsyncProcessTerminatedError{}
	case st := <-waitCh:
		if logEnabled {
			logBuf.WriteString(RSYNC_CMD)
			if len(args) > 0 {
				logBuf.WriteString(" ")
				logBuf.WriteString(strings.Join(args, " "))
			}
			if log.EnableIntensiveLog {
				logBuf.WriteString(fmt.Sprintln())
				logBuf.WriteString(fmt.Sprintln(">>>>>>>>>>>>>>>> Stdout start >>>>>>>>>>>>>>>>"))
				logBuf.WriteString(fmt.Sprintln(strings.TrimRight(stdOut2.String(), "\n")))
				logBuf.WriteString(fmt.Sprint("<<<<<<<<<<<<<<<< Stdout end <<<<<<<<<<<<<<<<"))
			}
			log.Log.Info(logBuf.String())
		}
		if st.Error != nil {
			return st.Error
		} else if st.ExitCode != 0 {
			lg.Debugf("STDERR: %v", stdErr.String())
			return NewRsyncCallFailedError(st.ExitCode, stdErr)
		}
		return nil
		// case <-time.After(5 * time.Second):
		// 	stdOutLastLen = stdOut2.Len()
		// 	stdErrLastLen = stdErr.Len()
		// 	lg.Debugf("stdOutLastLen=%v", stdOutLastLen)
		// 	lg.Debugf("stdErrLastLen=%v", stdErrLastLen)
	}
	// }
	return nil
}
