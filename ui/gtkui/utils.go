package gtkui

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

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

func ImageFromAssetsNew(assetIconName string, resizeToDestWidth, resizeToDestHeight int) (*gtk.Image, error) {
	pb, err := PixbufFromAssetsNew(assetIconName)
	if err != nil {
		return nil, err
	}
	pb2 := pb
	if resizeToDestWidth >= 0 && resizeToDestHeight >= 0 {
		pb2, err = pb.ScaleSimple(resizeToDestWidth, resizeToDestHeight, gdk.INTERP_BILINEAR)
		if err != nil {
			return nil, err
		}
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
					_, err := glib.IdleAdd(func() {
						v.progressBar.Pulse()
					})
					if err != nil {
						lg.Fatal(err)
					}
				case <-stopPulse:
					v.Lock()
					v.pulse.Stop()
					v.Unlock()
					_, err := glib.IdleAdd(func() {
						v.progressBar.SetFraction(0)
					})
					if err != nil {
						lg.Fatal(err)
					}
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

	_, err := glib.IdleAdd(func() {
		v.progressBar.SetFraction(value)
	})
	if err != nil {
		return err
	}
	return nil
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
