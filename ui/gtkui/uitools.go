//--------------------------------------------------------------------------------------------------
// This file is a part of Gorsync Backup project (backup RSYNC frontend).
// Copyright (c) 2017-2020 Denis Dyakov <denis.dyakov@gmail.com>
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
// BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//--------------------------------------------------------------------------------------------------

package gtkui

import (
	"github.com/d2r2/gotk3/gdk"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/d2r2/gotk3/pango"
	"github.com/davecgh/go-spew/spew"
)

// ========================================================================================
// ************************* GTK+ UI UTILITIES SECTION START ******************************
// ========================================================================================
//	In real application use this code section as utilities to simplify creation
//	of GLIB/GTK+ components and widgets, including menus, dialog boxes, messages,
//	application settings and so on...

// SetupLabelJustifyRight create GtkLabel with justification to the right by default.
func SetupLabelJustifyRight(caption string) (*gtk.Label, error) {
	lbl, err := gtk.LabelNew(caption)
	if err != nil {
		return nil, err
	}
	lbl.SetHAlign(gtk.ALIGN_END)
	lbl.SetJustify(gtk.JUSTIFY_RIGHT)
	return lbl, nil
}

// SetupLabelJustifyLeft create GtkLabel with justification to the left by default.
func SetupLabelJustifyLeft(caption string) (*gtk.Label, error) {
	lbl, err := gtk.LabelNew(caption)
	if err != nil {
		return nil, err
	}
	lbl.SetHAlign(gtk.ALIGN_START)
	lbl.SetJustify(gtk.JUSTIFY_LEFT)
	return lbl, nil
}

// SetupLabelJustifyCenter create GtkLabel with justification to the center by default.
func SetupLabelJustifyCenter(caption string) (*gtk.Label, error) {
	lbl, err := gtk.LabelNew(caption)
	if err != nil {
		return nil, err
	}
	lbl.SetHAlign(gtk.ALIGN_CENTER)
	lbl.SetJustify(gtk.JUSTIFY_CENTER)
	return lbl, nil
}

// SetupLabelMarkupJustifyRight create GtkLabel with justification to the right by default.
func SetupLabelMarkupJustifyRight(caption *Markup) (*gtk.Label, error) {
	captionText := ""
	if caption != nil {
		captionText = caption.String()
	}
	lbl, err := gtk.LabelNew(captionText)
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_END)
	lbl.SetJustify(gtk.JUSTIFY_RIGHT)
	return lbl, nil
}

// SetupLabelMarkupJustifyLeft create GtkLabel with justification to the left by default.
func SetupLabelMarkupJustifyLeft(caption *Markup) (*gtk.Label, error) {
	captionText := ""
	if caption != nil {
		captionText = caption.String()
	}
	lbl, err := gtk.LabelNew(captionText)
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_START)
	lbl.SetJustify(gtk.JUSTIFY_LEFT)
	return lbl, nil
}

// SetupLabelMarkupJustifyCenter create GtkLabel with justification to the center by default.
func SetupLabelMarkupJustifyCenter(caption *Markup) (*gtk.Label, error) {
	captionText := ""
	if caption != nil {
		captionText = caption.String()
	}
	lbl, err := gtk.LabelNew(captionText)
	if err != nil {
		return nil, err
	}
	lbl.SetUseMarkup(true)
	lbl.SetHAlign(gtk.ALIGN_CENTER)
	lbl.SetJustify(gtk.JUSTIFY_CENTER)
	return lbl, nil
}

// SetupHeader construct Header widget with standard initialization.
func SetupHeader(title, subtitle string, showCloseButton bool) (*gtk.HeaderBar, error) {
	hdr, err := gtk.HeaderBarNew()
	if err != nil {
		return nil, err
	}
	hdr.SetShowCloseButton(showCloseButton)
	hdr.SetTitle(title)
	if subtitle != "" {
		hdr.SetSubtitle(subtitle)
	}
	return hdr, nil
}

