package gtkui

import (
	"context"
	"sync"
)

// ContextPack keeps context with it cancel function.
type ContextPack struct {
	Context context.Context
	Cancel  func()
}

// ForkContext create child context from parent.
func ForkContext(parent context.Context) *ContextPack {
	child, cancel := context.WithCancel(parent)
	v := &ContextPack{Context: child, Cancel: cancel}
	return v
}

// RunningContexts keeps all currently started services,
// preliminary added to the list, which we would like
// to control, tracking and managing their states.
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

// Start forks new context for some thread.
func (v *BackupSessionStatus) Start() *ContextPack {
	pack := ForkContext(v.parent)
	v.running.AddContext(pack)
	return pack
}

// IsRunning checks if any threads are alive.
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
