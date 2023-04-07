package process

import (
	"errors"
	"sync"
)

type Manager struct {
	wg            sync.WaitGroup
	workers       []Worker
	workersLock   sync.RWMutex
	isRunning     bool
	isRunningLock sync.RWMutex
}

func NewManagerWithWorkers(workers []Worker) *Manager {
	manager := &Manager{
		workers: workers,
	}

	return manager
}

func NewManager() *Manager {
	return NewManagerWithWorkers([]Worker{})
}

func (m *Manager) AddWorker(worker Worker) {
	m.workersLock.Lock()
	defer m.workersLock.Unlock()

	m.workers = append(m.workers, worker)

	if m.IsRunning() {
		m.wg.Add(1)
		go m.startWorker(worker)
	}
}

func (m *Manager) StartAll() {
	m.isRunningLock.Lock()
	defer m.isRunningLock.Unlock()

	if m.isRunning {
		log.Error("manager is already run", errors.New("manager is already run"), nil)
		return
	}

	for _, w := range m.workers {
		m.wg.Add(1)
		go m.startWorker(w)
	}

	m.isRunning = true
}

func (m *Manager) startWorker(w Worker) {
	defer m.StopAll()

	err := w.Start()
	if err != nil {
		log.Error("the channel raised an error", err, nil)
	}
	m.wg.Done()
}

func (m *Manager) stop() bool {
	m.isRunningLock.Lock()
	defer m.isRunningLock.Unlock()

	if !m.isRunning {
		return false
	}

	m.isRunning = false

	return true
}

func (m *Manager) IsRunning() bool {
	m.isRunningLock.RLock()
	defer m.isRunningLock.RUnlock()

	return m.isRunning
}

func (m *Manager) StopAll() {
	if !m.stop() {
		return
	}

	for _, w := range m.workers {
		err := w.Stop()
		if err != nil {
			log.Error("the channel raised an error", err, nil)
		}
	}
}

func (m *Manager) AwaitAll() {
	m.wg.Wait()
	log.Info("all background processes were stopped", nil)
}