// SetupMenuItemWithIcon construct MenuItem widget with icon image.
func SetupMenuItemWithIcon(label, detailedAction string, icon *glib.Icon) (*glib.MenuItem, error) {
	mi, err := glib.MenuItemNew(label, detailedAction)
	if err != nil {
		return nil, err
	}
	//mi.SetAttributeValue("verb-icon", iconNameVar)
	mi.SetIcon(icon)
	return mi, nil
}

// SetupMenuItemWithThemedIcon construct MenuItem widget with image
// taken by iconName from themed icons image lib.
func SetupMenuItemWithThemedIcon(label, detailedAction, iconName string) (*glib.MenuItem, error) {
	iconNameVar, err := glib.VariantStringNew(iconName)
	if err != nil {
		return nil, err
	}
	mi, err := glib.MenuItemNew(label, detailedAction)
	if err != nil {
		return nil, err
	}
	mi.SetAttributeValue("verb-icon", iconNameVar)
	return mi, nil
}

// SetupToolButton construct ToolButton widget with standart initialization.
func SetupToolButton(themedIconName, label string) (*gtk.ToolButton, error) {
	var btn *gtk.ToolButton
	var img *gtk.Image
	var err error
	if themedIconName != "" {
		img, err = gtk.ImageNewFromIconName(themedIconName, gtk.ICON_SIZE_BUTTON)
		if err != nil {
			return nil, err
		}
	}

	btn, err = gtk.ToolButtonNew(img, label)
	if err != nil {
		return nil, err
	}
	return btn, nil
}

// SetupButtonWithThemedImage construct Button widget with image
// taken by themedIconName from themed icons image lib.
func SetupButtonWithThemedImage(themedIconName string) (*gtk.Button, error) {
	img, err := gtk.ImageNewFromIconName(themedIconName, gtk.ICON_SIZE_BUTTON)
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

// getPixbufFromBytes create gdk.PixBuf loaded from raw bytes buffer
func getPixbufFromBytes(bytes []byte) (*gdk.Pixbuf, error) {
	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	pbl, err := gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}
	err = pbl.WriteBytes(b2)
	if err != nil {
		return nil, err
	}
	err = pbl.Close()
	if err != nil {
		return nil, err
	}

	pb, err := pbl.GetPixbuf()
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// getPixbufFromBytesWithResize create gdk.PixBuf loaded from raw bytes buffer, applying resize
func getPixbufFromBytesWithResize(bytes []byte, resizeToWidth, resizeToHeight int) (*gdk.Pixbuf, error) {
	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	pbl, err := gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}
	pbl.SetSize(resizeToWidth, resizeToHeight)
	err = pbl.WriteBytes(b2)
	if err != nil {
		return nil, err
	}
	err = pbl.Close()
	if err != nil {
		return nil, err
	}

	pb, err := pbl.GetPixbuf()
	if err != nil {
		return nil, err
	}
	return pb, nil
}

// getPixbufFromBytes create gdk.PixbufAnimation loaded from raw bytes buffer
func getPixbufAnimationFromBytes(bytes []byte) (*gdk.PixbufAnimation, error) {
	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	pbl, err := gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}
	err = pbl.WriteBytes(b2)
	if err != nil {
		return nil, err
	}
	err = pbl.Close()
	if err != nil {
		return nil, err
	}

	pba, err := pbl.GetPixbufAnimation()
	if err != nil {
		return nil, err
	}
	return pba, nil
}

// getPixbufFromBytesWithResize create gdk.PixbufAnimation loaded from raw bytes buffer, applying resize
func getPixbufAnimationFromBytesWithResize(bytes []byte, resizeToWidth,
	resizeToHeight int) (*gdk.PixbufAnimation, error) {

	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	pbl, err := gdk.PixbufLoaderNew()
	if err != nil {
		return nil, err
	}
	pbl.SetSize(resizeToWidth, resizeToHeight)
	err = pbl.WriteBytes(b2)
	if err != nil {
		return nil, err
	}
	err = pbl.Close()
	if err != nil {
		return nil, err
	}

	pba, err := pbl.GetPixbufAnimation()
	if err != nil {
		return nil, err
	}
	return pba, nil
}

