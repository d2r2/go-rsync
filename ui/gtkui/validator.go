package gtkui

import (
	"context"
	"sync"

	"github.com/d2r2/gotk3/glib"
)

// ValidatorData is an array of arbitrary data
// used to pass to the validation process.
type ValidatorData struct {
	Items []interface{}
}

// ValidatorInit init validation process with next attributes:
// - Synchronous call.
// - Should take a limited time.
// - Allowed to updated GTK widgets.
type ValidatorInit func(data *ValidatorData, group []*ValidatorData) error

// ValidatorRun run validation process with next characteristics:
// - Asynchronous call.
// - Can take long time to run.
// - GTK widgets should not be updated here (read only allowed).
type ValidatorRun func(ctx context.Context, data *ValidatorData,
	group []*ValidatorData) ([]interface{}, error)

// ValidatorEnd finalize validation process with next characteristics:
// - Synchronous call.
// - Should take a limited time.
// - Allowed to updated GTK widgets.
type ValidatorEnd func(data *ValidatorData, results []interface{}) error

// ValidatorEntry stores validation data all together,
// including 3-step validation process (init, run, finallize).
type ValidatorEntry struct {
	GroupName string
	init      ValidatorInit
	run       ValidatorRun
	end       ValidatorEnd
	Data      *ValidatorData
}

// GroupMap gives thread-safe dictionary,
// which allow manipulations in asynchronous mode.
type GroupMap struct {
	sync.RWMutex
	m map[string]*ContextPack
}

func GroupMapNew() *GroupMap {
	v := &GroupMap{m: make(map[string]*ContextPack)}
	return v
}

func (v *GroupMap) Add(groupName string, ctxPack *ContextPack) {
	v.Lock()
	defer v.Unlock()

	v.m[groupName] = ctxPack
}

func (v *GroupMap) Remove(groupName string) {
	v.Lock()
	defer v.Unlock()

	delete(v.m, groupName)
}

func (v *GroupMap) Get(groupName string) (*ContextPack, bool) {
	v.RLock()
	defer v.RUnlock()

	ctxPack, ok := v.m[groupName]
	return ctxPack, ok
}

