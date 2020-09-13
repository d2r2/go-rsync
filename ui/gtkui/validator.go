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
	"context"
	"fmt"
	"sync"
)

// ValidatorData is an array of arbitrary data
// used to pass to the validation process.
type ValidatorData struct {
	Items []interface{}
}

// ValidatorInit initialize validation process with next attributes:
// - Synchronous call.
// - Should take a limited time to execute.
// - Allowed to updated GTK+ widgets.
type ValidatorInit func(data *ValidatorData, group []*ValidatorData) error

// ValidatorRun run validation process with next characteristics:
// - Asynchronous call.
// - Can take long time to run.
// - GTK+ widgets should not be updated here (read only allowed).
type ValidatorRun func(groupLock *sync.Mutex, ctx context.Context, data *ValidatorData,
	group []*ValidatorData) ([]interface{}, error)

// ValidatorEnd finalize validation process with next characteristics:
// - Asynchronous call.
// - Should take a limited time to execute.
// - GTK+ widgets might be updated here, if you wrap calls to glib.IdleAdd method.
type ValidatorEnd func(groupLock *sync.Mutex, data *ValidatorData, results []interface{}) error

// ValidatorEntry stores validation data all together,
// including 3-step validation process (initialize, run, finalize).
type ValidatorEntry struct {
	group string
	index string
	init  ValidatorInit
	run   ValidatorRun
	end   ValidatorEnd
	Data  *ValidatorData
}

// GroupMap gives thread-safe indexed dictionary,
// which allow manipulations in asynchronous mode.
// Store validator groups uniquely identified
// by group and index identifiers.
type GroupMap struct {
	sync.RWMutex
	m    map[string]*ContextPack // keep Context object here indexed by group+index identifiers
	lock map[string]*sync.Mutex  // keep lock object here indexed by group identifier
}

func GroupMapNew() *GroupMap {
	v := &GroupMap{m: make(map[string]*ContextPack),
		lock: make(map[string]*sync.Mutex)}
	return v
}

// getFullIndex return complex index from concatenation
// of group and index identifiers.
func getFullIndex(group, index string) string {
	return fmt.Sprintf("%s_%s", group, index)
}

// Add create new group identified by group+index identifiers.
// if not exists create, either return it.
func (v *GroupMap) Add(group, index string, ctxPack *ContextPack) {
	v.Lock()
	defer v.Unlock()

	v.m[getFullIndex(group, index)] = ctxPack
}

// Get return group context identified by group+index identifiers if exists.
func (v *GroupMap) Get(group, index string) (*ContextPack, bool) {
	v.RLock()
	defer v.RUnlock()

	ctxPack, ok := v.m[getFullIndex(group, index)]
	return ctxPack, ok
}

// GetLock return lock object identified by group identifier.
func (v *GroupMap) GetLock(group string) *sync.Mutex {
	v.Lock()
	defer v.Unlock()

	if _, ok := v.lock[group]; !ok {
		v.lock[group] = &sync.Mutex{}
	}
	return v.lock[group]
}

// Remove delete group object identified by group+index identifiers.
func (v *GroupMap) Remove(group, index string) {
	v.Lock()
	defer v.Unlock()

	delete(v.m, getFullIndex(group, index))
}

// UIValidator simplify GTK UI validation process
// mixing synchronized and asynchronous calls,
// which all together does not freeze GTK UI,
// providing beautiful GTK UI response.
// UIValidator is a thread-safe (except cases
// when you need update GtkWidget components -
// you must be careful in such circumstances).
type UIValidator struct {
	sync.RWMutex
	entries         map[int]*ValidatorEntry
	sorted          []int
	key             int
	parent          context.Context
	runningContexts RunningContexts
	groupRunning    *GroupMap
}

func UIValidatorNew(parent context.Context) *UIValidator {
	entries := make(map[int]*ValidatorEntry)
	groupRunning := GroupMapNew()
	v := &UIValidator{entries: entries, parent: parent, groupRunning: groupRunning}
	return v
}

// AddEntry creates new validating process with specific groupID and subGroupID identifiers.
// Provide additionally 3 callback methods: to initialize, to run and to finalize validation.
func (v *UIValidator) AddEntry(group, index string, init ValidatorInit, run ValidatorRun,
	end ValidatorEnd, data ...interface{}) int {

	v.Lock()
	defer v.Unlock()

	vEntry := &ValidatorEntry{group: group, index: index,
		init: init, run: run, end: end, Data: &ValidatorData{data}}
	key := v.key
	v.entries[key] = vEntry
	v.sorted = append(v.sorted, key)
	v.key++
	return key
}

// RemoveEntry remove validating process via index key.
func (v *UIValidator) RemoveEntry(key int) {
	v.Lock()
	defer v.Unlock()

	if val, ok := v.entries[key]; ok {
		v.cancelValidatesIfRunning(val.group, val.index)

		lg.Debugf("Delete item %q with index %v", getFullIndex(val.group, val.index), key)
		delete(v.entries, key)
	}
	for ind, k := range v.sorted {
		if key == k {
			v.sorted = append(v.sorted[:ind], v.sorted[ind+1:]...)
			break
		}
	}
}

// GetCount return number of validating processes.
func (v *UIValidator) GetCount() int {
	v.RLock()
	defer v.RUnlock()

	return len(v.entries)
}