// SetupMenuButtonWithThemedImage construct gtk.MenuButton widget with image
// taken by themedIconName from themed icons image lib.
func SetupMenuButtonWithThemedImage(themedIconName string) (*gtk.MenuButton, error) {
	img, err := gtk.ImageNewFromIconName(themedIconName, gtk.ICON_SIZE_BUTTON)
	if err != nil {
		return nil, err
	}

	btn, err := gtk.MenuButtonNew()
	if err != nil {
		return nil, err
	}

	btn.Add(img)

	return btn, nil
}

// AppendSectionAsHorzButtons used for gtk.Popover widget menu
// as a hint to display items as a horizontal buttons.
func AppendSectionAsHorzButtons(main, section *glib.Menu) error {
	val1, err := glib.VariantStringNew("horizontal-buttons")
	if err != nil {
		return err
	}
	mi1, err := glib.MenuItemNew("", "")
	if err != nil {
		return err
	}
	mi1.SetSection(section)
	mi1.SetAttributeValue("display-hint", val1)
	main.AppendItem(mi1)
	//section.AppendItem(mi1)
	return nil
}

// DialogButton simplify dialog window initialization.
// Keep all necessary information about how attached
// dialog button should look and act.
type DialogButton struct {
	Text      string
	Response  gtk.ResponseType
	Default   bool
	Customize func(button *gtk.Button) error
}

// GetActiveWindow find real active window in application running.
func GetActiveWindow(win *gtk.Window) (*gtk.Window, error) {
	app, err := win.GetApplication()
	if err != nil {
		return nil, err
	}
	return app.GetActiveWindow(), nil
}

// IsResponseYes gives true if dialog window
// responded with gtk.RESPONSE_YES.
func IsResponseYes(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_YES
}

// IsResponseNo gives true if dialog window
// responded with gtk.RESPONSE_NO.
func IsResponseNo(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_NO
}

// IsResponseNone gives true if dialog window
// responded with gtk.RESPONSE_NONE.
func IsResponseNone(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_NONE
}

// IsResponseOk gives true if dialog window
// responded with gtk.RESPONSE_OK.
func IsResponseOk(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_OK
}

// IsResponseCancel gives true if dialog window
// responded with gtk.RESPONSE_CANCEL.
func IsResponseCancel(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_CANCEL
}

// IsResponseReject gives true if dialog window
// responded with gtk.RESPONSE_REJECT.
func IsResponseReject(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_REJECT
}

// IsResponseClose gives true if dialog window
// responded with gtk.RESPONSE_CLOSE.
func IsResponseClose(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_CLOSE
}

// IsResponseDeleteEvent gives true if dialog window
// responded with gtk.RESPONSE_DELETE_EVENT.
func IsResponseDeleteEvent(response gtk.ResponseType) bool {
	return response == gtk.RESPONSE_DELETE_EVENT
}

// PrintDialogResponse print and debug dialog response.
func PrintDialogResponse(response gtk.ResponseType) {
	if IsResponseNo(response) {
		lg.Debug("Dialog result = NO")
	} else if IsResponseYes(response) {
		lg.Debug("Dialog result = YES")
	} else if IsResponseNone(response) {
		lg.Debug("Dialog result = NONE")
	} else if IsResponseOk(response) {
		lg.Debug("Dialog result = OK")
	} else if IsResponseReject(response) {
		lg.Debug("Dialog result = REJECT")
	} else if IsResponseCancel(response) {
		lg.Debug("Dialog result = CANCEL")
	} else if IsResponseClose(response) {
		lg.Debug("Dialog result = CLOSE")
	} else if IsResponseDeleteEvent(response) {
		lg.Debug("Dialog result = DELETE_EVENT")
	}
}

// DialogParagraph is an object which keep text paragraph added
// to dialog window, complemented with all necessary format options.
type DialogParagraph struct {
	Text          string
	Markup        bool
	HorizAlign    gtk.Align
	Justify       gtk.Justification
	Ellipsize     pango.EllipsizeMode
	MaxWidthChars int
}

// NewDialogParagraph create new text paragraph instance,
// with default align, justification and so on.
func NewDialogParagraph(text string) *DialogParagraph {
	v := &DialogParagraph{Text: text, HorizAlign: gtk.Align(-1), Justify: gtk.Justification(-1),
		Ellipsize: pango.EllipsizeMode(-1), MaxWidthChars: -1}
	return v
}

