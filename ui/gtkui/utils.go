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

package gtkui

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/d2r2/go-rsync/backup"
	"github.com/d2r2/go-rsync/data"
	"github.com/d2r2/go-rsync/locale"
	"github.com/d2r2/gotk3/gdk"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/davecgh/go-spew/spew"
)

func PixbufFromAssetsNew(assetIconName string) (*gdk.Pixbuf, error) {
	file, err := data.Assets.Open(assetIconName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pb, err := getPixbufFromBytes(b)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func PixbufFromAssetsNewWithResize(assetIconName string,
	resizeToWidth, resizeToHeight int) (*gdk.Pixbuf, error) {

	file, err := data.Assets.Open(assetIconName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pb, err := getPixbufFromBytesWithResize(b, resizeToWidth, resizeToHeight)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func PixbufAnimationFromAssetsNew(assetIconName string) (*gdk.PixbufAnimation, error) {
	file, err := data.Assets.Open(assetIconName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pb, err := getPixbufAnimationFromBytes(b)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func PixbufAnimationFromAssetsNewWithResize(assetIconName string,
	resizeToWidth, resizeToHeight int) (*gdk.PixbufAnimation, error) {

	file, err := data.Assets.Open(assetIconName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	pb, err := getPixbufAnimationFromBytesWithResize(b, resizeToWidth, resizeToHeight)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func AnimationImageFromAssetsNew(assetIconName string) (*gtk.Image, error) {
	pba, err := PixbufAnimationFromAssetsNew(assetIconName)
	if err != nil {
		return nil, err
	}
	img, err := gtk.ImageNewFromAnimation(pba)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func AnimationImageFromAssetsNewWithResize(assetIconName string,
	resizeToWidth, resizeToHeight int) (*gtk.Image, error) {

	pba, err := PixbufAnimationFromAssetsNewWithResize(assetIconName, resizeToWidth, resizeToHeight)
	if err != nil {
		return nil, err
	}
	img, err := gtk.ImageNewFromAnimation(pba)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ImageFromAssetsNew(assetIconName string) (*gtk.Image, error) {
	pb2, err := PixbufFromAssetsNew(assetIconName)
	if err != nil {
		return nil, err
	}

	img, err := gtk.ImageNewFromPixbuf(pb2)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func ImageFromAssetsNewWithResize(assetIconName string, resizeToWidth, resizeToHeight int) (*gtk.Image, error) {
	pb2, err := PixbufFromAssetsNewWithResize(assetIconName, resizeToWidth, resizeToHeight)
	if err != nil {
		return nil, err
	}

	img, err := gtk.ImageNewFromPixbuf(pb2)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func SetEntryIconWithAssetImage(entry *gtk.Entry, iconPos gtk.EntryIconPosition, assetIconName string) error {
	pb, err := PixbufFromAssetsNew(assetIconName)
	if err != nil {
		return err
	}
	entry.SetIconFromPixbuf(iconPos, pb)
	return nil
}

func SetupButtonWithAssetAnimationImage(assetIconName string) (*gtk.Button, error) {
	img, err := AnimationImageFromAssetsNew(assetIconName)
	if err != nil {
		return nil, err
	}

	btn, err := gtk.ButtonNew()
	if err != nil {
		return nil, err
	}

	btn.Add(img)

	return btn, nil
}

// CheckSchemaSettingsIsInstalled verify, that GLib Setting's schema is installed, otherwise return false.
func CheckSchemaSettingsIsInstalled(settingsID string, app *gtk.Application, extraMsg *string) (bool, error) {
	parent := app.GetActiveWindow()
	// Verify that GSettingsSchema is installed
	schemaSource := glib.SettingsSchemaSourceGetDefault()
	if schemaSource == nil {
		//title := "<span weight='bold' size='larger'>Schema settings configuration error</span>"
		text := locale.T(MsgSchemaConfigDlgNoSchemaFoundError, nil)
		err := schemaSettingsErrorDialog(parent, text, extraMsg)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	schema := schemaSource.Lookup(settingsID, false)
	if schema == nil {
		//title := "<span weight='bold' size='larger'>Schema settings configuration error</span>"
		text := locale.T(MsgSchemaConfigDlgSchemaDoesNotFoundError, nil)
		err := schemaSettingsErrorDialog(parent, text, extraMsg)
		if err != nil {
			return false, err
		}
		return false, nil
	}
	return true, nil
}

// ProgressBarManage simplify setting up GtkProgressBar to pulse either progress mode.
type ProgressBarManage struct {
	sync.Mutex
	progressBar *gtk.ProgressBar
	pulse       *time.Ticker
	stopPulse   chan struct{}
}

func NewProgressBarManage(pb *gtk.ProgressBar) *ProgressBarManage {
	p := &ProgressBarManage{progressBar: pb}
	return p
}

func (v *ProgressBarManage) StartPulse() {
	v.Lock()
	defer v.Unlock()

	if v.stopPulse == nil {
		v.progressBar.SetPulseStep(0.5)
		v.progressBar.Pulse()
		//v.progressBar.Pulse()
		v.stopPulse = make(chan struct{})
		v.pulse = time.NewTicker(time.Millisecond * 2000)
		go func(stopPulse chan struct{}) {
			for {
				select {
				case <-v.pulse.C:
					MustIdleAdd(func() {
						v.progressBar.Pulse()
					})
				case <-stopPulse:
					v.Lock()
					v.pulse.Stop()
					v.Unlock()
					MustIdleAdd(func() {
						v.progressBar.SetFraction(0)
					})
					return
				}
			}
		}(v.stopPulse)
	}
}

func (v *ProgressBarManage) StopPulse() {
	v.Lock()
	defer v.Unlock()

	if v.stopPulse != nil {
		close(v.stopPulse)
		v.stopPulse = nil
	}
}

func (v *ProgressBarManage) SetFraction(value float64) error {
	v.StopPulse()
	v.Lock()
	defer v.Unlock()

	MustIdleAdd(func() {
		v.progressBar.SetFraction(value)
	})
	return nil
}

func (v *ProgressBarManage) AddProgressBarStyleClass(cssClass string) error {
	v.Lock()
	defer v.Unlock()

	MustIdleAdd(func() {
		err := AddStyleClass(&v.progressBar.Widget, cssClass)
		if err != nil {
			lg.Fatal(err)
		}
	})
	return nil
}

func (v *ProgressBarManage) RemoveProgressBarStyleClass(cssClass string) error {
	v.Lock()
	defer v.Unlock()

	MustIdleAdd(func() {
		err := RemoveStyleClass(&v.progressBar.Widget, cssClass)
		if err != nil {
			lg.Fatal(err)
		}
	})
	return nil
}

// ControlWithStatus wraps control to the box to attach extra status widget to the right.
// Status widget would be a error icon either spin control to show active process.
type ControlWithStatus struct {
	box       *gtk.Box
	control   *gtk.Widget
	statusBox *gtk.Box
}

func NewControlWithStatus(control *gtk.Widget) (*ControlWithStatus, error) {
	box, err := gtk.BoxNew(gtk.ORIENTATION_HORIZONTAL, 6)
	if err != nil {
		return nil, err
	}
	box.Add(control)
	box.SetHExpand(true)
	v := &ControlWithStatus{box: box, control: control}
	return v, nil
}

func (v *ControlWithStatus) ReplaceStatus(statusBox *gtk.Box) {
	if v.statusBox != nil {
		v.statusBox.Destroy()
		v.statusBox = nil
	}
	if statusBox != nil {
		v.statusBox = statusBox
		v.box.Add(statusBox)
		v.box.ShowAll()
	}
}

func (v *ControlWithStatus) GetBox() *gtk.Box {
	return v.box
}

type GLibIdleCallStub struct {
	sync.Mutex
}

// var glibIdleCallStub GLibIdleCallStub

var IdleAdd = func(f interface{}, args ...interface{}) (glib.SourceHandle, error) {
	// glibIdleCallStub.Lock()
	// defer glibIdleCallStub.Unlock()
	return glib.IdleAdd(f, args...)
}

var MustIdleAdd = func(f interface{}, args ...interface{}) {
	// glibIdleCallStub.Lock()
	// defer glibIdleCallStub.Unlock()
	_, err := glib.IdleAdd(f, args...)
	if err != nil {
		lg.Fatalf("error creating call glib.IdleAdd: %v", err)
	}
}

/*
var MustIdleAddWait = func(call func()) {
	glibIdleCallStub.Lock()
	defer glibIdleCallStub.Unlock()

	ch := make(chan struct{})
	_, err := glib.IdleAdd(func() {
		call()
		close(ch)
	})
	if err != nil {
		lg.Fatalf("error creating call glib.IdleAdd: %v", err)
	}
	<-ch
}
*/

// GetBaseApplicationCSS read from assets CSS file, which
// give UI styles used for customization of application interface.
func GetBaseApplicationCSS() (string, error) {
	// Load CSS styles
	file, err := data.Assets.Open("base.css")
	if err != nil {
		return "", err
	}
	defer file.Close()
	b, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}
	buf := bytes.NewBuffer(b)
	return buf.String(), nil
}

func normalizeSubpath(subpath string) string {
	subpath = strings.TrimSpace(subpath)
	subpath1 := []rune(subpath)
	if len(subpath1) > 0 && subpath1[0] == rune(os.PathSeparator) {
		subpath1 = subpath1[1:]
	}
	if len(subpath1) > 0 && subpath1[len(subpath1)-1] == rune(os.PathSeparator) {
		subpath1 = subpath1[:len(subpath1)-1]
	}
	subpath = string(subpath1)
	return subpath
}

// markupTooltip create hint for GtkWidget to display description
// with standard template: "Status: ... Description: ...".
func markupTooltip(status *Markup, description string) *Markup {
	var mp *Markup
	if status != nil {
		mp = NewMarkup(0, 0, 0, nil, nil,
			NewMarkup(0, MARKUP_COLOR_LIGHT_GRAY, 0,
				spew.Sprintf("%s ", locale.T(MsgGeneralHintStatusCaption, nil)), nil),
			status,
			NewMarkup(0, MARKUP_COLOR_LIGHT_GRAY, 0,
				spew.Sprintf("\n%s ", locale.T(MsgGeneralHintDescriptionCaption, nil)), nil),
			NewMarkup(0, 0, 0, description, nil),
		)
	} else {
		mp = NewMarkup(0, 0, 0, description, nil)
	}
	return mp
}

// isDestPathError verify file system path availability status.
// Returns error, if path isn't reachable.
func isDestPathError(destPath string, formatMultiline bool) (bool, string) {
	if destPath == "" {
		msg := locale.T(MsgAppWindowDestPathIsEmptyError1, nil)
		return true, msg
	} else {
		_, err := os.Stat(destPath)
		if err != nil {
			var msg string
			if os.IsNotExist(err) {
				var buf bytes.Buffer
				buf.WriteString(locale.T(MsgAppWindowDestPathIsNotExistError,
					struct{ FolderPath string }{FolderPath: destPath}))
				if formatMultiline {
					buf.WriteString(spew.Sprintln())
				} else {
					buf.WriteString(" ")
				}
				buf.WriteString(locale.T(MsgAppWindowDestPathIsNotExistAdvise, nil))
				msg = buf.String()
			} else {
				msg = err.Error()
			}
			return true, msg
		}
	}
	return false, ""
}

func isModulesConfigError(modules []backup.Module, formatMultiline bool) (bool, string) {
	for _, module := range modules {
		// check for empty RSYNC source path
		if module.SourceRsync == "" {
			msg := locale.T(MsgAppWindowRsyncPathIsEmptyError, nil)
			if !formatMultiline {
				// Do not use strings.ReplaceAll() since it was implemented
				// only in Go 1.12. Instead use strings.Replace().
				msg = strings.Replace(msg, "\n", " ", -1)
			}
			return true, msg
		}
	}
	return false, ""
}

// RestartTimer restart timer with call fire after specific millisecond period.
// Used as a trigger for validation events.
func RestartTimer(timer *time.Timer, milliseconds time.Duration) {
	timer.Stop()
	timer.Reset(time.Millisecond * milliseconds)
}