// UIValidator simplify GTK UI validation process
// mixing synchronized and asynchroniouse calls,
// which all together does not freeze GTK UI,
// providing beautifull GTK UI responce.
// UIValidator is a thread-safe (except cases
// when you need update GtkWidget components -
// you must be careful in such circumstances).
type UIValidator struct {
	sync.Mutex
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

func (v *UIValidator) AddEntry(groupName string, init ValidatorInit, run ValidatorRun, end ValidatorEnd,
	data ...interface{}) int {

	v.Lock()
	defer v.Unlock()

	vEntry := &ValidatorEntry{GroupName: groupName,
		init: init, run: run, end: end, Data: &ValidatorData{data}}
	key := v.key
	v.entries[key] = vEntry
	v.sorted = append(v.sorted, key)
	v.key++
	return key
}

func (v *UIValidator) RemoveEntry(key int) {
	v.Lock()
	defer v.Unlock()

	if val, ok := v.entries[key]; ok {
		v.cancelValidateIfRunning(val.GroupName)

		lg.Debugf("Delete group %q with index %v", val.GroupName, key)
		delete(v.entries, key)
	}
	for ind, k := range v.sorted {
		if key == k {
			v.sorted = append(v.sorted[:ind], v.sorted[ind+1:]...)
			break
		}
	}
}

func (v *UIValidator) GetCount() int {
	v.Lock()
	defer v.Unlock()

	return len(v.entries)
}

func (v *UIValidator) getGroupEntries(groupName string) []*ValidatorEntry {
	var list []*ValidatorEntry
	for _, key := range v.sorted {
		if v.entries[key].GroupName == groupName {
			list = append(list, v.entries[key])
		}
	}
	return list
}

func (v *UIValidator) getGroupData(groupName string) []*ValidatorData {
	var list []*ValidatorData
	for _, key := range v.sorted {
		if v.entries[key].GroupName == groupName {
			list = append(list, v.entries[key].Data)
		}
	}
	return list
}

// resultsOrError used to get results from
// asynchonous context.
type resultsOrError struct {
	Entry   *ValidatorEntry
	Results []interface{}
	Error   error
}

// callEnd run 3rd validation step asynchronously,
// but synchronize call with Gtk+ context via
// glib.IdleAdd function.
func (v *UIValidator) callEnd(r resultsOrError) {
	_, err := glib.IdleAdd(func() {
		err := r.Entry.end(r.Entry.Data, r.Results)
		if err != nil {
			lg.Fatal(err)
		}
	})
	if err != nil {
		lg.Fatal(err)
	}
}

// callRun run 2nd validation step asynchronously.
func (v *UIValidator) callRun(ctx context.Context, entry *ValidatorEntry,
	dataList []*ValidatorData) ([]interface{}, error) {

	return entry.run(ctx, entry.Data, dataList)
}

// callRun run 1st validation step synchronously.
func (v *UIValidator) callInit(entry *ValidatorEntry, dataList []*ValidatorData) error {

	return entry.init(entry.Data, dataList)
}

// runAsync run 2nd and 3rd validation process steps.
func (v *UIValidator) runAsync(groupName string, entryList []*ValidatorEntry,
	dataList []*ValidatorData) {

	waitCh := make(chan resultsOrError)
	done := make(chan struct{})

	ctxPack := ForkContext(v.parent)
	v.runningContexts.AddContext(ctxPack)
	v.groupRunning.Add(groupName, ctxPack)

	// run here 3rd validation step, with
	go func() {
		for {
			select {
			case r := <-waitCh:
				terminated := false
				select {
				case <-ctxPack.Context.Done():
					terminated = true
				default:
				}
				if !terminated {
					lg.Debugf("Read Validator results: %v", r)
					err := r.Error
					if r.Error == nil {
						lg.Debugf("Call Validator End")
						v.callEnd(r)
					}
					if err != nil {
						lg.Fatal(err)
					}
				}
			case <-done:
				lg.Debugf("Complete group %q validation 2", groupName)
				close(waitCh)
				return
			}
		}
	}()

	// run here 2nd validation step
	go func() {
		for _, item := range entryList {
			r := resultsOrError{Entry: item}
			results, err := v.callRun(ctxPack.Context, item, dataList)
			if err != nil {
				r.Error = err
			} else {
				r.Results = results
			}
			select {
			case waitCh <- r:
				lg.Debugf("Send Validator results: %v", r)
			case <-ctxPack.Context.Done():
				break
			}
		}
		lg.Debugf("Complete group %q validation 1", groupName)
		close(done)
		v.runningContexts.RemoveContext(ctxPack.Context)
		v.groupRunning.Remove(groupName)
	}()
}

// cancelValidateIfRunning checks if validation process
// in progress and cancel it.
func (v *UIValidator) cancelValidateIfRunning(groupName string) {
	if ctxPack, ok := v.groupRunning.Get(groupName); ok {
		lg.Debugf("Cancel group %q validation", groupName)
		ctxPack.Cancel()
	}
}

// Validate is main entry point to start validation process
// for specific group.
// Validate process trigger next strictly sequential steps:
// 1) Call "init validatation" custom function in synchronous context.
// So, it's safe to update Gtk+ widgets here.
// 2) Call "run validation" custom function in asynchronous context.
// You should never update Gtk+ widgets here (you can read widgets), but might run
// long-term operations here (for instance run some external application).
// 3) Call "finalize validation" custom function in asynchronous context.
// Still you can update Gtk+ widgets here again, because this call synchonized
// with Gtk+ context via glib.IdleAdd() function from GOTK+ library.
func (v *UIValidator) Validate(groupName string) error {
	v.Lock()
	defer v.Unlock()

	entryList := v.getGroupEntries(groupName)
	dataList := v.getGroupData(groupName)
	if len(entryList) > 0 {
		v.cancelValidateIfRunning(groupName)

		for _, item := range entryList {
			// 1st step of validation process
			err := v.callInit(item, dataList)
			if err != nil {
				return err
			}
			// 2nd and 3rd steps of validation process
			v.runAsync(groupName, entryList, dataList)
		}
	}
	return nil
}

func (v *UIValidator) CancelValidate(groupName string) {
	v.Lock()
	defer v.Unlock()

	v.cancelValidateIfRunning(groupName)
}

func (v *UIValidator) CancelAll() {
	v.runningContexts.CancelAll()
}