// SetMarkup update Markup flag.
func (v *DialogParagraph) SetMarkup(markup bool) *DialogParagraph {
	v.Markup = markup
	return v
}

// SetHorizAlign set horizontal alignment of text paragraph.
func (v *DialogParagraph) SetHorizAlign(align gtk.Align) *DialogParagraph {
	v.HorizAlign = align
	return v
}

// SetJustify set text justification.
func (v *DialogParagraph) SetJustify(justify gtk.Justification) *DialogParagraph {
	v.Justify = justify
	return v
}

// SetEllipsize set text ellipsis mode.
func (v *DialogParagraph) SetEllipsize(ellipsize pango.EllipsizeMode) *DialogParagraph {
	v.Ellipsize = ellipsize
	return v
}

// SetMaxWidthChars set maximum number of chars in one line.
func (v *DialogParagraph) SetMaxWidthChars(maxWidthChars int) *DialogParagraph {
	v.MaxWidthChars = maxWidthChars
	return v
}

// createLabel create gtk.Label widget to put paragraph text in.
func (v *DialogParagraph) createLabel() (*gtk.Label, error) {
	lbl, err := gtk.LabelNew("")
	if err != nil {
		return nil, err
	}
	if v.Markup {
		lbl.SetMarkup(v.Text)
	} else {
		lbl.SetText(v.Text)
	}
	if v.HorizAlign != gtk.Align(-1) {
		lbl.SetHAlign(v.HorizAlign)
	}
	if v.Justify != gtk.Justification(-1) {
		lbl.SetJustify(v.Justify)
	}
	if v.Ellipsize != pango.EllipsizeMode(-1) {
		lbl.SetEllipsize(v.Ellipsize)
	}
	if v.MaxWidthChars != -1 {
		lbl.SetMaxWidthChars(v.MaxWidthChars)
	}
	return lbl, nil
}

// TextToDialogParagraphs multi-line text to DialogParagraph instance.
func TextToDialogParagraphs(lines []string) []*DialogParagraph {
	var msgs []*DialogParagraph
	for _, line := range lines {
		msgs = append(msgs, NewDialogParagraph(line))
	}
	return msgs
}

// TextToMarkupDialogParagraphs multi-line markup text to DialogParagraph instance.
func TextToMarkupDialogParagraphs(makrupLines []string) []*DialogParagraph {
	var msgs []*DialogParagraph
	for _, markupLine := range makrupLines {
		msgs = append(msgs, NewDialogParagraph(markupLine).SetMarkup(true))
	}
	return msgs
}

type MessageDialog struct {
	dialog *gtk.MessageDialog
}

// SetupMessageDialog construct MessageDialog widget with customized settings.
func SetupMessageDialog(parent *gtk.Window, markupTitle string, secondaryMarkupTitle string,
	paragraphs []*DialogParagraph, addButtons []DialogButton,
	addExtraControls func(area *gtk.Box) error) (*MessageDialog, error) {

	var active *gtk.Window
	var err error

	if parent != nil {
		active, err = GetActiveWindow(parent)
		if err != nil {
			return nil, err
		}
	}

	dlg, err := gtk.MessageDialogNew(active, /*gtk.DIALOG_MODAL|*/
		gtk.DIALOG_USE_HEADER_BAR, gtk.MESSAGE_WARNING, gtk.BUTTONS_NONE, nil, nil)
	if err != nil {
		return nil, err
	}
	if active != nil {
		dlg.SetTransientFor(active)
	}
	dlg.SetMarkup(markupTitle)
	if secondaryMarkupTitle != "" {
		dlg.FormatSecondaryMarkup(secondaryMarkupTitle)
	}

	for _, button := range addButtons {
		btn, err := dlg.AddButton(button.Text, button.Response)
		if err != nil {
			return nil, err
		}

		if button.Default {
			dlg.SetDefaultResponse(button.Response)
		}

		if button.Customize != nil {
			err := button.Customize(btn)
			if err != nil {
				return nil, err
			}
		}
	}

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}

	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_CENTER)

	box, err := dlg.GetMessageArea()
	if err != nil {
		return nil, err
	}
	box.Add(grid)

	col := 1
	row := 0

	// add empty line after title
	paragraphs = append([]*DialogParagraph{NewDialogParagraph("")}, paragraphs...)

	for _, paragraph := range paragraphs {
		lbl, err := paragraph.createLabel()
		if err != nil {
			return nil, err
		}
		grid.Attach(lbl, col, row, 1, 1)
		row++
	}

	if addExtraControls != nil {
		box1, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
		if err != nil {
			return nil, err
		}
		grid.Attach(box1, col, row, 1, 1)

		err = addExtraControls(box1)
		if err != nil {
			return nil, err
		}
	}

	box.ShowAll()

	v := &MessageDialog{dialog: dlg}
	return v, nil
}

