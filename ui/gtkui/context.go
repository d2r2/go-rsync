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
	"context"
	"sync"
)

// ContextPack keeps cancellable context with its cancel function.
type ContextPack struct {
	Context context.Context
	Cancel  func()
}

// ForkContext create child context from the parent.
func ForkContext(parent context.Context) *ContextPack {
	child, cancel := context.WithCancel(parent)
	v := &ContextPack{Context: child, Cancel: cancel}
	return v
}

// RunningContexts keeps all contexts of currently started services,
// preliminary added to the list, which we would like to control,
// tracking and managing their states.
// All methods of RunningContexts type are thread-safe.
type RunningContexts struct {
	sync.RWMutex
	running []*ContextPack
}

// AddContext add new service to track.
func (v *RunningContexts) AddContext(pack *ContextPack) {
	v.Lock()
	defer v.Unlock()
	v.running = append(v.running, pack)
}

func (v *RunningContexts) findIndex(ctx context.Context) int {
	index := -1
	for i, item := range v.running {
		if item.Context == ctx {
			index = i
			break
		}
	}
	return index
}

// RemoveContext remove service from the list.
func (v *RunningContexts) RemoveContext(ctx context.Context) {
	v.Lock()
	defer v.Unlock()
	index := v.findIndex(ctx)
	if index != -1 {
		v.running = append(v.running[:index], v.running[index+1:]...)
	}
}

// CancelContext cancel service from the list.
func (v *RunningContexts) CancelContext(ctx context.Context) {
	v.Lock()
	defer v.Unlock()
	index := v.findIndex(ctx)
	if index != -1 {
		v.running[index].Cancel()
		v.running = append(v.running[:index], v.running[:index+1]...)
	}
}

// CancelAll cancel all services in the list.
func (v *RunningContexts) CancelAll() {
	v.Lock()
	defer v.Unlock()
	for _, item := range v.running {
		item.Cancel()
	}
	v.running = []*ContextPack{}
}

// FindContext finds service by context object.
func (v *RunningContexts) FindContext(ctx context.Context) *ContextPack {
	v.RLock()
	defer v.RUnlock()
	index := v.findIndex(ctx)
	if index != -1 {
		return v.running[index]
	}
	return nil
}

// GetCount returns number of services in the list to control.
func (v *RunningContexts) GetCount() int {
	v.RLock()
	defer v.RUnlock()
	return len(v.running)
}

// BackupSessionStatus keeps contexts - live multi-thread processes,
// which life cycle should be controlled.
type BackupSessionStatus struct {
	parent  context.Context
	running RunningContexts
}

func NewBackupSessionStatus(parent context.Context) *BackupSessionStatus {
	v := &BackupSessionStatus{parent: parent}
	return v
}

// Start forks new context for parent thread.
func (v *BackupSessionStatus) Start() *ContextPack {
	pack := ForkContext(v.parent)
	v.running.AddContext(pack)
	return pack
}

// IsRunning checks if any children threads are alive.
func (v *BackupSessionStatus) IsRunning() bool {
	return v.running.GetCount() > 0
}

// Stop terminates all live thread's contexts.
func (v *BackupSessionStatus) Stop() {
	v.running.CancelAll()
}

// Done removes context from the pool of controlled threads.
func (v *BackupSessionStatus) Done(ctx context.Context) {
	v.running.RemoveContext(ctx)
}
