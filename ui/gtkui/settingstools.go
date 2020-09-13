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
	"bytes"
	"errors"
	"strconv"

	"github.com/d2r2/gotk3/glib"
)

// ==========================================================================================
// ************************* GLIB SETTINGS UTILITIES SECTION START **************************
// ==========================================================================================
//	In real application use this code section as utilities to simplify creation
//	of GLIB/GTK+ components and widgets, including menus, dialog boxes, messages,
//	application settings and so on...

// SettingsStore simplify work with glib.Settings.
type SettingsStore struct {
	settings *glib.Settings
	schemaID string
	path     string
}

// removeExcessSlashChars normalize path and remove excess path divider in glib.Settings schema path.
func removeExcessSlashChars(path string) string {
	var buf bytes.Buffer
	lastCharIsSlash := false
	for _, ch := range path {
		if ch == '/' {
			if lastCharIsSlash {
				continue
			}
			lastCharIsSlash = true
		} else {
			lastCharIsSlash = false
		}
		buf.WriteRune(ch)
	}

	path = buf.String()

	return path
}

// NewSettingsStore create new SettingsStore object - wrapper on glib.Settings.
func NewSettingsStore(schemaID string, path string, changed func()) (*SettingsStore, error) {
	path = removeExcessSlashChars(path)
	lg.Debugf("glib.GSettings path: %s", path)
	gs, err := glib.SettingsNewWithPath(schemaID, path)
	if err != nil {
		return nil, err
	}
	_, err = gs.Connect("changed", func() {
		if changed != nil {
			changed()
		}
	})
	if err != nil {
		return nil, err
	}
	v := &SettingsStore{settings: gs, schemaID: schemaID, path: path}
	return v, nil
}

// GetChildSettingsStore generate child glib.Settings object to manipulate with nested scheme.
func (v *SettingsStore) GetChildSettingsStore(suffixSchemaID string, suffixPath string,
	changed func()) (*SettingsStore, error) {

	newSchemaID := v.schemaID + "." + suffixSchemaID
	newPath := v.path + "/" + suffixPath + "/"
	settings, err := NewSettingsStore(newSchemaID, newPath, changed)
	return settings, err
}

// GetSchema obtains glib.SettingsSchema from glib.Settings.
func (v *SettingsStore) GetSchema() (*glib.SettingsSchema, error) {
	val, err := v.settings.GetProperty("settings-schema")
	if err != nil {
		return nil, err
	}
	if schema, ok := val.(*glib.SettingsSchema); ok {
		return schema, nil
	} else {
		return nil, errors.New("GLib settings-schema property is not convertible to SettingsSchema")
	}
}

// SettingsArray is a way how to create multiple (indexed) GLib setting's group
// based on single schema. For instance, multiple backup profiles with identical
// settings inside of each profile.
type SettingsArray struct {
	store   *SettingsStore
	arrayID string
}

// NewSettingsArray creates new SettingsArray, to keep/add/delete new
// indexed glib.Settings object based on single schema.
func (v *SettingsStore) NewSettingsArray(arrayID string) *SettingsArray {
	sa := &SettingsArray{store: v, arrayID: arrayID}
	return sa
}

// DeleteNode delete specific indexed glib.Settings defined by nodeID.
func (v *SettingsArray) DeleteNode(childStore *SettingsStore, nodeID string) error {
	// Delete/reset whole child settings object.
	schema, err := childStore.GetSchema()
	if err != nil {
		return err
	}
	keys := schema.ListKeys()
	for _, key := range keys {
		childStore.settings.Reset(key)
	}

	// Delete index from the array, which identify
	// child object settings.
	original := v.store.settings.GetStrv(v.arrayID)
	var updated []string
	for _, id := range original {
		if id != nodeID {
			updated = append(updated, id)
		}
	}
	v.store.settings.SetStrv(v.arrayID, updated)
	return nil
}

// AddNode add specific indexed glib.Settings identified by returned nodeID.
func (v *SettingsArray) AddNode() (nodeID string, err error) {
	list := v.store.settings.GetStrv(v.arrayID)
	// Append index to the end of array, which reference to the list
	// of child settings based on single settings schema.
	var ni int
	if len(list) > 0 {
		ni, err = strconv.Atoi(list[len(list)-1])
		if err != nil {
			return "", err
		}
		ni++
	}
	list = append(list, strconv.Itoa(ni))
	v.store.settings.SetStrv(v.arrayID, list)
	return list[len(list)-1], nil
}

// GetArrayIDs return identifiers of glib.Settings with common schema,
// which can be accessed using id from the list.
func (v *SettingsArray) GetArrayIDs() []string {
	list := v.store.settings.GetStrv(v.arrayID)
	return list
}

// Binding cache link between Key string identifier and GLIB object property.
// Code partially taken from https://github.com/gnunn1/tilix project.
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
	settings *SettingsStore
}

// NewBindingHelper creates new BindingHelper object.
func (v *SettingsStore) NewBindingHelper() *BindingHelper {
	bh := &BindingHelper{settings: v}
	return bh
}

// SetSettings will replace underlying GLIB Settings object to unbind
// previously set bindings and re-bind to the new settings automatically.
func (v *BindingHelper) SetSettings(value *SettingsStore) {
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
			v.settings.settings.Bind(b.Key, b.Object, b.Property, b.Flags)
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
		v.settings.settings.Bind(key, object, property, flags)
	}
}

// Unbind all added binds from settings object.
func (v *BindingHelper) Unbind() {
	for _, b := range v.bindings {
		v.settings.settings.Unbind(b.Object, b.Property)
	}
}

// Clear unbind all bindings and clears list of bindings.
func (v *BindingHelper) Clear() {
	v.Unbind()
	v.bindings = nil
}

// ==========================================================================================
// ************************* GLIB SETTINGS UTILITIES SECTION END ****************************
// ==========================================================================================