// Run run MessageDialog widget with customized settings.
func (v *MessageDialog) Run(ignoreCloseBox bool) gtk.ResponseType {

	defer v.dialog.Destroy()

	v.dialog.ShowAll()
	var res gtk.ResponseType
	res = v.dialog.Run()
	for gtk.ResponseType(res) == gtk.RESPONSE_NONE || gtk.ResponseType(res) == gtk.RESPONSE_DELETE_EVENT && ignoreCloseBox {
		res = v.dialog.Run()
	}
	return gtk.ResponseType(res)
}

// SetupDialog construct Dialog widget with customized settings.
func SetupDialog(parent *gtk.Window, messageType gtk.MessageType, userHeaderbar bool,
	title string, paragraphs []*DialogParagraph, addButtons []DialogButton,
	addExtraControls func(area *gtk.Box) error) (*gtk.Dialog, error) {

	var active *gtk.Window
	var err error

	if parent != nil {
		active, err = GetActiveWindow(parent)
		if err != nil {
			return nil, err
		}
	}

	flags := gtk.DIALOG_MODAL
	if userHeaderbar {
		flags |= gtk.DIALOG_USE_HEADER_BAR
	}
	dlg, err := gtk.DialogWithFlagsNew(title, active, flags)
	if err != nil {
		return nil, err
	}

	dlg.SetDefaultSize(100, 100)
	dlg.SetTransientFor(active)
	dlg.SetDeletable(false)

	var img *gtk.Image
	size := gtk.ICON_SIZE_DIALOG
	if userHeaderbar {
		size = gtk.ICON_SIZE_LARGE_TOOLBAR
	}
	var iconName string
	switch messageType {
	case gtk.MESSAGE_WARNING:
		iconName = "dialog-warning"
	case gtk.MESSAGE_ERROR:
		iconName = "dialog-error"
	case gtk.MESSAGE_INFO:
		iconName = "dialog-information"
	case gtk.MESSAGE_QUESTION:
		iconName = "dialog-question"
	}

	if iconName != "" {
		img, err = gtk.ImageNewFromIconName(iconName, size)
		if err != nil {
			return nil, err
		}
	}

	grid, err := gtk.GridNew()
	if err != nil {
		return nil, err
	}

	grid.SetBorderWidth(10)
	grid.SetColumnSpacing(10)
	grid.SetRowSpacing(6)
	grid.SetHAlign(gtk.ALIGN_CENTER)

	box, err := dlg.GetContentArea()
	if err != nil {
		return nil, err
	}
	box.Add(grid)

	if img != nil {
		if userHeaderbar {
			hdr, err := dlg.GetHeaderBar()
			if err != nil {
				return nil, err
			}

			hdr.PackStart(img)
		} else {
			grid.Attach(img, 0, 0, 1, 1)
		}
	}

	for _, button := range addButtons {
		btn, err := dlg.AddButton(button.Text, button.Response)
		if err != nil {
			return nil, err
		}

		if button.Default {
			dlg.SetDefaultResponse(button.Response)
		}

		if button.Customize != nil {
			err := button.Customize(btn)
			if err != nil {
				return nil, err
			}
		}
	}

	col := 1
	row := 0

	for _, paragraph := range paragraphs {
		lbl, err := paragraph.createLabel()
		if err != nil {
			return nil, err
		}
		grid.Attach(lbl, col, row, 1, 1)
		row++
	}

	if addExtraControls != nil {
		box1, err := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 6)
		if err != nil {
			return nil, err
		}
		grid.Attach(box1, col, row, 1, 1)

		err = addExtraControls(box1)
		if err != nil {
			return nil, err
		}
	}

	_, w := dlg.GetPreferredWidth()
	_, h := dlg.GetPreferredHeight()
	dlg.Resize(w, h)

	return dlg, nil
}