// getGroupEntries return list of ValidatorEntry objects,
// identified by group + index identifiers.
func (v *UIValidator) getGroupEntries(group, index string) []*ValidatorEntry {
	var list []*ValidatorEntry
	for _, key := range v.sorted {
		if v.entries[key].group == group && v.entries[key].index == index {
			list = append(list, v.entries[key])
		}
	}
	return list
}

// getGroupData return list of ValidatorData objects,
// identified by group + index identifiers.
func (v *UIValidator) getGroupData(group, index string) []*ValidatorData {
	var list []*ValidatorData
	for _, key := range v.sorted {
		if v.entries[key].group == group && v.entries[key].index == index {
			list = append(list, v.entries[key].Data)
		}
	}
	return list
}

// resultsOrError used to get results from
// validator asynchronous context execution.
type resultsOrError struct {
	Entry   *ValidatorEntry
	Results []interface{}
	Error   error
}

// callInit run 1st validation step synchronously.
func (v *UIValidator) callInit(entry *ValidatorEntry, dataList []*ValidatorData) error {

	return entry.init(entry.Data, dataList)
}

// callRun run 2nd validation step asynchronously.
func (v *UIValidator) callRun(groupLock *sync.Mutex, ctx context.Context,
	entry *ValidatorEntry, dataList []*ValidatorData) ([]interface{}, error) {

	return entry.run(groupLock, ctx, entry.Data, dataList)
}

// callEnd run 3rd validation step asynchronously, but can be
// synchronized with GTK+ context via glib.IdleAdd function.
func (v *UIValidator) callEnd(groupLock *sync.Mutex, r resultsOrError) {
	err := r.Entry.end(groupLock, r.Entry.Data, r.Results)
	if err != nil {
		lg.Fatal(err)
	}
}

// runAsync run 2nd and 3rd validation process steps.
func (v *UIValidator) runAsync(group, index string, entryList []*ValidatorEntry,
	dataList []*ValidatorData) {

	resultCh := make(chan resultsOrError)

	ctxPack := ForkContext(v.parent)
	v.runningContexts.AddContext(ctxPack)
	v.groupRunning.Add(group, index, ctxPack)
	groupLock := v.groupRunning.GetLock(group)
	var wait sync.WaitGroup
	wait.Add(1)

	// Run 3rd validation step in advance, to wait for results
	// from 2nd validation steps.
	go func() {
		defer wait.Done()

		terminated := false
		for {
			select {
			case r, ok := <-resultCh:
				if ok {
					err := r.Error
					if err == nil {
						lg.Debugf("Read Validator results %v", r.Results)
						lg.Debugf("Call Validator End")
						v.callEnd(groupLock, r)
					} else {
						lg.Fatal(err)
					}
				} else {
					lg.Debugf("Complete group %q validation 2", getFullIndex(group, index))
					terminated = true
				}
			case <-ctxPack.Context.Done():
				terminated = true
			}
			if terminated {
				break
			}
		}
	}()

	// Run 2nd validation step.
	go func() {
		terminated := false
		for _, item := range entryList {
			r := resultsOrError{Entry: item}
			results, err := v.callRun(groupLock, ctxPack.Context, item, dataList)
			if err != nil {
				r.Error = err
			} else {
				r.Results = results
			}
			select {
			case resultCh <- r:
			case <-ctxPack.Context.Done():
				terminated = true
			}
			if terminated {
				break
			}
		}
		lg.Debugf("Complete group %q validation 1", getFullIndex(group, index))
		close(resultCh)
		// Wait for completion of 3rd validation step (finalizer), before exit.
		wait.Wait()
		v.runningContexts.RemoveContext(ctxPack.Context)
		v.groupRunning.Remove(group, index)
	}()
}

// cancelValidatesIfRunning checks if validation process
// in progress and thus cancel it.
func (v *UIValidator) cancelValidatesIfRunning(group, index string) {
	if ctxPack, ok := v.groupRunning.Get(group, index); ok {
		lg.Debugf("Cancel group %q validation", getFullIndex(group, index))
		ctxPack.Cancel()
	}
}

// Validate is a main entry point to start validation process
// for specific group.
// Validate process trigger next strictly sequential steps:
// 1) Call "init validation" custom function in synchronous context.
// So, it's safe to update GTK+ widgets here.
// 2) Call "run validation" custom function in asynchronous context.
// You should never update GTK+ widgets here (you can read widgets), but might run
// long-term operations here (for instance run some external application).
// 3) Call "finalize validation" custom function in asynchronous context.
// You can update GTK+ widgets here, if you wrap code there with glib.IdleAdd()
// function from GOTK+ library to synchronize with GTK+ context.
func (v *UIValidator) Validate(group, index string) error {
	v.Lock()
	defer v.Unlock()

	entryList := v.getGroupEntries(group, index)
	dataList := v.getGroupData(group, index)
	if len(entryList) > 0 {
		// If found that previous validation processes in progress
		// with specific id still, then cancel it.
		v.cancelValidatesIfRunning(group, index)

		for _, item := range entryList {
			// 1st step of validation process
			err := v.callInit(item, dataList)
			if err != nil {
				return err
			}
			// 2nd and 3rd steps of validation process
			v.runAsync(group, index, entryList, dataList)
		}
	}
	return nil
}

// CancelValidates cancel processes identified
// by group + index identifiers, if running.
func (v *UIValidator) CancelValidates(group, index string) {
	v.Lock()
	defer v.Unlock()

	v.cancelValidatesIfRunning(group, index)
}

// CancelAll cancel all pending processes if running.
func (v *UIValidator) CancelAll() {
	v.runningContexts.CancelAll()
}
