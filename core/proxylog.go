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

package core

import (
	"fmt"

	logger "github.com/d2r2/go-logger"
	"github.com/davecgh/go-spew/spew"
)

// WriteLine is a delegate to describe log output call.
type WriteLine func(line string) error

// ProxyLog is used to substitute regular log console output
// with output to the file, either to the GUI window.
// ProxyLog implements logger.PackageLog interface which
// provide regular log methods.
type ProxyLog struct {
	parent      logger.PackageLog
	packageName string
	packageLen  int
	timeFormat  string

	customWriteLine WriteLine
	customLogLevel  logger.LogLevel
}

// Static cast to verify that type implement specific interface
var _ logger.PackageLog = &ProxyLog{}

func NewProxyLog(parent logger.PackageLog, packageName string, packageLen int,
	timeFormat string, writeLine WriteLine, customLogLevel logger.LogLevel) *ProxyLog {

	v := &ProxyLog{parent: parent, packageName: packageName, packageLen: packageLen,
		timeFormat: timeFormat, customLogLevel: customLogLevel,
		customWriteLine: writeLine}
	return v
}

func (v *ProxyLog) getFormat() logger.FormatOptions {
	options := logger.FormatOptions{TimeFormat: v.timeFormat,
		LevelLength: logger.LevelShort, PackageLength: v.packageLen}
	return options
}

// Printf implement logger.PackageLog.Printf method.
func (v *ProxyLog) Printf(level logger.LogLevel, format string, args ...interface{}) {
	if v.parent != nil {
		v.parent.Printf(level, format, args...)
	}
	if v.customWriteLine != nil && level <= v.customLogLevel {
		msg := spew.Sprintf(format, args...)
		packageName := v.packageName
		out := logger.FormatMessage(v.getFormat(), level, packageName, msg, false)
		err := v.customWriteLine(out + fmt.Sprintln())
		if err != nil {
			v.parent.Fatal(err)
		}
	}
}

// Print implement logger.PackageLog.Print method.
func (v *ProxyLog) Print(level logger.LogLevel, args ...interface{}) {
	if v.parent != nil {
		v.parent.Print(level, args...)
	}
	if v.customWriteLine != nil && level <= v.customLogLevel {
		msg := fmt.Sprint(args...)
		packageName := v.packageName
		out := logger.FormatMessage(v.getFormat(), level, packageName, msg, false)
		err := v.customWriteLine(out + fmt.Sprintln())
		if err != nil {
			v.parent.Fatal(err)
		}
	}
}

// Debugf implement logger.PackageLog.Debugf method.
func (v *ProxyLog) Debugf(format string, args ...interface{}) {
	v.Printf(logger.DebugLevel, format, args...)
}

// Debug implement logger.PackageLog.Debug method.
func (v *ProxyLog) Debug(args ...interface{}) {
	v.Print(logger.DebugLevel, args...)
}

// Infof implement logger.PackageLog.Infof method.
func (v *ProxyLog) Infof(format string, args ...interface{}) {
	v.Printf(logger.InfoLevel, format, args...)
}

// Info implement logger.PackageLog.Info method.
func (v *ProxyLog) Info(args ...interface{}) {
	v.Print(logger.InfoLevel, args...)
}

// Notifyf implement logger.PackageLog.Notifyf method.
func (v *ProxyLog) Notifyf(format string, args ...interface{}) {
	v.Printf(logger.NotifyLevel, format, args...)
}

// Notify implement logger.PackageLog.Notify method.
func (v *ProxyLog) Notify(args ...interface{}) {
	v.Print(logger.NotifyLevel, args...)
}

// Warningf implement logger.PackageLog.Warningf method.
func (v *ProxyLog) Warningf(format string, args ...interface{}) {
	v.Printf(logger.WarnLevel, format, args...)
}

// Warnf implement logger.PackageLog.Warnf method.
func (v *ProxyLog) Warnf(format string, args ...interface{}) {
	v.Printf(logger.WarnLevel, format, args...)
}

// Warning implement logger.PackageLog.Warning method.
func (v *ProxyLog) Warning(args ...interface{}) {
	v.Print(logger.WarnLevel, args...)
}

// Warn implement logger.PackageLog.Warn method.
func (v *ProxyLog) Warn(args ...interface{}) {
	v.Print(logger.WarnLevel, args...)
}

// Errorf implement logger.PackageLog.Errorf method.
func (v *ProxyLog) Errorf(format string, args ...interface{}) {
	v.Printf(logger.ErrorLevel, format, args...)
}

// Error implement logger.PackageLog.Error method.
func (v *ProxyLog) Error(args ...interface{}) {
	v.Print(logger.ErrorLevel, args...)
}

// Panicf implement logger.PackageLog.Panicf method.
func (v *ProxyLog) Panicf(format string, args ...interface{}) {
	v.Printf(logger.PanicLevel, format, args...)
}

// Panic implement logger.PackageLog.Panic method.
func (v *ProxyLog) Panic(args ...interface{}) {
	v.Print(logger.PanicLevel, args...)
}

// Fatalf implement logger.PackageLog.Fatalf method.
func (v *ProxyLog) Fatalf(format string, args ...interface{}) {
	v.Printf(logger.FatalLevel, format, args...)
}

// Fatal implement logger.PackageLog.Fatal method.
func (v *ProxyLog) Fatal(args ...interface{}) {
	v.Print(logger.FatalLevel, args...)
}