// RunDialog construct and run Dialog widget with customized settings.
func RunDialog(parent *gtk.Window, messageType gtk.MessageType, userHeaderbar bool,
	title string, paragraphs []*DialogParagraph, ignoreCloseBox bool, addButtons []DialogButton,
	addExtraControls func(area *gtk.Box) error) (gtk.ResponseType, error) {

	dlg, err := SetupDialog(parent, messageType, userHeaderbar, title,
		paragraphs, addButtons, addExtraControls)
	if err != nil {
		return 0, err
	}
	defer dlg.Destroy()

	//dlg.ShowAll()
	dlg.ShowAll()
	res := dlg.Run()
	for gtk.ResponseType(res) == gtk.RESPONSE_NONE || gtk.ResponseType(res) == gtk.RESPONSE_DELETE_EVENT && ignoreCloseBox {
		res = dlg.Run()
	}
	return gtk.ResponseType(res), nil
}

// ErrorMessage build and run error message dialog.
func ErrorMessage(parent *gtk.Window, titleMarkup string, text []*DialogParagraph) error {
	buttons := []DialogButton{
		{"_OK", gtk.RESPONSE_OK, false, nil},
	}
	dialog, err := SetupMessageDialog(parent, titleMarkup, "", text, buttons, nil)
	if err != nil {
		return err
	}
	dialog.Run(false)
	return nil
}

// QuestionDialog build and run question message dialog with Yes/No choice.
func QuestionDialog(parent *gtk.Window, title string,
	messages []*DialogParagraph, defaultYes bool) (bool, error) {

	title2 := spew.Sprintf("%s", title)
	buttons := []DialogButton{
		{"_YES", gtk.RESPONSE_YES, defaultYes, nil},
		{"_NO", gtk.RESPONSE_NO, !defaultYes, nil},
	}
	response, err := RunDialog(parent, gtk.MESSAGE_QUESTION, true, title2,
		messages, false, buttons, nil)
	if err != nil {
		return false, err
	}
	PrintDialogResponse(response)
	return IsResponseYes(response), nil
}

// GetActionNameAndState display status of action-with-state, which used in
// menu-with-state behavior. Convenient for debug purpose.
func GetActionNameAndState(act *glib.SimpleAction) (string, *glib.Variant, error) {
	name, err := act.GetName()
	if err != nil {
		return "", nil, err
	}
	state := act.GetState()
	return name, state, nil
}

// SetMargins set margins of a widget to the passed values,
// replacing 4 calls with only one.
func SetMargins(widget gtk.IWidget, left int, top int, right int, bottom int) {
	w := widget.GetWidget()
	w.SetMarginStart(left)
	w.SetMarginTop(top)
	w.SetMarginEnd(right)
	w.SetMarginBottom(bottom)
}

// SetAllMargins set all margins of a widget to the same value.
func SetAllMargins(widget gtk.IWidget, margin int) {
	SetMargins(widget, margin, margin, margin, margin)
}

// AppendValues append multiple values to a row in a list store.
func AppendValues(ls *gtk.ListStore, values ...interface{}) (*gtk.TreeIter, error) {
	iter := ls.Append()
	for i := 0; i < len(values); i++ {
		err := ls.SetValue(iter, i, values[i])
		if err != nil {
			return nil, err
		}
	}
	return iter, nil
}

