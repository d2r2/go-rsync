package gtkui

import (
	"errors"
	"strconv"

	"github.com/d2r2/gotk3/gdk"
	"github.com/d2r2/gotk3/glib"
	"github.com/d2r2/gotk3/gtk"
	"github.com/d2r2/gotk3/pango"
	"github.com/davecgh/go-spew/spew"
)

// ========================================================================================
// ************************* GTK GUI UTILITIES SECTION START ******************************
// ========================================================================================
//	In real application copy this section to separate file as utilities functions to simplify
//	creation of GLIB/GTK+ components and widgets, including menus, dialog boxes, messages,
//	application settings and so on...

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

func getPixbufFromBytes(bytes []byte) (*gdk.Pixbuf, error) {
	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	ms, err := glib.MemoryInputStreamFromBytesNew(b2)
	if err != nil {
		return nil, err
	}
	pb, err := gdk.PixbufNewFromStream(&ms.InputStream, nil)
	if err != nil {
		return nil, err
	}
	return pb, nil
}

func getPixbufAnimationFromBytes(bytes []byte) (*gdk.PixbufAnimation, error) {
	b2, err := glib.BytesNew(bytes)
	if err != nil {
		return nil, err
	}
	ms, err := glib.MemoryInputStreamFromBytesNew(b2)
	if err != nil {
		return nil, err
	}
	pba, err := gdk.PixbufAnimationNewFromStream(&ms.InputStream, nil)
	if err != nil {
		return nil, err
	}
	return pba, nil
}

// SetupMenuButtonWithThemedImage construct MenuButton widget with image
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

// AppendSectionAsHorzButtons used for Popover widget menu
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

// DialogButton simplify Dialog window initialization.
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

// PrintDialogResponse print and debug dialog responce.
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

func NewDialogParagraph(text string) *DialogParagraph {
	v := &DialogParagraph{Text: text, HorizAlign: gtk.Align(-1), Justify: gtk.Justification(-1),
		Ellipsize: pango.EllipsizeMode(-1), MaxWidthChars: -1}
	return v
}

func NewMarkupDialogParagraph(text string) *DialogParagraph {
	v := &DialogParagraph{Text: text, Markup: true, HorizAlign: gtk.Align(-1), Justify: gtk.Justification(-1),
		Ellipsize: pango.EllipsizeMode(-1), MaxWidthChars: -1}
	return v
}

func (v *DialogParagraph) SetHorizAlign(align gtk.Align) *DialogParagraph {
	v.HorizAlign = align
	return v
}

func (v *DialogParagraph) SetJustify(justify gtk.Justification) *DialogParagraph {
	v.Justify = justify
	return v
}

func (v *DialogParagraph) SetEllipsize(ellipsize pango.EllipsizeMode) *DialogParagraph {
	v.Ellipsize = ellipsize
	return v
}

func (v *DialogParagraph) SetMaxWidthChars(maxWidthChars int) *DialogParagraph {
	v.MaxWidthChars = maxWidthChars
	return v
}

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

func TextToDialogParagraphs(lines []string) []*DialogParagraph {
	var msgs []*DialogParagraph
	for _, line := range lines {
		msgs = append(msgs, NewDialogParagraph(line))
	}
	return msgs
}

func TextToMarkupDialogParagraphs(lines []string) []*DialogParagraph {
	var msgs []*DialogParagraph
	for _, line := range lines {
		msgs = append(msgs, NewMarkupDialogParagraph(line))
	}
	return msgs
}

// SetupMessageDialog construct MessageDialog widget with customized settings.
func SetupMessageDialog(parent *gtk.Window, markupTitle string, secondaryMarkupTitle string,
	paragraphs []*DialogParagraph, addButtons []DialogButton,
	addExtraControls func(area *gtk.Box) error) (*gtk.MessageDialog, error) {

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

	return dlg, nil
}