// CreateNameValueCombo create a GtkComboBox that holds
// a set of name/value pairs where the name is displayed.
func CreateNameValueCombo(keyValues []struct{ value, key string }) (*gtk.ComboBox, error) {
	ls, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		return nil, err
	}

	for _, item := range keyValues {
		_, err = AppendValues(ls, item.value, item.key)
		if err != nil {
			return nil, err
		}
	}

	cb, err := gtk.ComboBoxNew()
	if err != nil {
		return nil, err
	}
	err = UpdateNameValueCombo(cb, keyValues)
	if err != nil {
		return nil, err
	}
	cb.SetFocusOnClick(false)
	cb.SetIDColumn(1)
	cell, err := gtk.CellRendererTextNew()
	if err != nil {
		return nil, err
	}
	cell.SetAlignment(0, 0)
	cb.PackStart(cell, false)
	cb.AddAttribute(cell, "text", 0)

	return cb, nil
}

// UpdateNameValueCombo update GtkComboBox list of name/value pairs.
func UpdateNameValueCombo(cb *gtk.ComboBox, keyValues []struct{ value, key string }) error {
	ls, err := gtk.ListStoreNew(glib.TYPE_STRING, glib.TYPE_STRING)
	if err != nil {
		return err
	}

	for _, item := range keyValues {
		_, err = AppendValues(ls, item.value, item.key)
		if err != nil {
			return err
		}
	}

	cb.SetModel(ls)
	return nil
}

// GetComboValue return GtkComboBox selected value from specific column.
func GetComboValue(cb *gtk.ComboBox, columnID int) (*glib.Value, error) {
	ti, err := cb.GetActiveIter()
	if err != nil {
		return nil, err
	}
	tm, err := cb.GetModel()
	if err != nil {
		return nil, err
	}
	val, err := tm.GetValue(ti, 0)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// GetGtkVersion return actually installed GTK+ version.
func GetGtkVersion() (magor, minor, micro uint) {
	magor = gtk.GetMajorVersion()
	minor = gtk.GetMinorVersion()
	micro = gtk.GetMicroVersion()
	return
}

// GetGlibVersion return actually installed GLIB version.
func GetGlibVersion() (magor, minor, micro uint) {
	magor = glib.GetMajorVersion()
	minor = glib.GetMinorVersion()
	micro = glib.GetMicroVersion()
	return
}

// GetGdkVersion return actually installed GDK version.
func GetGdkVersion() (magor, minor, micro uint) {
	magor = gdk.GetMajorVersion()
	minor = gdk.GetMinorVersion()
	micro = gdk.GetMicroVersion()
	return
}

// ApplyStyleCSS apply custom CSS to specific widget.
func ApplyStyleCSS(widget *gtk.Widget, css string) error {
	//	provider, err := gtk.CssProviderNew()
	provider, err := gtk.CssProviderNew()
	if err != nil {
		return err
	}
	err = provider.LoadFromData(css)
	if err != nil {
		return err
	}
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	sc.AddProvider(provider, gtk.STYLE_PROVIDER_PRIORITY_USER)
	return nil
}

// AddStyleClasses apply specific CSS style classes to the widget.
func AddStyleClasses(widget *gtk.Widget, cssClasses []string) error {
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	for _, className := range cssClasses {
		sc.AddClass(className)
	}
	return nil
}

// AddStyleClass apply specific CSS style class to the widget.
func AddStyleClass(widget *gtk.Widget, cssClass string) error {
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	sc.AddClass(cssClass)
	return nil
}

// RemoveStyleClass remove specific CSS style class from the widget.
func RemoveStyleClass(widget *gtk.Widget, cssClass string) error {
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	sc.RemoveClass(cssClass)
	return nil
}

// RemoveStyleClasses remove specific CSS style classes from the widget.
func RemoveStyleClasses(widget *gtk.Widget, cssClasses []string) error {
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	for _, className := range cssClasses {
		sc.RemoveClass(className)
	}
	return nil
}

// RemoveStyleClassesAll remove all style classes from the widget.
func RemoveStyleClassesAll(widget *gtk.Widget) error {
	sc, err := widget.GetStyleContext()
	if err != nil {
		return err
	}
	list := sc.ListClasses()
	list.Foreach(func(item interface{}) {
		cssClass := item.(string)
		sc.RemoveClass(cssClass)
	})
	return nil
}

// ========================================================================================
// ************************* GTK+ UI UTILITIES SECTION END ********************************
// ========================================================================================