// RunMessageDialog construct and run MessageDialog widget with customized settings.
func RunMessageDialog(parent *gtk.Window, markupTitle string, secondaryMarkupTitle string,
	paragraphs []*DialogParagraph, ignoreCloseBox bool, addButtons []DialogButton,
	addExtraControls func(area *gtk.Box) error) (gtk.ResponseType, error) {

	dlg, err := SetupMessageDialog(parent, markupTitle, secondaryMarkupTitle,
		paragraphs, addButtons, addExtraControls)
	if err != nil {
		return 0, err
	}
	defer dlg.Destroy()

	dlg.ShowAll()
	res := dlg.Run()
	for gtk.ResponseType(res) == gtk.RESPONSE_NONE || gtk.ResponseType(res) == gtk.RESPONSE_DELETE_EVENT && ignoreCloseBox {
		res = dlg.Run()
	}
	return gtk.ResponseType(res), nil
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

func ErrorMessage(parent *gtk.Window, titleMarkup string, text []*DialogParagraph) error {
	buttons := []DialogButton{
		{"_OK", gtk.RESPONSE_OK, false, nil},
	}
	_, err := RunMessageDialog(parent, titleMarkup, "", text, false, buttons, nil)
	if err != nil {
		return err
	}
	return nil
}

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

// GetSchema obtains glib.SettingsSchema from glib.Settings.
func GetSchema(v *glib.Settings) (*glib.SettingsSchema, error) {
	val, err := v.GetProperty("settings-schema")
	if err != nil {
		return nil, err
	}
	if schema, ok := val.(*glib.SettingsSchema); ok {
		return schema, nil
	} else {
		return nil, errors.New("GLib settings-schema property is not convertible to SettingsSchema")
	}
}

// FixProgressBarCSS eliminate issue with default GtkProgressBar control formating.
func applyStyleCSS(widget *gtk.Widget, css string) error {
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
	//sc.AddClass("osd")
	return nil
}

// Binding cache link between Key string identifier and GLIB object property.
// Code taken from https://github.com/gnunn1/tilix project.
type Binding struct {
	Key      string
	Object   glib.IObject
	Property string
	Flags    glib.SettingsBindFlags
}

// BindingHelper is a bookkeeping class that keeps track of objects which are
// binded to a GSettings object so they can be unbinded later. it
// also supports the concept of deferred bindings where a binding
// can be added but is not actually attached to a Settings object
// until one is set.
type BindingHelper struct {
	bindings []Binding
	settings *glib.Settings
}

// BindingHelperNew creates new BindingHelper object.
func BindingHelperNew(settings *glib.Settings) *BindingHelper {
	bh := &BindingHelper{settings: settings}
	return bh
}

// SetSettings will replace underlying GLIB Settings object to unbind
// previously set bindings and re-bind to the new settings automatically.
func (v *BindingHelper) SetSettings(value *glib.Settings) {
	if value != v.settings {
		if v.settings != nil {
			v.Unbind()
		}
		v.settings = value
		if v.settings != nil {
			v.bindAll()
		}
	}
}

func (v *BindingHelper) bindAll() {
	if v.settings != nil {
		for _, b := range v.bindings {
			v.settings.Bind(b.Key, b.Object, b.Property, b.Flags)
		}
	}
}

// addBind add a binding to the list
func (v *BindingHelper) addBind(key string, object glib.IObject, property string, flags glib.SettingsBindFlags) {
	v.bindings = append(v.bindings, Binding{key, object, property, flags})
}

// Bind add a binding to list and binds to Settings if it is set.
func (v *BindingHelper) Bind(key string, object glib.IObject, property string, flags glib.SettingsBindFlags) {
	v.addBind(key, object, property, flags)
	if v.settings != nil {
		v.settings.Bind(key, object, property, flags)
	}
}

// Unbind all added binds from settings object.
func (v *BindingHelper) Unbind() {
	for _, b := range v.bindings {
		v.settings.Unbind(b.Object, b.Property)
	}
}

// Clear unbind all bindings and clears list of bindings.
func (v *BindingHelper) Clear() {
	v.Unbind()
	v.bindings = nil
}

// SettingsArray is a way how to create multiple (indexed) GLib setting's group.
// For instance, multiple backup profiles with identical
// settings inside of each profile. Either each backup profile may
// contain more than one data source for backup.
type SettingsArray struct {
	settings *glib.Settings
	arrayID  string
}

func NewSettingsArray(settings *glib.Settings, arrayID string) *SettingsArray {
	v := &SettingsArray{settings: settings, arrayID: arrayID}
	return v
}

func (v *SettingsArray) DeleteNode(childSettings *glib.Settings, nodeID string) error {
	schema, err := GetSchema(childSettings)
	if err != nil {
		return err
	}
	keys := schema.ListKeys()
	lg.Debug(spew.Sprintf("%+v", keys))
	for _, key := range keys {
		childSettings.Reset(key)
	}

	sources := v.settings.GetStrv(v.arrayID)
	var newSources []string
	for _, id := range sources {
		if id != nodeID {
			newSources = append(newSources, id)
		}
	}
	v.settings.SetStrv(v.arrayID, newSources)
	return nil
}

func (v *SettingsArray) AddNode() (nodeID string, err error) {
	sources := v.settings.GetStrv(v.arrayID)
	var ni int
	if len(sources) > 0 {
		ni, err = strconv.Atoi(sources[len(sources)-1])
		if err != nil {
			return "", err
		}
		ni++
	}
	//lg.Println(spew.Sprintf("New node id: %+v", ni))
	sources = append(sources, strconv.Itoa(ni))
	v.settings.SetStrv(v.arrayID, sources)
	return sources[len(sources)-1], nil
}

func (v *SettingsArray) GetArrayIDs() []string {
	sources := v.settings.GetStrv(v.arrayID)
	return sources
}

// ========================================================================================
// ************************* GTK GUI UTILITIES SECTION END ********************************
// ========================================================================================
